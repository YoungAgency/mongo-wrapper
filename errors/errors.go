package errors

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	duplicateCode  = 11000
	validationCode = 121
)

// DuplicateKey returns true if error rapresent Mongo error_code DuplicateKey
func DuplicateKey(err error) bool {
	if commandError, ok := err.(mongo.CommandError); ok {
		return commandError.Code == duplicateCode
	}
	if writeEx, ok := err.(mongo.WriteException); ok {
		if len(writeEx.WriteErrors) > 0 {
			return writeEx.WriteErrors[0].Code == duplicateCode
		}
	}
	return false
}

// Validation returns true if error represents Mongo error_code Validation
func Validation(err error) bool {
	if writeEx, ok := err.(mongo.WriteException); ok {
		if len(writeEx.WriteErrors) > 0 {
			return writeEx.WriteErrors[0].Code == validationCode
		}
	}
	return false
}
