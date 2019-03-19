package query

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FieldIn returns a bson to document to use as MongoDB query filter
// Documents which field value is contained in vals will be returned
func FieldIn(field string, vals ...string) bson.D {
	a := make(bson.A, len(vals))
	for i, v := range vals {
		a[i] = v
	}

	return bson.D{
		{
			Key: field,
			Value: bson.D{
				{"$in", a},
			},
		},
	}
}

func Or(documents ...bson.D) bson.D {
	a := make(bson.A, len(documents))
	for i, d := range documents {
		a[i] = d
	}

	return bson.D{
		{
			Key:   "$or",
			Value: a,
		},
	}
}

func And(documents ...bson.D) bson.D {
	a := make(bson.A, len(documents))
	for i, d := range documents {
		a[i] = d
	}

	return bson.D{
		{
			Key:   "$and",
			Value: a,
		},
	}
}

// FieldNotIn returns a bson to document to use as MongoDB query filter
// Documents which field value is not contained in vals will be returned
func FieldNotIn(field string, vals ...string) bson.D {
	a := make(bson.A, len(vals))
	for i, v := range vals {
		a[i] = v
	}

	return bson.D{
		{
			Key: field,
			Value: bson.D{
				{"$nin", a},
			},
		},
	}
}

// FieldCompare return a bson document to use as MongoDB query filter
func FieldCompare(field string, op string, val interface{}) bson.D {
	var mongoOp string
	switch op {
	case "<":
		mongoOp = "$lt"
		break
	case "<=":
		mongoOp = "$lte"
		break
	case "=":
		mongoOp = "$eq"
		break
	case ">":
		mongoOp = "$gt"
		break
	case ">=":
		mongoOp = "$gte"
		break
	case "!=":
		mongoOp = "$ne"
		break
	default:
		panic("Invalid operator")
	}

	return bson.D{
		{
			Key: field,
			Value: bson.D{
				{
					Key:   mongoOp,
					Value: val,
				},
			},
		},
	}
}

// MergeDocuments returns a new document which contains all given documents fields
// Duplicated fields are not handled
func MergeDocuments(documents ...bson.D) bson.D {
	e := make([]bson.E, 0, len(documents))
	for _, d := range documents {
		for _, a := range d {
			e = append(e, a)
		}
	}
	return bson.D(e)
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
