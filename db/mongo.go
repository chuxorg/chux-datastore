// db package
package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/chuxorg/chux-datastore/errors"
	"github.com/chuxorg/chux-datastore/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// The IMongoDocument interface is used to define the methods that must be implemented by a struct that is used to store
// a Mongo Document. The interface is used to allow for a single struct to be used for multiple collections. The struct
// can implement the interface and override the default collection and database name with its own implementation of the
// interface methods.
// Example:
//
//		type MyMongoDocument struct {
//				ID        primitive.ObjectID `bson:"_id,omitempty"`
//				FirstName string             `bson:"firstName"`
//				LastName  string             `bson:"lastName"`
//		}
//
//		func (m *MyMongoDocument) GetCollectionName() string {
//			return "myCollection"
//		}
//	 func (m *MyMongoDocument) GetDatabaseName() string {
//			return "myDatabase"
//		}
//	 func (m *MyMongoDocument) GetURI() string {
//			Use this if you want to override the default URI
//			return "mongodb://localhost:27017"
//		}

//go:generate mockery --name MongoDocument
type IMongoDocument interface {
	GetCollectionName() string
	GetDatabaseName() string
	GetURI() string
	GetID() primitive.ObjectID
	SetID(id primitive.ObjectID)
}

type IMongoClientMethods interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Database(name string) *mongo.Database
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

//go:generate mockery --name MongoClient
type IMongoClient interface {
	IMongoClientMethods
}

//go:generate mockery --name MongoDB
type IMongoDB interface {
	Connect() (*mongo.Client, error)
	Create(doc IMongoDocument, id string) error
	Delete(doc IMongoDocument) error
	GetByID(doc IMongoDocument, id string) error
	GetAll(doc IMongoDocument) ([]IMongoDocument, error)
	Update(doc IMongoDocument, id string) error
	CreateIndicies(fieldNames ...interface{}) (bool, error)
}

// The MongoDB struct is used to store the MongoDB configuration
type MongoDB struct {
	ID             primitive.ObjectID
	CollectionName string
	DatabaseName   string
	URI            string
	Timeout        float64
	_client        IMongoClient
	Logger         *logging.Logger
}

// The _client variable is used to store the MongoDB client
var _client *mongo.Client

// The New func constructs the MongoDB struct with the given options.
// Example:
//
//	mongoDB := New(
//		WithURI("mongodb://localhost:27017"),
//		WithTimeout(30),
//		WithDatabaseName("test"),
//		WithCollectionName("test"),
//	)
func New(options ...func(*MongoDB)) *MongoDB {

	mdb := &MongoDB{}
	for _, o := range options {
		o(mdb)
	}

	return mdb
}

func WithLogger(logger logging.Logger) func(*MongoDB) {

	return func(s *MongoDB) {
		s.Logger = &logger
	}
}

// WithURI is a functional option that sets the MongoDB URI.
//
// Example:
//
//	mongoDB := New(
//		WithURI("mongodb://localhost:27017"),
//	)
func WithURI(uri string) func(*MongoDB) {

	return func(s *MongoDB) {
		s.URI = uri
	}
}

// WithTimeout is a functional option that sets the MongoDB Timeout.
//
// Example:
//
//	mongoDB := New(
//		withTimeout(30),
//	)
func WithTimeout(timeout float64) func(*MongoDB) {

	return func(s *MongoDB) {
		if timeout == 0 {
			timeout = 30
		}
		s.Timeout = timeout
	}
}

// WithDatabaseName is a functional option that sets the MongoDB Database Name.
// Example:
//
//	mongoDB := New(
//		withDatabaseName("test"),
//	)
func WithDatabaseName(databaseName string) func(*MongoDB) {

	return func(s *MongoDB) {
		s.DatabaseName = databaseName
	}
}

// WithCollectionName is a functional option that sets the MongoDB Collection Name.
// Example:
//
//	mongoDB := New(
//		withCollectionName("test"),
//	)
func WithCollectionName(collectionName string) func(*MongoDB) {

	return func(s *MongoDB) {
		s.CollectionName = collectionName
	}
}

