package query

import (
	"go.mongodb.org/mongo-driver/bson"
)

// FieldIn returns a bson to document to use as MongoDB query filter
// Documents which field value is contained in vals will be returned
func FieldIn(field string, vals ...interface{}) bson.D {
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
func FieldNotIn(field string, vals ...interface{}) bson.D {
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

// FieldRange return a bson document rapresenting query by range on given fields
// from and to may be included in results
func FieldRange(field string, from interface{}, to interface{}) bson.D {
	return And(
		FieldCompare(field, ">=", from),
		FieldCompare(field, "<=", to),
	)
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

// Update returns a bson Document rapresenting the update to perform
func Update(op string, e ...bson.E) bson.D {
	var mongoOp string
	switch op {
	case "inc":
		mongoOp = "$inc"
		break
	case "set":
		mongoOp = "$set"
		break
	case "push":
		mongoOp = "$push"
		break
	default:
		panic("Invalid operator")
	}

	return bson.D{
		{
			Key:   mongoOp,
			Value: e,
		},
	}
}
