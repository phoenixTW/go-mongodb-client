package mongodb

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// StorageReader describes interface for read operations for storage
type StorageReader interface {
	FindOne(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error)
	FindAll(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error)
	FindMany(
		ctx context.Context,
		collection string,
		filter interface{},
		limit, offset uint64,
		sort string,
		dest interface{},
	) (total uint64, err error)
}

// StorageWriter describes interface for write operations for storage
type StorageWriter interface {
	RunInTransaction(ctx context.Context, fn func(context.Context) error) error
	Insert(ctx context.Context, collection string, document interface{}) error
	Update(ctx context.Context, collection string, docID primitive.ObjectID, update interface{}) (modifiedCount int64, err error)
	Upsert(ctx context.Context, collection string, docID interface{}, update interface{}) (upsertedCount int64, err error)
	Delete(ctx context.Context, collection string, docID primitive.ObjectID) (deletedCount int64, err error)
	DeleteMany(ctx context.Context, collection string, filter interface{}) (deletedCount int64, err error)
}

// StorageReaderWriter describes interface for both read and write operations for storage
type StorageReaderWriter interface {
	StorageReader
	StorageWriter

	GetDatabaseName() string
}

// ObjectID will convert a string-compatible type to primitive.ObjectID
func ObjectID[T ~string](domainID T) primitive.ObjectID {
	// nolint: errcheck // reason: we trust domain validation
	objectID, _ := primitive.ObjectIDFromHex(string(domainID))

	return objectID
}

// Storage manages query builders and database requests.
type Storage struct {
	database *mongo.Database
}

// GetDatabaseName returns the name of the current database
func (s *Storage) GetDatabaseName() string {
	return s.database.Name()
}

// MakeStorage initializes database storage.
func MakeStorage(db *mongo.Database) StorageReaderWriter {
	return &Storage{database: db}
}

// RunInTransaction encapsulates the function that needs to run in a transaction.
func (s *Storage) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	sess, err := s.database.Client().StartSession(
		// writeconcern is WMajority by default
		options.Session().SetDefaultReadConcern(readconcern.Majority()),
		// read preference in a transaction must be primary
		options.Session().SetDefaultReadPreference(readpref.Primary()),
	)
	if err != nil {
		return err
	}
	defer sess.EndSession(ctx)

	err = mongo.WithSession(ctx, sess, func(sessCtx mongo.SessionContext) error {
		if err = sess.StartTransaction(); err != nil {
			return err
		}

		if err = fn(sessCtx); err != nil {
			return err
		}

		return sess.CommitTransaction(sessCtx)
	})
	if err != nil {
		// abort fails if either the transaction was committed or already aborted (according to docs)
		if abortErr := sess.AbortTransaction(ctx); abortErr != nil {
			return fmt.Errorf("%w %v", abortErr, err)
		}

		return err
	}

	return nil
}

// FindOne returns a row into destination.
func (s *Storage) FindOne(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	return s.database.Collection(collection).FindOne(ctx, filter).Decode(dest)
}

// FindAll returns all rows matching filter into destination.
func (s *Storage) FindAll(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	cursor, err := s.database.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}

	return cursor.All(ctx, dest)
}

// FindMany returns rows into destination.
func (s *Storage) FindMany(
	ctx context.Context,
	collection string,
	filter interface{},
	limit, offset uint64,
	sort string,
	dest interface{},
) (total uint64, err error) {
	count, err := s.database.Collection(collection).CountDocuments(ctx, filter)
	if err != nil {
		return uint64(count), err
	}

	findOptions := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	if sort != "" {
		sortKey := sort
		sortValue := 1
		if strings.HasPrefix(sort, "-") {
			sortKey = strings.TrimPrefix(sort, "-")
			sortValue = -1
		}
		findOptions.SetSort(bson.D{{Key: sortKey, Value: sortValue}})
	}

	cursor, err := s.database.Collection(collection).Find(ctx, filter, findOptions)
	if err != nil {
		return uint64(count), err
	}

	return uint64(count), cursor.All(ctx, dest)
}

// Insert makes insert into database.
func (s *Storage) Insert(ctx context.Context, collection string, document interface{}) error {
	_, err := s.database.Collection(collection).InsertOne(ctx, document)

	return err
}

// Update updates documents in the database.
func (s *Storage) Update(ctx context.Context, collection string, docID primitive.ObjectID, update interface{}) (modifiedCount int64, err error) {
	result, err := s.database.Collection(collection).UpdateOne(ctx, bson.M{"_id": docID}, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// Upsert updates or inserts document in the database.
func (s *Storage) Upsert(ctx context.Context, collection string, docID interface{}, update interface{}) (upsertedCount int64, err error) {
	result, err := s.database.Collection(collection).UpdateOne(ctx, docID, update, options.Update().SetUpsert(true))
	if err != nil {
		return 0, err
	}

	return result.UpsertedCount, nil
}

// Delete deletes document in the database.
func (s *Storage) Delete(ctx context.Context, collection string, docID primitive.ObjectID) (deletedCount int64, err error) {
	result, err := s.database.Collection(collection).DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// DeleteMany deletes filtered documents in the database.
func (s *Storage) DeleteMany(ctx context.Context, collection string, filter interface{}) (deletedCount int64, err error) {
	result, err := s.database.Collection(collection).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}
