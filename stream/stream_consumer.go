package stream

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Database     string
	Collection   string
	Encoder      EventEncoder
	TokenManager OffsetManager
}

type Consumer[T any, K any] struct {
	client       *mongo.Client
	encoder      EventEncoder
	tokenManager OffsetManager
	database     string
	collection   string
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
		encoder:      encoder,
		tokenManager: tokenManger,
		client:       client,
		database:     conf.Database,
		collection:   conf.Collection,
	}
}

func (c *Consumer[T, K]) ConsumeHandler(ctx context.Context, streamOptions *options.ChangeStreamOptions, handler HandlerFn[T, K]) error {
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

func (c *Consumer[T, K]) ConsumeInList(ctx context.Context, streamOptions *options.ChangeStreamOptions) error {
	stream, err := c.getStream(ctx, streamOptions)
	if err != nil {
		return err
	}
	doc := StreamEvent[T, K]{}

	tokenMan := c.tokenManager.(OffsetManagerList)
	for stream.Next(ctx) {
		if err := stream.Decode(&doc); err != nil {
			return err
		}
		jsonMsg, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		if c.encoder != nil {
			jsonMsg, err = c.encoder.Encode(jsonMsg)
			if err != nil {
				return err
			}
		}
		err = tokenMan.SetOffsetAndPush(ctx, StreamOffset{
			ResumeToken: doc.ID.Data,
			Timestamp:   doc.ClusterTime,
		}, jsonMsg)
		if err != nil {
			return err
		}
	}
	return stream.Close(ctx)
}

func (c *Consumer[T, K]) getStream(ctx context.Context, streamOptions *options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	resumeToken, err := c.tokenManager.GetOffset(ctx)
	if err != nil {
		return nil, err
	}
	if resumeToken != nil && resumeToken.ResumeToken != "" {
		streamOptions.SetStartAfter(primitive.M{
			"_data": resumeToken.ResumeToken,
		})
	} else if resumeToken != nil && !resumeToken.Timestamp.IsZero() {
		// if timestamp is out of range for oplog it will be ignored
		dt := &primitive.Timestamp{T: uint32(resumeToken.Timestamp.UTC().Unix()), I: 0}
		streamOptions.SetStartAtOperationTime(dt)
	}
	stream, err := c.client.Database(c.database).
		Collection(c.collection).
		Watch(ctx, primitive.D{}, streamOptions)
	return stream, err
}