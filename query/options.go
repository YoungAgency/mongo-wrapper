package query

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOptions is a wrapper around options.FindOptions
type FindOptions struct {
	options *options.FindOptions
	sort    []SortStruct
}

func NewFindOptions() *FindOptions {
	return &FindOptions{
		options: options.Find(),
		sort:    make([]SortStruct, 0),
	}
}

func (o *FindOptions) Sort(field string, ascending bool) *FindOptions {
	o.sort = append(o.sort, SortStruct{Field: field, Ascending: ascending})
	return o
}

func (o *FindOptions) Page(batch, page int) *FindOptions {
	o.options = Page(o.options, batch, page)
	return o
}

func (o FindOptions) Options() *options.FindOptions {
	o.options = Sort(o.options, o.sort...)
	return o.options
}

// Sort set options sort with given sort struct
func Sort(options *options.FindOptions, ss ...SortStruct) *options.FindOptions {
	e := make([]bson.E, len(ss))
	for i, s := range ss {
		order := -1 // descending
		if s.Ascending {
			order = 1
		}
		e[i] = bson.E{
			Key:   s.Field,
			Value: order,
		}
	}
	return options.SetSort(bson.D(e))
}

// Page set batch size and correct skip on given options
// Use for paginated requests
func Page(opt *options.FindOptions, batch int, page int) *options.FindOptions {
	return opt.SetBatchSize(int32(batch)).
		SetSkip(int64(batch * (page - 1))).SetLimit(int64(batch))
}

// Pull returns a bson document with the pull operation
func Pull(ps ...PullStruct) bson.D {
	e := make([]bson.E, len(ps))
	for i, s := range ps {
		e[i] = bson.E{
			Key:   s.Key,
			Value: s.Filter,
		}
	}
	return bson.D{
		{
			Key:   "$pull",
			Value: e,
		},
	}
}

// PullStruct is used to represent pull in MongoDB
type PullStruct struct {
	Key    string // Key of array
	Filter bson.D // Filter
}

// SortStruct is used to represent sort in MongoDB
type SortStruct struct {
	Field     string // Document field name
	Ascending bool   // Sort order
}

func DefaultFindOneOptions() *options.FindOptions {
	options := options.Find()
	options.SetLimit(int64(1)).SetBatchSize(int32(1))
	return options
}
