package query

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestUpdateBuilder_Set(t *testing.T) {
	type fields struct {
		set  bson.D
		push bson.D
		inc  bson.D
	}
	type args struct {
		field string
		value interface{}
	}
	tests := []struct {
		name    string
		initial *UpdateBuilder
		args    args
		want    *UpdateBuilder
	}{
		{
			name:    "Test set with new update builder",
			initial: NewUpdateBuilder(),
			args: args{
				field: "foo",
				value: "bar",
			},
			want: &UpdateBuilder{
				set: bson.D{
					bsonE("foo", "bar"),
				},
				push: make([]bson.E, 0),
				inc:  make([]bson.E, 0),
			},
		},
		{
			name: "Test set with already used builder",
			initial: &UpdateBuilder{
				set:  bson.D{bsonE("foo", "bar")},
				push: make([]bson.E, 0),
				inc:  make([]bson.E, 0),
			},
			args: args{
				field: "foo2",
				value: "bar2",
			},
			want: &UpdateBuilder{
				set: bson.D{
					bsonE("foo", "bar"),
					bsonE("foo2", "bar2"),
				},
				push: make([]bson.E, 0),
				inc:  make([]bson.E, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.initial
			if got := b.Set(tt.args.field, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBuilder.Set() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUpdateBuilder_Inc(t *testing.T) {
	type fields struct {
		set  bson.D
		push bson.D
		inc  bson.D
	}
	type args struct {
		field string
		value interface{}
	}
	tests := []struct {
		name    string
		initial *UpdateBuilder
		args    args
		want    *UpdateBuilder
	}{
		{
			name:    "Test inc with new update builder",
			initial: NewUpdateBuilder(),
			args: args{
				field: "foo",
				value: 1,
			},
			want: &UpdateBuilder{
				set:  make([]bson.E, 0),
				push: make([]bson.E, 0),
				inc: bson.D{
					bsonE("foo", 1),
				},
			},
		},
		{
			name: "Test inc with already used builder",
			initial: &UpdateBuilder{
				set:  make([]bson.E, 0),
				inc:  bson.D{bsonE("foo", 1)},
				push: make([]bson.E, 0),
			},
			args: args{
				field: "foo2",
				value: 2,
			},
			want: &UpdateBuilder{
				set: make([]bson.E, 0),
				inc: bson.D{
					bsonE("foo", 1),
					bsonE("foo2", 2),
				},
				push: make([]bson.E, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.initial
			if got := b.Inc(tt.args.field, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBuilder.Inc() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUpdateBuilder_Push(t *testing.T) {
	type fields struct {
		set  bson.D
		push bson.D
		inc  bson.D
	}
	type args struct {
		field  string
		values []interface{}
	}
	tests := []struct {
		name    string
		initial *UpdateBuilder
		args    args
		want    *UpdateBuilder
	}{
		{
			name:    "Test push single value with new update builder",
			initial: NewUpdateBuilder(),
			args: args{
				field:  "foo",
				values: []interface{}{1},
			},
			want: &UpdateBuilder{
				set: make([]bson.E, 0),
				push: bson.D{
					bsonE("foo", 1),
				},
				inc: make([]bson.E, 0),
			},
		},
		{
			name:    "Test push multiple values with new update builder",
			initial: NewUpdateBuilder(),
			args: args{
				field:  "foo",
				values: []interface{}{1, 2, 3},
			},
			want: &UpdateBuilder{
				set: make([]bson.E, 0),
				push: bson.D{
					bsonE("foo", bson.M{"$each": []interface{}{1, 2, 3}}),
				},
				inc: make([]bson.E, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.initial
			if got := b.Push(tt.args.field, tt.args.values...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBuilder.Push() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUpdateBuilder_Update(t *testing.T) {
	type fields struct {
		set  bson.D
		push bson.D
		inc  bson.D
	}

	tests := []struct {
		name   string
		fields fields
		want   bson.D
	}{
		{
			name: "Test with empty inc and push",
			fields: fields{
				set: bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", 2),
				},
				push: nil,
				inc:  nil,
			},
			want: bson.D{
				bsonE("$set", bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", 2),
				}),
			},
		},
		{
			name: "Test with empty inc",
			fields: fields{
				set: bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", 2),
				},
				push: bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", bson.M{"$each": []interface{}{1, 2, 3}}),
				},
				inc: nil,
			},
			want: bson.D{
				bsonE("$set", bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", 2),
				}),
				bsonE("$push", bson.D{
					bsonE("foo", "bar"),
					bsonE("baz", bson.M{"$each": []interface{}{1, 2, 3}}),
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &UpdateBuilder{
				set:  tt.fields.set,
				push: tt.fields.push,
				inc:  tt.fields.inc,
			}
			if got := b.Update(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBuilder.Update() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