// The GetID() method is used to return the ID of the MongoDB struct.
func (m *MongoDB) GetID() primitive.ObjectID {
	m.Logger.Debug("MongoDB.GetID() Getting MongoDB ID '%s'", m.ID)
	return m.ID
}

// The Connect func is used to connect to the MongoDB server. It returns a mongo.Client and an error.
// Example:
//
//	mongoDB, err := Connect()
//	if err != nil {
//		return err
//	}
func (m *MongoDB) Connect() (*mongo.Client, error) {
	logging := m.Logger
	logging.Debug("MongoDB.Connect() Connecting to MongoDB")
	if _client != nil {
		// Client has already been created. Return it
		logging.Debug("MongoDB.Connect() Client has been created, returning _client")
		return _client, nil
	}

	timeoutDuration := time.Duration(m.Timeout) * time.Second
	if m.Timeout == 0 {
		logging.Debug("MongoDB.Connect() Timeout is not set, using default value of 30")
		m.Timeout = 30
		timeoutDuration = 30 * time.Second // default value
	}

	var err error
	var uri string
	// Check the URI
	if len(m.URI) == 0 {
		// Set the uri to a default value
		logging.Debug("MongoDB.Connect() URI is not set. Using default mongodb://localhost:27017")
		uri = "mongodb://localhost:27017"
	} else {
		masked := fmt.Sprintf(m.URI, "*****", "*****")
		logging.Debug("MongoDB.Connect() URI is set. Using: '%s'", masked)
		// Set the uri to the value passed in
		uri = m.URI
	}

	logging.Debug("MongoDB.Connect() Setting client options")
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(timeoutDuration).        // Increase connection timeout
		SetServerSelectionTimeout(timeoutDuration) // Increase server selection timeout

	_client, err = mongo.NewClient(clientOptions)
	if err != nil {
		uri := fmt.Sprintf(m.URI, "*****", "*****")
		msg := fmt.Sprintf("MongoDB.Connect() Did not create mongo client for %s. Check the inner error for details", uri)
		logging.Error(msg)
		return nil, errors.NewChuxDataStoreError(msg, 1000, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration) // Increase context timeout
	defer cancel()

	err = _client.Connect(ctx)
	if err != nil {
		uri := fmt.Sprintf(m.URI, "*****", "*****")
		msg := fmt.Sprintf("MongoDB.Connect() Did not connect to mongo client %s. Check the inner error for details", uri)
		logging.Error(msg, err)
		return nil, errors.NewChuxDataStoreError(msg, 1001, err)
	}

	return _client, nil
}

// Creates a Mongo Document in the configured Mongo DB if the document does not exist.
// Updates a Mongo Document if the configured Mongo DB if the document exists.
// Example:
//
//	type MyMongoDocument struct {
//			ID        primitive.ObjectID `bson:"_id,omitempty"`
//			FirstName string             `bson:"first_name,omitempty"`
//			LastName  string             `bson:"last_name,omitempty"`
//		}
//	 .. IMongoDocument interface methods
//		func (m *MyMongoDocument) GetID() primitive.ObjectID {
//			return m.ID
//		}
//		func (m *MyMongoDocument) SetID(id primitive.ObjectID) {
//			m.ID = id
//		}
//
//		...
//		mongoDB := New()
//		err := mongoDB.Upsert(&MyMongoDocument{
//			FirstName: "John",
//			LastName:  "Doe",
//		})
// Add the 'fields' variadic parameter
func (m *MongoDB) Upsert(doc IMongoDocument, filterFields ...string) error {

	logging := m.Logger

	// Get the collection and insert the document
	collection, err := m.getCollection(doc)
	if err != nil {
		msg := "MongoDB.Connect() Did not get mongo collection. Check the inner error for details."
		logging.Error(msg, err)
		return errors.NewChuxDataStoreError(msg, 1000, err)
	}

	// Create a context with a timeout of 30 seconds by default
	if m.Timeout == 0 {
		m.Timeout = 30 // default value
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Timeout)*time.Second)
	logging.Debug("MongoDB.Upsert() Upserting document with timeout of %d seconds", m.Timeout)
	defer cancel()

	// Get the document ID
	id := doc.GetID()

	// Build the filter using the provided fields
	filter := bson.M{}
	if len(filterFields) == 0 {
		// If no fields are provided, use the default "_id" field
		filter["_id"] = id
	} else {
		// Use the provided fields to build the filter
		for _, field := range filterFields {
			fieldValue, err := m.GetFieldValue(doc, field)
			if err != nil {
				msg := fmt.Sprintf("MongoDB.Upsert() Error getting field value for field '%s': %s", field, err)
				logging.Error(msg, err)
			
				//return errors.NewChuxDataStoreError(msg, 1003, err)
			}
			filter[field] = fieldValue
		}
	}

	// Check if document exists
	var result bson.M
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil && err != mongo.ErrNoDocuments {
		msg := fmt.Sprintf("MongoDB.Upsert() Error checking if document exists: %s", err)
		logging.Error(msg, err)
		return errors.NewChuxDataStoreError(msg, 1004, err)
	}

	// If the document doesn't exist in collection and doesn't have an ID, create a new ObjectID and set it as the document's ID
	if err == mongo.ErrNoDocuments && id == primitive.NilObjectID {
		id = primitive.NewObjectID()
		doc.SetID(id)
	}

	// Upsert operation
	_, err = collection.UpdateOne(
		ctx,
		filter,
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		msg := fmt.Sprintf("MongoDB.Upsert() Error upserting document: %s", err)
		logging.Error(msg, err)
		return errors.NewChuxDataStoreError(msg, 1005, err)
	}

	return nil
}

