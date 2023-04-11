// db package
package db

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

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
}

// The MongoDB struct is used to store the MongoDB configuration
type MongoDB struct {
	ID             primitive.ObjectID
	CollectionName string
	DatabaseName   string
	URI            string
	Timeout        float64
	_client        IMongoClient
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
	if _client != nil {
		// Client has already been created. Return it
		return _client, nil
	}
	if m.Timeout == 0 {
		m.Timeout = 30 // default value
	}
	timeoutDuration := time.Duration(m.Timeout) * time.Second
	if m.Timeout == 0 {
		timeoutDuration = 30 * time.Second // default value
	}

	var err error
	var uri string
	// Check the URI
	if len(m.URI) == 0 {
		// Set the uri to a default value
		uri = "mongodb://localhost:27017"
	} else {
		uri = m.URI
	}
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(timeoutDuration).        // Increase connection timeout
		SetServerSelectionTimeout(timeoutDuration) // Increase server selection timeout

	_client, err = mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, NewChuxMongoError(fmt.Sprintf("Did not create mongo client for %s. . Check the inner error for details", m.URI), 1000, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration) // Increase context timeout
	defer cancel()

	err = _client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		return nil, NewChuxMongoError(fmt.Sprintf("Did not connect mongo client for %s. Check the inner error for details.", m.URI), 1001, err)
	}

	return _client, nil
}

// Creates a Mongo Document in the configured Mongo DB
// Example:
//
//	mongoDB := New()
//	err := mongoDB.Create(&MyMongoDocument{
//		FirstName: "John",
//		LastName:  "Doe",
//	})
func (m *MongoDB) Create(doc IMongoDocument) error {

	// Get the collection and insert then document
	collection, err := m.getCollection(doc)
	if err != nil {
		return err
	}
	// Create a context with a timeout of 10 seconds
	if m.Timeout == 0 {
		m.Timeout = 30 // default value
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Timeout)*time.Second)
	defer cancel()

	// Insert the document
	insertResult, err := collection.InsertOne(ctx, doc)
	if err != nil {
		log.Fatal(err)
	}

	// Update the ID field of the document with the inserted ID
	if oid, ok := insertResult.InsertedID.(primitive.ObjectID); ok {
		val := reflect.ValueOf(doc).Elem()
		idField := val.FieldByName("ID")
		if idField.IsValid() && idField.CanSet() {
			idField.Set(reflect.ValueOf(oid))
		}
	}

	fmt.Println("Inserted document with ID:", insertResult.InsertedID)
	return nil
}

// Returns a Mongo Document by its ID from the configured Mongo DB
func (m *MongoDB) GetByID(doc IMongoDocument, id string) (interface{}, error) {
	client, err := m.Connect()
	if err != nil {
		return nil, err
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, NewChuxMongoError("GetByID() Failed to Get ObjectIDFromHex. Check the inner error.", 1003, err)
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, NewChuxMongoError("Document not found.", 1003, err)
		}
		return nil, NewChuxMongoError("GetByID failed. Check the inner error.", 1003, err)
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
	// Check if the number of arguments is even (key-value pairs)
	if len(queries)%2 != 0 {
		return nil, NewChuxMongoError("Query() requires an even number of arguments for key-value pairs.", 1006, nil)
	}

	// Connect to the MongoDB client
	client, err := m.Connect()
	if err != nil {
		return nil, err
	}

	// Get the collection from the specified database and collection names
	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())

	// Initialize the filter bson.M (a map) for MongoDB filtering
	filter := bson.M{}

	// Loop through the provided query parameters, in key-value pairs
	for i := 0; i < len(queries); i += 2 {
		// Cast the key to a string and check if the casting was successful
		key, ok := queries[i].(string)
		if !ok {
			return nil, NewChuxMongoError("Query() expects keys to be of type string.", 1006, nil)
		}
		// Add the key-value pair to the filter map
		filter[key] = queries[i+1]
	}

	// Execute the Find operation on the collection with the filter
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, NewChuxMongoError("Query() Failed to find documents. Check the inner error.", 1006, err)
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
			return nil, NewChuxMongoError("Query() Failed to decode document. Check the inner error.", 1006, err)
		}

		// Append the new document to the docs slice
		docs = append(docs, newDoc)
	}

	// Check for errors in the cursor
	if err := cursor.Err(); err != nil {
		return nil, NewChuxMongoError("Query() Cursor error. Check the inner error.", 1006, err)
	}

	// If no documents were found, return an error
	if len(docs) == 0 {
		return nil, NewChuxMongoError("No documents found.", 1006, nil)
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
	client, err := m.Connect()
	if err != nil {
		return nil, err
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, NewChuxMongoError("GetAll() Failed to find documents. Check the inner error.", 1004, err)
	}

	defer cursor.Close(context.Background())

	var docs []IMongoDocument
	for cursor.Next(context.Background()) {
		newDoc := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(IMongoDocument)
		err := cursor.Decode(newDoc)
		if err != nil {
			return nil, NewChuxMongoError("GetAll() Failed to decode document. Check the inner error.", 1004, err)
		}
		docs = append(docs, newDoc)
	}

	if err := cursor.Err(); err != nil {
		return nil, NewChuxMongoError("GetAll() Cursor error. Check the inner error.", 1004, err)
	}

	if len(docs) == 0 {
		return nil, NewChuxMongoError("No documents found.", 1004, nil)
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
	client, err := m.Connect()

	if err != nil {
		return err
	}

	collection := client.Database(doc.GetDatabaseName()).Collection(doc.GetCollectionName())
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
		return NewChuxMongoError("Update() Failed to Get ObjectIDFromHex. Check the inner error.", 1004, err)
	}
	update := bson.M{
		"$set": doc,
	}
	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	if err != nil {
		log.Fatal(err)
		return NewChuxMongoError("Update() Failed to Update. Check the inner error.", 1004, err)
	}
	fmt.Println("Updated ", result.ModifiedCount, " Document(s)")

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
	collection, err := m.getCollection(doc)
	if err != nil {
		return err
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatal(err)
		return NewChuxMongoError("Delete() Failed to Get ObjectIDFromHex. Check the inner error.", 1005, err)
	}
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		log.Fatal(err)
		return NewChuxMongoError("Delete() Failed to Delete. Check the inner error.", 1005, err)
	}
	fmt.Println("Deleted ", result.DeletedCount, " Document(s)")

	return nil
}

// Returns a collection and db name from the IMongoDocument interface
// or the configured values if the interface is not implemented
func (m *MongoDB) getDBAndCollectionName(doc IMongoDocument) (string, string, error) {
	// Using the IMongoDocument interface, get the collection and database name. When the interface is implemented
	// the struct that implements it will have the option of overriding the configured collection and database name
	// with their implementation of the interface methods. This allows for a single struct to be used for multiple collections
	// Its a point of extensibility that is not needed by all use cases.

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
		return "", "", NewChuxMongoError("Did not create document. . Check the inner error for details", 1000, nil)
	}

	return collectionName, dbName, nil
}

// Returns the MongoDB collection from the IMongoDocument interface
func (m *MongoDB) getCollection(doc IMongoDocument) (*mongo.Collection, error) {
	client, err := m.Connect()
	if err != nil {
		return nil, err
	}
	collectionName, dbName, err := m.getDBAndCollectionName(doc)
	if err != nil {
		return nil, err
	}
	collection := client.Database(dbName).Collection(collectionName)
	if collection == nil {
		return nil, NewChuxMongoError(fmt.Sprintf("Unable to get the collection: %s from database: %s Check the inner error for details", collectionName, dbName), 1000, nil)
	}
	return collection, nil
}
