package errors

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	duplicateCode = 11000
)

// DuplicateKey returns true if error rapresent Mongo error_code DuplicateKey
func DuplicateKey(err error) bool {
	if commandError, ok := err.(mongo.CommandError); ok {
		return commandError.Code == duplicateCode
	}
	return false
}
