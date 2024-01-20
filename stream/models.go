package stream

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type StreamEvent[T any, K any] struct {
	ID struct {
		Data string `bson:"_data" json:"_data"`
	} `bson:"_id" json:"_id"`
	DocumentKey struct {
		ID K `bson:"_id" json:"_id"`
	} `bson:"documentKey" json:"documentKey"`
	OperationType            string             `bson:"operationType" json:"operationType"`
	FullDocument             T                  `bson:"fullDocument" json:"fullDocument"`
	ClusterTime              time.Time          `bson:"clusterTime" json:"clusterTime"`
	FullDocumentBeforeChange *T                 `bson:"fullDocumentBeforeChange" json:"fullDocumentBeforeChange"` // needs changeStreamPreAndPostImages
	UpdateDescription        *UpdateDescription `bson:"updateDescription" json:"updateDescription"`
	NS                       struct {
		DB   string `bson:"db" json:"db"`
		Coll string `bson:"coll" json:"coll"`
	} `bson:"ns" json:"ns"`
}

func (s StreamEvent[T, K]) GetStreamOffset() *StreamOffset {
	return &StreamOffset{
		ResumeToken: s.ID.Data,
		Timestamp:   s.ClusterTime,
	}
}

// CollectionUUID primitive.Binary `bson:"collectionUUID" json:"collectionUUID"`
// WallTime time.Time `bson:"wallTime" json:"wallTime"`

type UpdateDescription struct {
	UpdatedFields   map[string]interface{}  `bson:"updatedFields" json:"updatedFields"`
	RemovedFields   []string                `bson:"removedFields" json:"removedFields"`
	TruncatedArrays []TruncatedArrayElement `bson:"truncatedArrays" json:"truncatedArrays"`
}

func (ud UpdateDescription) GetUpdatedObject(t reflect.Type) interface{} {
	if t.Kind() != reflect.Struct {
		panic("type must be struct")
	}
	obj := reflect.New(t).Interface()
	// TODO understand if it makes sense to define
	// UpdateDescription as generic and unmarsal directly into obj
	// this is most flexible but also not very nice or "performant"
	bb, err := bson.Marshal(ud.UpdatedFields)
	if err != nil {
		panic(err)
	}
	err = bson.Unmarshal(bb, obj)
	if err != nil {
		panic(err)
	}
	return obj
}

type TruncatedArrayElement struct {
	Field   string `bson:"field" json:"field"`
	NewSize int    `bson:"newSize" json:"newSize"`
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
	SetOffsetAndPublish(ctx context.Context, offset *StreamOffset, channel, msg string) error
}

type OffsetManager interface {
	GetOffset(ctx context.Context) (*StreamOffset, error)
	SetOffset(ctx context.Context, offset StreamOffset) error
}