// Returns a Mongo Document by its ID from the configured Mongo DB
func (m *MongoDB) GetByID(doc IMongoDocument, id string) (interface{}, error) {
	logging := m.Logger
	logging.Debug("MongoDB.GetByID() Connecting to Mongo")

	client, err := m.Connect()
	if err != nil {
		msg := fmt.Sprintf("MongoDB.GetByID() An error occurred connection to Mongo '%s'", err)
		logging.Error(msg, err)
		return nil, errors.NewChuxDataStoreError(msg, 1003, err)
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())
	logging.Debug("MongoDB.GetByID() Getting document with ID '%s' from Database '%s' in Collection '%s'", id, doc.GetDatabaseName(), doc.GetCollectionName())
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		msg := fmt.Sprintf("MongoDB.GetByID() Failed to Get ObjectIDFromHex '%s'", err)
		logging.Error(msg, err)
		return nil, errors.NewChuxDataStoreError(msg, 1003, err)
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Error("MongoDB.GetByID() Document not found '%s'", err)
			return nil, errors.NewChuxDataStoreError("Document not found.", 1003, err)
		}
		logging.Error("MongoDB.GetByID() Failed to FindOne '%s'", err)
		return nil, errors.NewChuxDataStoreError("GetByID failed. Check the inner error.", 1003, err)
	}
	return doc, nil
}

