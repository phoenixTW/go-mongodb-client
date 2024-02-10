package mock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockedStorageReader is a mock for StorageReader interface
type MockedStorageReader struct {
	FindMock     func(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error)
	FindAllMock  func(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error)
	FindManyMock func(
		ctx context.Context,
		collection string,
		filter interface{},
		limit, offset uint64,
		sort string,
		dest interface{},
	) (total uint64, err error)
}

// FindOne returns a row into destination.
func (mock *MockedStorageReader) FindOne(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	return mock.FindMock(ctx, collection, filter, dest)
}

// FindAll returns rows into destination.
func (mock *MockedStorageReader) FindAll(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
	return mock.FindAllMock(ctx, collection, filter, dest)
}

// FindMany returns rows into destination.
func (mock *MockedStorageReader) FindMany(ctx context.Context, collection string, filter interface{}, limit, offset uint64, sort string, dest interface{}) (total uint64, err error) {
	return mock.FindManyMock(ctx, collection, filter, limit, offset, sort, dest)
}

// NewStorageReaderStub will return a stub for StorageReader that will return given result
func NewStorageReaderStub(t *testing.T, result string) *MockedStorageReader {
	return &MockedStorageReader{FindAllMock: func(ctx context.Context, collection string, filter interface{}, dest interface{}) (err error) {
		assert.NoError(t, bson.UnmarshalExtJSON([]byte(result), true, dest))

		return nil
	}}
}

// MockedStorageWriter is a mock for StorageWriter interface
type MockedStorageWriter struct {
	RunInTransactionMock func(ctx context.Context, fn func(context.Context) error) error
	InsertMock           func(ctx context.Context, collection string, document interface{}) error
	UpdateMock           func(ctx context.Context, collection string, docID interface{}, update interface{}) (modifiedCount int64, err error)
	UpsertMock           func(ctx context.Context, collection string, docID interface{}, update interface{}) (upsertedCount int64, err error)
	DeleteMock           func(ctx context.Context, collection string, docID primitive.ObjectID) (deletedCount int64, err error)
	DeleteManyMock       func(ctx context.Context, collection string, filter interface{}) (deletedCount int64, err error)
}

// RunInTransaction encapsulates the function that needs to run in a transaction.
func (mock *MockedStorageWriter) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return mock.RunInTransactionMock(ctx, fn)
}

// Insert makes insert into database.
func (mock *MockedStorageWriter) Insert(ctx context.Context, collection string, document interface{}) error {
	return mock.InsertMock(ctx, collection, document)
}

// Update updates documents in the database.
func (mock *MockedStorageWriter) Update(ctx context.Context, collection string, docID primitive.ObjectID, update interface{}) (modifiedCount int64, err error) {
	return mock.UpdateMock(ctx, collection, docID, update)
}

// Upsert updates or inserts document in the database.
func (mock *MockedStorageWriter) Upsert(ctx context.Context, collection string, docID interface{}, update interface{}) (upsertedCount int64, err error) {
	return mock.UpsertMock(ctx, collection, docID, update)
}

// Delete deletes document in the database.
func (mock *MockedStorageWriter) Delete(ctx context.Context, collection string, docID primitive.ObjectID) (deletedCount int64, err error) {
	return mock.DeleteMock(ctx, collection, docID)
}

// DeleteMany deletes filtered documents in the database.
func (mock *MockedStorageWriter) DeleteMany(ctx context.Context, collection string, filter interface{}) (deletedCount int64, err error) {
	return mock.DeleteManyMock(ctx, collection, filter)
}

// MockedStorageReaderWriter is mock for StorageReaderWriter interface
type MockedStorageReaderWriter struct {
	MockedStorageReader
	MockedStorageWriter
}

// GetDatabaseName returns test database name
func (mock MockedStorageReaderWriter) GetDatabaseName() string {
	return "test-database"
}
