package stream

import (
	"context"
	"encoding/json"
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

func (c *Consumer[T, K]) ConsumeInList(ctx context.Context, streamOptions *options.ChangeStreamOptionsBuilder, handler HandlerFn[T, K]) error {
	stream, err := c.getStream(ctx, streamOptions)
	if err != nil {
		return err
	}
	tokenMan := c.tokenManager.(OffsetManagerList)
	fn := func(ctx context.Context, doc StreamEvent[T, K]) error {
		jsonMsg, err := json.Marshal(doc)
		if err != nil {
			return nil
		}
		err = tokenMan.SetOffsetAndPush(ctx, StreamOffset{
			ResumeToken: doc.ID.Data,
			Timestamp:   doc.ClusterTime,
		}, jsonMsg)

		return err
	}
	return c.consumeInternal(ctx, stream, handler, fn)
}

func (c *Consumer[T, K]) ConsumePublish(ctx context.Context, streamOptions *options.ChangeStreamOptionsBuilder, msgFn func(event StreamEvent[T, K]) (channel, msg string)) error {
	stream, err := c.getStream(ctx, streamOptions)
	if err != nil {
		return err
	}
	defer stream.Close(ctx)
	tokenMan := c.tokenManager.(OffsetManagerList)
	handler := func(ctx context.Context, doc StreamEvent[T, K]) error {
		return nil
	}
	fnb := func(ctx context.Context, doc StreamEvent[T, K]) error {
		channel, msg := msgFn(doc)
		so := doc.GetStreamOffset()
		if channel == "" {
			return tokenMan.SetOffset(ctx, *so)
		}
		return tokenMan.SetOffsetAndPublish(ctx, so, channel, msg)
	}
	return c.consumeInternal(ctx, stream, handler, fnb)
}

func (c *Consumer[T, K]) consumeInternal(ctx context.Context, stream *mongo.ChangeStream, handler HandlerFn[T, K], fn func(ctx context.Context, doc StreamEvent[T, K]) error) error {
	doc := StreamEvent[T, K]{}
	lastErr := error(nil)

	for stream.Next(ctx) {
		if err := stream.Decode(&doc); err != nil {
			return err
		}
		for {
			if err := handler(ctx, doc); err != nil {
				<-time.After(1 * time.Second)
				continue
			}
			success := false
			retries := 3
			for retries > 0 && !success {
				if err := fn(ctx, doc); err != nil {
					<-time.After(1 * time.Second)
					retries--
					lastErr = err
				}
				success = true
			}
			if success {
				break
			} else {
				return lastErr
			}
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
			if mongoErr.Code == 286 {
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
