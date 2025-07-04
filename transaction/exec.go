package transaction

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type ExecFn func(ctx context.Context) (any, error)

// mongoExecSnapshot opens a session with Snapshot options using given client.
// Session is closed when this method returns
func ExecSnapshot(ctx context.Context, client *mongo.Client, fn ExecFn) (any, error) {
	opt := mongoSnapshotOptions()
	session, err := client.StartSession(opt)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	res, err := session.WithTransaction(ctx, fn)
	return res, err
}

type MongoTransactionRunner struct {
	client *mongo.Client
}

func NewMongoTransactionRunner(client *mongo.Client) *MongoTransactionRunner {
	return &MongoTransactionRunner{
		client: client,
	}
}

func (m *MongoTransactionRunner) ExecSnapshot(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	// check if ctx is already a session context
	if session := mongo.SessionFromContext(ctx); session != nil {
		// If it is, we can use the existing session
		return fn(ctx)
	}
	opt := mongoSnapshotOptions()
	session, err := m.client.StartSession(opt)
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	// Use session.WithTransaction to run the function in a transaction
	res, err := session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		return fn(ctx)
	})

	return res, err
}

// mongoSnapshotOptions returns session options with snapshot readconcern.
// If attached to Session, this options will be applied to all transaction runned within that session.
func mongoSnapshotOptions() *options.SessionOptionsBuilder {
	txnOpts := options.Transaction().
		SetReadPreference(readpref.Primary()).
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())
	sessOpts := options.Session().SetDefaultTransactionOptions(txnOpts)
	return sessOpts
}
