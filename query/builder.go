package query

import "go.mongodb.org/mongo-driver/v2/bson"

// Builder struct allows to build a mongodb query filter
type Builder struct {
	doc bson.D
}

func NewBuilder() *Builder {
	return &Builder{
		doc: make([]bson.E, 0, 2),
	}
}

func (b *Builder) Eq(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "=", value))
	return b
}

func (b *Builder) Ne(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "!=", value))
	return b
}

func (b *Builder) Lt(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "<", value))
	return b
}

func (b *Builder) Lte(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "<=", value))
	return b
}

func (b *Builder) Gt(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, ">", value))
	return b
}

func (b *Builder) Gte(field string, value interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, ">=", value))
	return b
}

func (b *Builder) Range(field string, from, to interface{}) *Builder {
	b.doc = MergeDocuments(b.doc, FieldRange(field, from, to))
	return b
}

func (b *Builder) In(field string, value ...interface{}) *Builder {
	inFilter := FieldIn(field, value...)
	b.doc = MergeDocuments(b.doc, inFilter)
	return b
}

func (b *Builder) Nin(field string, value ...interface{}) *Builder {
	inFilter := FieldNotIn(field, value...)
	b.doc = MergeDocuments(b.doc, inFilter)
	return b
}

// Deprecated: use Build() instead
func (b Builder) Filter() bson.D {
	return b.doc
}

func (b Builder) Build() bson.D {
	return b.doc
}

func (b *Builder) Reset() {
	b.doc = make([]bson.E, 0, 2)
}

// Set current builder document, it overwrite existing one
func (b *Builder) Set(doc bson.D) {
	b.doc = doc
}

// Append appends given bson.E to current query document
func (b *Builder) Append(e bson.E) *Builder {
	b.doc = append(b.doc, e)
	return b
}
