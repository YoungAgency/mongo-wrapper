package query

import "go.mongodb.org/mongo-driver/v2/bson"

// UpdateBuilder struct permits to create a mongodb update
type UpdateBuilder struct {
	set  bson.D
	push bson.D
	inc  bson.D
}

func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		set:  make([]bson.E, 0),
		push: make([]bson.E, 0),
		inc:  make([]bson.E, 0),
	}
}

func (b *UpdateBuilder) Set(field string, value interface{}) *UpdateBuilder {
	b.set = append(b.set, bsonE(field, value))
	return b
}

func (b *UpdateBuilder) Inc(field string, value interface{}) *UpdateBuilder {
	b.inc = append(b.inc, bsonE(field, value))
	return b
}

func (b *UpdateBuilder) Push(field string, values ...interface{}) *UpdateBuilder {
	if len(values) == 0 {
		return b
	}
	var pushUpdate bson.E
	if len(values) == 1 {
		pushUpdate = bson.E{
			Key:   field,
			Value: values[0],
		}
	} else {
		pushUpdate = bson.E{
			Key: field,
			Value: bson.M{
				"$each": values,
			},
		}
	}
	b.push = append(b.push, pushUpdate)
	return b
}

// Update returns the current update document for builder
// Deprecated: use Build() instead
func (b *UpdateBuilder) Update() bson.D {
	ret := bson.D{}
	if len(b.set) > 0 {
		ret = append(ret, bsonE("$set", b.set))
	}
	if len(b.push) > 0 {
		ret = append(ret, bsonE("$push", b.push))
	}
	if len(b.inc) > 0 {
		ret = append(ret, bsonE("$inc", b.inc))
	}
	return ret
}

func (b *UpdateBuilder) Build() bson.D {
	ret := bson.D{}
	if len(b.set) > 0 {
		ret = append(ret, bsonE("$set", b.set))
	}
	if len(b.push) > 0 {
		ret = append(ret, bsonE("$push", b.push))
	}
	if len(b.inc) > 0 {
		ret = append(ret, bsonE("$inc", b.inc))
	}
	return ret
}

func bsonE(key string, value interface{}) bson.E {
	return bson.E{Key: key, Value: value}
}
