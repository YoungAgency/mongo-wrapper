package stream

import (
	"context"
	"time"
)

type StreamEvent[T any, K any] struct {
	ID struct {
		Data string `bson:"_data" json:"_data"`
	} `bson:"_id" json:"_id"`
	DocumentKey struct {
		ID K `bson:"_id" json:"_id"`
	} `bson:"documentKey" json:"documentKey"`
	OperationType string    `bson:"operationType" json:"operationType"`
	FullDocument  T         `bson:"fullDocument" json:"fullDocument"`
	ClusterTime   time.Time `bson:"clusterTime" json:"clusterTime"`
	// CollectionUUID primitive.Binary `bson:"collectionUUID" json:"collectionUUID"`
	// WallTime time.Time `bson:"wallTime" json:"wallTime"`
	NS struct {
		DB   string `bson:"db" json:"db"`
		Coll string `bson:"coll" json:"coll"`
	} `bson:"ns" json:"ns"`
}

type StreamOffset struct {
	ResumeToken string    `json:"token"`
	Timestamp   time.Time `json:"ts"`
}

type EventEncoder interface {
	Encode(raw []byte) ([]byte, error)
}

type OffsetManagerList interface {
	OffsetManager
	SetOffsetAndPush(ctx context.Context, offset StreamOffset, msg []byte) error
}

type OffsetManager interface {
	GetOffset(ctx context.Context) (*StreamOffset, error)
	SetOffset(ctx context.Context, offset StreamOffset) error
}