// Query allows for variadic parameters to query any field with any value of any type from MongoDB
// Example:
//
//	mongoDB := New()
//	docs, err := mongoDB.Query(&MyMongoDocument{}, "firstName", "John", "lastName", "Doe")
//	if err != nil {
//		return err
//	}
//	for _, doc := range docs {
//		fmt.Println(doc)
//	}
func (m *MongoDB) Query(doc IMongoDocument, queries ...interface{}) ([]IMongoDocument, error) {
	logging := m.Logger
	logging.Debug("MongoDB.Query() Connecting to Mongo")

	// prepare an empty slice to return in case there are no results
	emptySlice := make([]IMongoDocument, 0)
	// Check if the number of arguments is even (key-value pairs)
	if len(queries)%2 != 0 {
		logging.Error("MongoDB.Query() requires an even number of arguments for key-value pairs.")
		return nil, errors.NewChuxDataStoreError("Query() requires an even number of arguments for key-value pairs.", 1006, nil)
	}

	// Connect to the MongoDB client
	client, err := m.Connect()
	if err != nil {
		logging.Error("MongoDB.Query() error occurred connecting to Mongo '%s'", err)
		return nil, errors.NewChuxDataStoreError("Query() error occurred connecting to Mongo", 1006, err)
	}

	// Get the collection from the specified database and collection names
	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())
	logging.Info("MongoDB.Query() Getting documents from Database '%s' in Collection '%s'", doc.GetDatabaseName(), doc.GetCollectionName())
	// Initialize the filter bson.M (a map) for MongoDB filtering
	filter := bson.M{}

	// Loop through the provided query parameters, in key-value pairs
	for i := 0; i < len(queries); i += 2 {
		// Cast the key to a string and check if the casting was successful
		key, ok := queries[i].(string)
		if !ok {
			logging.Error("MongoDB.Query() expects keys to be of type string.")
			return nil, errors.NewChuxDataStoreError("Query() expects keys to be of type string.", 1006, nil)
		}
		// Add the key-value pair to the filter map
		filter[key] = queries[i+1]
	}

	// Execute the Find operation on the collection with the filter
	cursor, err := collection.Find(context.Background(), filter)
	logging.Info("MongoDB.Query() Executing Find in Collection with filters '%s'", filter)
	if err != nil {
		// The query returned no results
		logging.Info("MongoDB.Query() No documents found '%s'", err)
		return emptySlice, nil
	}

	// Close the cursor when the function is done
	defer cursor.Close(context.Background())

	// Initialize a slice to store the decoded documents
	var docs []IMongoDocument

	// Loop through the cursor while there are more documents to fetch
	for cursor.Next(context.Background()) {
		// Create a new document instance based on the type of the provided doc
		newDoc := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(IMongoDocument)

		// Decode the document from the cursor into the new document instance
		err := cursor.Decode(newDoc)
		if err != nil {
			logging.Error("MongoDB.Query() Failed to decode document '%s'", err)
			return nil, errors.NewChuxDataStoreError("Query() Failed to decode document. Check the inner error.", 1006, err)
		}

		// Append the new document to the docs slice
		docs = append(docs, newDoc)
	}

	// Check for errors in the cursor
	if err := cursor.Err(); err != nil {
		logging.Error("MongoDB.Query() Cursor error '%s'", err)
		return nil, errors.NewChuxDataStoreError("Query() Cursor error. Check the inner error.", 1006, err)
	}

	// If no documents were found, return an error
	if len(docs) == 0 {
		return emptySlice, nil
	}

	// Return the documents slice and a nil error
	return docs, nil
}

// Returns all Mongo Documents from the configured Mongo DB
// Example:
//
//	mongoDB := New()
//	docs, err := mongoDB.GetAll(&MyMongoDocument{})
//	if err != nil {
//		return err
//	}
//	for _, doc := range docs {
//		fmt.Println(doc)
//	}
func (m *MongoDB) GetAll(doc IMongoDocument) ([]IMongoDocument, error) {
	logging := m.Logger
	logging.Debug("MongoDB.GetAll() Connecting to Mongo")
	client, err := m.Connect()
	if err != nil {
		logging.Error("MongoDB.GetAll() error occurred connecting to Mongo '%s'", err)
		return nil, errors.NewChuxDataStoreError("MongoDB.GetAll() error occurred connecting to Mongo", 1004, err)
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		logging.Error("MongoDB.GetAll() Failed to find documents '%s'", err)
		return nil, errors.NewChuxDataStoreError("MongoDB.GetAll() Failed to find documents. Check the inner error.", 1004, err)
	}

	defer cursor.Close(context.Background())

	var docs []IMongoDocument
	for cursor.Next(context.Background()) {
		newDoc := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(IMongoDocument)
		err := cursor.Decode(newDoc)
		if err != nil {
			logging.Error("MongoDB.GetAll() Failed to decode document '%s'", err)
			return nil, errors.NewChuxDataStoreError("MongoDB.GetAll() Failed to decode document. Check the inner error.", 1004, err)
		}
		docs = append(docs, newDoc)
	}

	if err := cursor.Err(); err != nil {
		logging.Error("MongoDB.GetAll() Cursor error '%s'", err)
		return nil, errors.NewChuxDataStoreError("MongoDB.GetAll() Cursor error. Check the inner error.", 1004, err)
	}

	if len(docs) == 0 {
		logging.Info("MongoDB.GetAll() No documents found.")
	}

	return docs, nil
}

