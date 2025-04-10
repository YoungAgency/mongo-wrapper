package transaction

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// Creates a mongo.SessionContext
func CreateTransaction(ctx context.Context, client *mongo.Client, commit chan bool, opts ...options.Lister[options.SessionOptions]) (chan context.Context, chan error) {
	out := make(chan context.Context)
	out2 := make(chan error)

	go func() {
		defer close(out)
		defer close(out2)

		session, err := client.StartSession(opts...)
		if err != nil {
			out2 <- err
			return
		}
		defer session.EndSession(ctx)
		err = session.StartTransaction()
		if err != nil {
			out2 <- err
			return
		}

		mongo.WithSession(ctx, session, func(sc context.Context) error {
			out <- sc
			select {
			case success := <-commit:
				if success {
					out2 <- session.CommitTransaction(ctx)
				} else {
					out2 <- session.AbortTransaction(ctx)
				}
			case <-ctx.Done():
				out2 <- session.AbortTransaction(context.TODO())
			}
			return nil
		})
	}()
	return out, out2
}

// SnapshotOptions returns session options builder with snapshot readconcern.
// If wrapped into a options.Lister it can be attached to a Session, this options will be applied
// to all transaction runned within that session.
func SnapshotOptions() *options.SessionOptionsBuilder {
	wc := writeconcern.Majority()

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetReadPreference(readpref.Primary()).
		SetWriteConcern(wc)

	return options.Session().
		SetCausalConsistency(false).
		SetDefaultTransactionOptions(txnOpts)
}
