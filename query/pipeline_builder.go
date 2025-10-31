package query

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PipelineBuilder struct {
	pipeline []bson.D
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: make([]bson.D, 0, 2),
	}
}

func (pb *PipelineBuilder) Match(value interface{}) *PipelineBuilder {
	stage := bson.D{
		bson.E{
			Key:   "$match",
			Value: value,
		},
	}
	pb.pipeline = append(pb.pipeline, stage)
	return pb
}

func (pb *PipelineBuilder) Sort(value interface{}) *PipelineBuilder {
	stage := bson.D{
		bson.E{
			Key:   "$sort",
			Value: value,
		},
	}
	pb.pipeline = append(pb.pipeline, stage)
	return pb
}

func (pb *PipelineBuilder) Project(value interface{}) *PipelineBuilder {
	stage := bson.D{
		bson.E{
			Key:   "$project",
			Value: value,
		},
	}
	pb.pipeline = append(pb.pipeline, stage)
	return pb
}

func (pb *PipelineBuilder) AppendStage(stage bson.D) *PipelineBuilder {
	pb.pipeline = append(pb.pipeline, stage)
	return pb
}

func (pb *PipelineBuilder) Build() mongo.Pipeline {
	return pb.pipeline
}