// Updates a Mongo Document by its ID from the configured Mongo DB
// Example:
//
//	mongoDB := New(
//		WithDatabaseName("mydb"),
//		WithCollectionName("mycollection"),
//		WithURI("mongodb://localhost:27017"),
//	)
//	err := mongoDB.Update(&MyMongoDocument{
//		FirstName: "John",
//		LastName:  "Doe",
//	}, "5e9b9b9b9b9b9b9b9b9b9b9b")
func (m *MongoDB) Update(doc IMongoDocument, id string) error {
	logging := m.Logger
	logging.Debug("MongoDB.Update() Connecting to Mongo")

	client, err := m.Connect()

	if err != nil {
		logging.Error("MongoDB.Update() error occurred connecting to Mongo '%s'", err)
		return errors.NewChuxDataStoreError("MongoDB.Update() error occurred connecting to Mongo", 1004, err)
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("MongoDB.Update() Failed to Get ObjectIDFromHex '%s'", err)
		return errors.NewChuxDataStoreError("MongoDB.Update() Failed to Get ObjectIDFromHex. Check the inner error.", 1004, err)
	}
	update := bson.M{
		"$set": doc,
	}
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		logging.Error("MongoDB.Update() Failed to Update '%s'", err)
		return errors.NewChuxDataStoreError("MongoDB.Update() Failed to Update. Check the inner error.", 1004, err)
	}
	logging.Info("MongoDB.Update() Updated ", result.ModifiedCount, " Document(s)")

	return nil
}

// Deletes a Mongo Document by its ID from the configured Mongo DB
// Example:
//
//	mongoDB := New(
//		WithDatabaseName("mydb"),
//		WithCollectionName("mycollection"),
//		WithURI("mongodb://localhost:27017"),
//	)
//	err := mongoDB.Delete(&MyMongoDocument{
//		FirstName: "John",
//		LastName:  "Doe",
//	}, "5e9b9b9b9b9b9b9b9b9b9b9b")
func (m *MongoDB) Delete(doc IMongoDocument, id string) error {
	logging := m.Logger
	logging.Debug("MongoDB.Delete() Connecting to Mongo")

	collection, err := m.getCollection(doc)

	if err != nil {
		logging.Error("MongoDB.Delete() error occurred connecting to Mongo '%s'", err)
		return errors.NewChuxDataStoreError("MongoDB.Delete() error occurred connecting to Mongo", 1004, err)
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("MongoDB.Delete() Failed to Get ObjectIDFromHex.", err)
		return errors.NewChuxDataStoreError("MongoDB.Delete() Failed to Get ObjectIDFromHex. Check the inner error.", 1005, err)
	}
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		msg := "MongoDB.Delete() did not delete ObjecID: v% from collection: %v"
		logging.Error(msg, err)
		return errors.NewChuxDataStoreError("MongoDB.Delete() Failed to Delete. Check the inner error.", 1005, err)
	}
	fmt.Println("Deleted ", result.DeletedCount, " Document(s)")

	return nil
}

// Returns a collection and db name from the IMongoDocument interface
// or the configured values if the interface is not implemented
func (m *MongoDB) getDBAndCollectionName(doc IMongoDocument) (string, string, error) {
	logging := m.Logger
	// Using the IMongoDocument interface, get the collection and database name. When the interface is implemented
	// the struct that implements it will have the option of overriding the configured collection and database name
	// with their implementation of the interface methods. This allows for a single struct to be used for multiple collections
	// Its a point of extensibility that is not needed by all use cases.
	logging.Debug("MongoDB.getDBAndCollectionName() Getting DB and Collection Name")
	var dbName string
	var collectionName string
	if len(doc.GetCollectionName()) > 0 {
		collectionName = doc.GetCollectionName()
	} else {
		collectionName = m.CollectionName
	}
	if len(doc.GetDatabaseName()) > 0 {
		dbName = doc.GetDatabaseName()
	} else {
		dbName = m.DatabaseName
	}

	if len(collectionName) == 0 || len(dbName) == 0 {
		logging.Warning("MongoDB.getDBAndCollectionName() Either no collection '%s' or database '%s' was found.", collectionName, dbName)
		return "", "", nil
	}

	return collectionName, dbName, nil
}

