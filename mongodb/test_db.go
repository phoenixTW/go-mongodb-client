package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/phoenixTW/go-mongodb-client/mongostorage"
	"os"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDBSuite defines a suite that can be embedded into other test suites. This provides out of the box
// capability to plugin DB for test cases. The lifecycle of the DB client is governed by the engulfing suite itself.
type TestDBSuite struct {
	suite.Suite
	DSN    string
	DBName string
	TestDB
}

// TestDB defines db client and data access layers.
type TestDB struct {
	MongoClient *mongo.Client
	Database    mongostorage.StorageReaderWriter
}

// GetMongoDSN returns DSN to connect to MongoDB
func GetMongoDSN() string {
	mongoDSN, exists := os.LookupEnv("MONGO_DSN")
	if !exists {
		mongoDSN = "mongodb://localhost:27017"
	}

	return mongoDSN
}

// NewTestDBSuite creates new test suite for tests dependent on a database
func NewTestDBSuite(database string) TestDBSuite {
	mongoDSN := GetMongoDSN()

	return TestDBSuite{
		DSN:    mongoDSN,
		DBName: database,
	}
}

// SetupSuite sets up the test db suite.
func (t *TestDBSuite) SetupSuite() {
	testDB, err := NewTestDatabase(t.DSN, t.DBName)
	if err != nil {
		t.FailNow(fmt.Sprintf("failed to setup test db %v", err))
	}

	t.TestDB = testDB
}

// TruncateCollection will remove all documents from a given collection
func (t *TestDBSuite) TruncateCollection(collection string) {
	t.TestDB.TruncateCollection(collection)
}

// DropCollection will drop the collection
func (t *TestDBSuite) DropCollection(collection string) {
	db := t.MongoClient.Database(t.DBName)
	t.NoError(db.Collection(collection).Drop(context.Background()))
	return
}

func NewTestDatabase(dsn, dbName string) (TestDB, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn).SetDirect(true))
	if err != nil {
		return TestDB{}, err
	}

	return TestDB{
		MongoClient: client,
		Database:    mongostorage.New(client.Database(dbName)),
	}, nil
}

// TruncateCollection will remove all documents from a given collection
func (t *TestDB) TruncateCollection(collection string) {
	// nolint: errcheck // reason: here we don't care as it's part of the tests
	_, _ = t.Database.DeleteMany(context.Background(), collection, bson.M{})
}

func loadSchema(filename string) (bson.M, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var schema bson.M
	if err := json.Unmarshal(fileBytes, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (t *TestDBSuite) EnforceCollectionSchema(collectionName string, schemaPath string) error {
	db := t.MongoClient.Database(t.DBName)
	schema, err := loadSchema(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema: %s", err)
	}

	// Create new collection with schema validation
	opts := options.CreateCollection().SetValidator(schema)
	err = db.CreateCollection(context.Background(), collectionName, opts)
	if err != nil {
		return fmt.Errorf("failed to create collection: %s", err)
	}

	return nil
}
