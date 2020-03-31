package query

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOptions is a wrapper around options.FindOptions
type FindOptions struct {
	options    *options.FindOptions
	sort       []SortStruct
	projection bson.D
}

func NewFindOptions() *FindOptions {
	return &FindOptions{
		options:    options.Find(),
		sort:       make([]SortStruct, 0, 2),
		projection: make([]bson.E, 0, 2),
	}
}

func (o *FindOptions) Sort(field string, ascending bool) *FindOptions {
	o.sort = append(o.sort, SortStruct{Field: field, Ascending: ascending})
	return o
}

func (o *FindOptions) AscSort(field string) *FindOptions {
	return o.Sort(field, true)
}

func (o *FindOptions) DescSort(field string) *FindOptions {
	return o.Sort(field, false)
}

func (o *FindOptions) Page(batch, page int) *FindOptions {
	o.options = Page(o.options, batch, page)
	return o
}

func (o *FindOptions) Projection(fields ...string) *FindOptions {
	for _, field := range fields {
		o.projection = append(o.projection, bson.E{Key: field, Value: 1})
	}
	return o
}

// PageProjection applies a paged projection on array field elements
// according to page and batch params
func (o *FindOptions) PageProjection(field string, page, batch int) *FindOptions {
	e := bsonE(field, bson.M{
		"$slice": []int{(page - 1) * batch, batch},
	})
	o.projection = append(o.projection, e)
	return o
}

// ExProjection sets an exclusive projection on options
func (o *FindOptions) ExProjection(fields ...string) *FindOptions {
	for _, field := range fields {
		o.projection = append(o.projection, bson.E{Key: field, Value: 0})
	}
	return o
}

func (o FindOptions) Options() *options.FindOptions {
	o.options = Sort(o.options, o.sort...)
	o.options = o.options.SetProjection(o.projection)
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

func Projection(fields ...string) bson.D {
	ret := make([]bson.E, len(fields))
	for i, field := range fields {
		ret[i] = bson.E{
			Key:   field,
			Value: 1,
		}
	}
	return ret
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
