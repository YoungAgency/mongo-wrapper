package errors

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// DuplicateKey returns true if error rapresent Mongo error_code DuplicateKey
func DuplicateKey(err error) bool {
	if commandError, ok := err.(mongo.CommandError); ok {
		return commandError.Code == 11000
	}
	return false
}
