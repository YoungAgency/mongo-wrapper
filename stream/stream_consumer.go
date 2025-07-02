package stream

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Config struct {
	Database     string
	Collection   string
	Encoder      EventEncoder
	TokenManager OffsetManager
	StreamAgg    bson.D
}

type Consumer[T any, K any] struct {
	client            *mongo.Client
	encoder           EventEncoder
	tokenManager      OffsetManager
	database          string
	collection        string
	streamAggregation bson.D
}

type HandlerFn[T any, K any] func(ctx context.Context, event StreamEvent[T, K]) error

func NewStreamConsumer[T any, K any](client *mongo.Client, conf *Config) *Consumer[T, K] {
	var encoder EventEncoder
	var tokenManger OffsetManager
	if conf != nil {
		if conf.TokenManager != nil {
			tokenManger = conf.TokenManager
		} else {
			tokenManger = &defaultOffsetManager{}
		}
		if conf.Encoder != nil {
			encoder = conf.Encoder
		}
	}
	return &Consumer[T, K]{
		encoder:           encoder,
		tokenManager:      tokenManger,
		client:            client,
		database:          conf.Database,
		collection:        conf.Collection,
		streamAggregation: conf.StreamAgg,
	}
}

func (c *Consumer[T, K]) ConsumeHandler(ctx context.Context, streamOptions *options.ChangeStreamOptionsBuilder, handler HandlerFn[T, K]) error {
	stream, err := c.getStream(ctx, streamOptions)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)

	doc := StreamEvent[T, K]{}
	for stream.Next(ctx) {
		if err := stream.Decode(&doc); err != nil {
			return err
		}
		for {
			if err := handler(ctx, doc); err != nil {
				<-time.After(1 * time.Second)
				continue
			}
			err = c.tokenManager.SetOffset(ctx, StreamOffset{
				ResumeToken: doc.ID.Data,
				Timestamp:   doc.ClusterTime,
			})
			if err != nil {
				<-time.After(1 * time.Second)
				continue
			}
			break
		}
	}
	return stream.Err()
}

func (c *Consumer[T, K]) getStream(ctx context.Context, streamOptions *options.ChangeStreamOptionsBuilder) (*mongo.ChangeStream, error) {
	resumeToken, err := c.tokenManager.GetOffset(ctx)
	if err != nil {
		return nil, err
	}
	if resumeToken != nil && resumeToken.ResumeToken != "" {
		streamOptions.SetStartAfter(bson.M{
			"_data": resumeToken.ResumeToken,
		})
	} else if resumeToken != nil && !resumeToken.Timestamp.IsZero() {
		// if timestamp is out of range for oplog it will be ignored
		dt := &bson.Timestamp{T: uint32(resumeToken.Timestamp.UTC().Unix()), I: 0}
		streamOptions.SetStartAtOperationTime(dt)
	}
	stream, err := c.client.Database(c.database).
		Collection(c.collection).
		Watch(ctx, c.streamAggregation, streamOptions)
	if err != nil {
		if mongoErr, ok := err.(mongo.CommandError); ok {
			if mongoErr.Code == 286 || mongoErr.Code == 280 {
				// Resume of change stream was not possible, reset offset
				resumeToken.ResumeToken = ""
				resumeToken.Timestamp = time.Time{}
				err = c.tokenManager.SetOffset(ctx, *resumeToken)
				if err != nil {
					return nil, err
				}
				streamOptions.SetStartAfter(nil)
				streamOptions.SetStartAtOperationTime(nil)
				return c.getStream(ctx, streamOptions)
			}
		}
	}
	return stream, err
}
