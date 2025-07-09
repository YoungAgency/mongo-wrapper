package query

import "go.mongodb.org/mongo-driver/v2/bson"

// FilterBuilder struct allows to build a mongodb query filter
type FilterBuilder struct {
	doc bson.D
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		doc: make([]bson.E, 0, 2),
	}
}

func (b *FilterBuilder) Eq(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "=", value))
	return b
}

func (b *FilterBuilder) Ne(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "!=", value))
	return b
}

func (b *FilterBuilder) Lt(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "<", value))
	return b
}

func (b *FilterBuilder) Lte(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, "<=", value))
	return b
}

func (b *FilterBuilder) Gt(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, ">", value))
	return b
}

func (b *FilterBuilder) Gte(field string, value any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldCompare(field, ">=", value))
	return b
}

func (b *FilterBuilder) Range(field string, from, to any) *FilterBuilder {
	b.doc = MergeDocuments(b.doc, FieldRange(field, from, to))
	return b
}

func (b *FilterBuilder) In(field string, value ...any) *FilterBuilder {
	inFilter := FieldIn(field, value...)
	b.doc = MergeDocuments(b.doc, inFilter)
	return b
}

func (b *FilterBuilder) Nin(field string, value ...any) *FilterBuilder {
	inFilter := FieldNotIn(field, value...)
	b.doc = MergeDocuments(b.doc, inFilter)
	return b
}

func (b *FilterBuilder) Exists(field string) *FilterBuilder {
	d := bson.D{
		{field, bson.D{{"$exists", true}}},
	}
	b.doc = MergeDocuments(b.doc, d)
	return b
}

func (b *FilterBuilder) NotExists(field string) *FilterBuilder {
	d := bson.D{
		{field, bson.D{{"$exists", false}}},
	}
	b.doc = MergeDocuments(b.doc, d)
	return b
}

// Deprecated: use Build() instead
func (b FilterBuilder) Filter() bson.D {
	return b.doc
}

func (b FilterBuilder) Build() bson.D {
	return b.doc
}

func (b *FilterBuilder) Reset() {
	b.doc = make([]bson.E, 0, 2)
}

// Set current builder document, it overwrite existing one
func (b *FilterBuilder) Set(doc bson.D) {
	b.doc = doc
}

// Append appends given bson.E to current query document
func (b *FilterBuilder) Append(e bson.E) *FilterBuilder {
	b.doc = append(b.doc, e)
	return b
}
