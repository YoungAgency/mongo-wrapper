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
		name   string
		fields fields
		args   args
		want   *UpdateBuilder
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &UpdateBuilder{
				set:  tt.fields.set,
				push: tt.fields.push,
				inc:  tt.fields.inc,
			}
			if got := b.Set(tt.args.field, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateBuilder.Set() = %v, want %v", got, tt.want)
			}
		})
	}
}