// Returns the MongoDB collection from the IMongoDocument interface
func (m *MongoDB) getCollection(doc IMongoDocument) (*mongo.Collection, error) {
	logging := m.Logger
	logging.Debug("MongoDB.getCollection() Connecting to Mongo")
	client, err := m.Connect()
	if err != nil {
		logging.Error("MongoDB.getCollection() error occurred connecting to Mongo '%s'", err)
		return nil, errors.NewChuxDataStoreError("MongoDB.getCollection() error occurred connecting to Mongo", 1004, err)
	}
	collectionName, dbName, err := m.getDBAndCollectionName(doc)
	if err != nil {
		logging.Error("MongoDB.getCollection() error occurred getting the collection name and database name from the IMongoDocument interface '%s'", err)
		return nil, errors.NewChuxDataStoreError("MongoDB.getCollection() error occurred getting the collection name and database name from the IMongoDocument interface", 1004, err)
	}
	collection := client.Database(dbName).Collection(collectionName)
	if collection == nil {
		logging.Error("MongoDB.getCollection() Unable to get the collection: %s from database: %s", collectionName, dbName)
		return nil, errors.NewChuxDataStoreError(fmt.Sprintf("Unable to get the collection: %s from database: %s Check the inner error for details", collectionName, dbName), 1000, nil)
	}
	return collection, nil
}

func (m *MongoDB) CreateIndices(doc IMongoDocument, fieldNames ...string) (bool, error) {
	logging := m.Logger
	logging.Debug("MongoDB.CreateIndices() Connecting to Mongo")

	client, err := m.Connect()
	if err != nil {
		logging.Error("MongoDB.CreateIndices() error occurred connecting to Mongo '%s'", err)
		return false, errors.NewChuxDataStoreError("MongoDB.CreateIndices() error occurred connecting to Mongo", 1004, err)
	}
	collectionName, dbName, err := m.getDBAndCollectionName(doc)
	if err != nil {
		logging.Error("MongoDB.CreateIndices() error occurred getting the collection name and database name from the IMongoDocument interface '%s'", err)
		return false, errors.NewChuxDataStoreError("Unable to get the collection name and database name from the IMongoDocument interface. Check the inner error for details", 1004, err)
	}
	collection := client.Database(dbName).Collection(collectionName)
	if collection == nil {
		logging.Error("MongoDB.CreateIndices() Unable to get the collection: %s from database: %s", collectionName, dbName)
		return false, errors.NewChuxDataStoreError(fmt.Sprintf("Unable to get the collection: %s from database: %s Check the inner error for details", collectionName, dbName), 1000, nil)
	}
	for _, fieldName := range fieldNames {
		indexView := collection.Indexes()
		indexModel := mongo.IndexModel{
			Keys: bson.M{
				fieldName: 1,
			},
			Options: options.Index().SetUnique(true),
		}
		_, err := indexView.CreateOne(context.Background(), indexModel)
		if err != nil {
			logging.Error("MongoDB.CreateIndices() Unable to create the indicies: %s on collection: %s", fieldNames, collectionName)
			return false, errors.NewChuxDataStoreError(fmt.Sprintf("Unable to create the indicies: %s on collection: %s Check the inner error for details", fieldNames, collectionName), 1000, nil)
		}
	}
	return true, nil
}

// Returns the value of a field in a document using reflection
// GetFieldValue method receives a field name and returns its value.
func (m *MongoDB) GetFieldValue(doc IMongoDocument, field string) (interface{}, error) {
    // Get the value of the document structure
    val := reflect.ValueOf(doc)

    // Check if the document value is a pointer, and if so, get the underlying value
    if val.Kind() == reflect.Ptr {
        val = val.Elem()
    }

    // Get the type of the document structure
    typ := val.Type()

    // Find the struct field that has a bson tag matching the provided field name
    for i := 0; i < typ.NumField(); i++ {
        bsonTag := typ.Field(i).Tag.Get("bson")
        if strings.Split(bsonTag, ",")[0] == field {
            // Get the field value
            fieldValue := val.Field(i)
            if !fieldValue.IsValid() {
                return nil, fmt.Errorf("unknown field: %s", field)
            }

            // Return the field value as an interface
            return fieldValue.Interface(), nil
        }
    }

    // Return an error if no matching field is found
    return nil, fmt.Errorf("unknown field: %s", field)
}
