package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
	"go.uber.org/zap"
)

// RetryingStorage wraps StorageReaderWriter for read side
type RetryingStorage struct {
	upstream StorageReaderWriter
	logger   *zap.Logger
}

// NewRetryingStorage creates new storage with retries
func NewRetryingStorage(upstream StorageReaderWriter, logger *zap.Logger) *RetryingStorage {
	return &RetryingStorage{upstream: upstream, logger: logger}
}

// FindOne returns a row into destination.
func (s *RetryingStorage) FindOne(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	return s.retry(func() error {
		return s.upstream.FindOne(ctx, collection, filter, dest)
	})
}

// FindAll returns all rows matching filter into destination.
func (s *RetryingStorage) FindAll(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	return s.retry(func() error {
		return s.upstream.FindAll(ctx, collection, filter, dest)
	})
}

// FindMany returns rows into destination.
func (s *RetryingStorage) FindMany(ctx context.Context, collection string, filter interface{}, limit, offset uint64, sort string, dest interface{}) (total uint64, err error) {
	err = s.retry(func() error {
		total, err = s.upstream.FindMany(ctx, collection, filter, limit, offset, sort, dest)
		return err
	})

	return total, err
}

// RunInTransaction encapsulates the function that needs to run in a transaction.
func (s *RetryingStorage) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return s.upstream.RunInTransaction(ctx, fn)
}

// Insert makes insert into database.
func (s *RetryingStorage) Insert(ctx context.Context, collection string, document interface{}) error {
	return s.upstream.Insert(ctx, collection, document)
}

// Update updates documents in the database.
func (s *RetryingStorage) Update(ctx context.Context, collection string, docID primitive.ObjectID, update interface{}) (modifiedCount int64, err error) {
	return s.upstream.Update(ctx, collection, docID, update)
}

// Upsert updates or inserts document in the database.
func (s *RetryingStorage) Upsert(ctx context.Context, collection string, docID interface{}, update interface{}) (upsertedCount int64, err error) {
	return s.upstream.Upsert(ctx, collection, docID, update)
}

// Delete deletes document in the database.
func (s *RetryingStorage) Delete(ctx context.Context, collection string, docID primitive.ObjectID) (deletedCount int64, err error) {
	return s.upstream.Delete(ctx, collection, docID)
}

// DeleteMany deletes filtered documents in the database.
func (s *RetryingStorage) DeleteMany(ctx context.Context, collection string, filter interface{}) (deletedCount int64, err error) {
	return s.upstream.DeleteMany(ctx, collection, filter)
}

// GetDatabaseName returns the name of the current database.
func (s *RetryingStorage) GetDatabaseName() string {
	return s.upstream.GetDatabaseName()
}

// retry keeps trying the function until the second argument returns false, or no error is returned.
// Adapted from https://github.com/matryer/try/blob/master/try.go
func (s *RetryingStorage) retry(fn func() (err error)) error {
	const maxRetries = 10

	var err error
	attempt := 1
	for {
		if attempt > maxRetries {
			return errors.Wrap(err, "exceeded retry limit")
		}

		err = fn()
		if err == nil {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			break
		}

		if errors.Is(err, mongo.ErrClientDisconnected) {
			s.logger.Info("retrying mongodb client disconnected",
				zap.Int("attempt", attempt), zap.String("error", err.Error()))

			time.Sleep(10 * time.Duration(attempt) * time.Millisecond)
			attempt++
			continue
		}

		if mongo.IsTimeout(err) {
			s.logger.Info("retrying mongodb timeout",
				zap.Int("attempt", attempt), zap.String("error", err.Error()))

			time.Sleep(10 * time.Duration(attempt) * time.Millisecond)
			attempt++
			continue
		}

		if mongo.IsNetworkError(err) {
			s.logger.Info("retrying mongodb network error",
				zap.Int("attempt", attempt), zap.String("error", err.Error()))

			time.Sleep(10 * time.Duration(attempt) * time.Millisecond)
			attempt++
			continue
		}

		if _, ok := err.(driver.RetryablePoolError); ok {
			s.logger.Info("retrying mongodb pool error",
				zap.Int("attempt", attempt), zap.String("error", err.Error()))

			time.Sleep(10 * time.Duration(attempt) * time.Millisecond)
			attempt++
			continue
		}

		var waitQueueTimeoutError topology.WaitQueueTimeoutError
		if errors.As(err, &waitQueueTimeoutError) {
			s.logger.Info("retrying WaitQueueTimeoutError",
				zap.Int("attempt", attempt), zap.String("error", err.Error()))

			time.Sleep(10 * time.Duration(attempt) * time.Millisecond)
			attempt++
			continue
		}

		// If we got here, we don't need to retry
		break
	}

	return err
}
