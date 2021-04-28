package transaction

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Creates a mongo.SessionContext
func CreateTransaction(ctx context.Context, client *mongo.Client, opts *options.SessionOptions, commit chan bool) (chan mongo.SessionContext, chan error) {
	out := make(chan mongo.SessionContext)
	out2 := make(chan error)
	go func() {
		defer close(out)
		defer close(out2)
		session, err := client.StartSession(opts)
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

		mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
			out <- sc
			select {
			case success := <-commit:
				if success {
					out2 <- sc.CommitTransaction(ctx)
				} else {
					out2 <- sc.AbortTransaction(ctx)
				}
			case <-ctx.Done():
				out2 <- sc.AbortTransaction(context.TODO())
			}
			return nil
		})
	}()
	return out, out2
}

// SnapshotOptions returns session options with snapshot readconcern.
// If attached to Session, this options will be applied to all transaction runned within that session.
func SnapshotOptions() *options.SessionOptions {
	return options.Session().
		SetCausalConsistency(false).
		SetDefaultReadConcern(readconcern.Snapshot()).
		SetDefaultReadPreference(readpref.Primary()).
		SetDefaultWriteConcern(writeconcern.New(writeconcern.WMajority()))
}
