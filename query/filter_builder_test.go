package query

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBuilder_In(t *testing.T) {
	type args struct {
		field string
		value []interface{}
	}
	tests := []struct {
		name string
		args args
		want bson.D
	}{
		{
			name: "In query produces correct query document",
			args: args{
				field: "_id",
				value: []interface{}{1, 2, 3, 4},
			},
			want: bson.D{
				bson.E{
					Key: "_id",
					Value: bson.D{
						bson.E{
							Key: "$in",
							Value: bson.A{
								1, 2, 3, 4,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewFilterBuilder()
			if got := b.In(tt.args.field, tt.args.value...).Build(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.In() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_Nin(t *testing.T) {
	type args struct {
		field string
		value []interface{}
	}
	tests := []struct {
		name string
		args args
		want bson.D
	}{
		{
			name: "Nin query produces correct query document",
			args: args{
				field: "_id",
				value: []interface{}{5, 6, 7},
			},
			want: bson.D{
				bson.E{
					Key: "_id",
					Value: bson.D{
						bson.E{
							Key: "$nin",
							Value: bson.A{
								5, 6, 7,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewFilterBuilder()
			if got := b.Nin(tt.args.field, tt.args.value...).Build(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.Nin() = %v, want %v", got, tt.want)
			}
		})
	}
}
