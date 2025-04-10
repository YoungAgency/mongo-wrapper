package transaction

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ExecFn is a function that executes a transaction (mongo session).
// It receives a context.Context as parameter, that contains the mongo session.
// Session can be retrieved from context using mongo.SessionFromContext(ctx)
type ExecFn func(sessCtx context.Context) (interface{}, error)

// ExecSnapshot opens a session with Snapshot options using given client.
// Session is closed when this method returns
func ExecSnapshot(ctx context.Context, client *mongo.Client, fn ExecFn) (interface{}, error) {
	opts := SnapshotOptions()
	optsList := options.Lister[options.SessionOptions](opts)
	session, err := client.StartSession(optsList)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	res, err := session.WithTransaction(ctx, fn)
	return res, err
}
