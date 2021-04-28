package transaction

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type ExecFn func(sessCtx mongo.SessionContext) (interface{}, error)

// ExecSnapshot opens a session with Snapshot options using given client.
// Session is closed when this method returns
func ExecSnapshot(ctx context.Context, client *mongo.Client, fn ExecFn) (interface{}, error) {
	session, err := client.StartSession(SnapshotOptions())
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	res, err := session.WithTransaction(ctx, fn)
	return res, err
}
