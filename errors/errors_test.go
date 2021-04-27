package errors

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestDuplicateKey(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"test correct code should return true",
			args{
				err: mongo.CommandError{Code: duplicateCode},
			},
			true,
		},
		{
			"test incorrect code should return false",
			args{
				err: mongo.CommandError{Code: 1246558},
			},
			false,
		},
		{
			"test with nil error should return false",
			args{
				err: nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DuplicateKey(tt.args.err); got != tt.want {
				t.Errorf("DuplicateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"test correct code should return true",
			args{
				err: mongo.WriteException{
					WriteErrors: mongo.WriteErrors{
						mongo.WriteError{
							Code: validationCode,
						},
					},
				},
			},
			true,
		},
		{
			"test incorrect code should return false",
			args{
				err: mongo.WriteException{
					WriteErrors: mongo.WriteErrors{
						mongo.WriteError{
							Code: 12323123,
						},
					},
				},
			},
			false,
		},
		{
			"test with nil error should return false",
			args{
				err: nil,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Validation(tt.args.err); got != tt.want {
				t.Errorf("Validation() = %v, want %v", got, tt.want)
			}
		})
	}
}
