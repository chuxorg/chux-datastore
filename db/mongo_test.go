package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew tests the New function of the MongoDB struct
func TestNew(t *testing.T) {
	uri := "mongodb://localhost:27017"
	timeout := 30.0
	databaseName := "test"
	collectionName := "test"

	mongoDB := New(
		WithURI(uri),
		WithTimeout(timeout),
		WithDatabaseName(databaseName),
		WithCollectionName(collectionName),
	)

	assert.Equal(t, uri, mongoDB.URI)
	assert.Equal(t, timeout, mongoDB.Timeout)
	assert.Equal(t, databaseName, mongoDB.DatabaseName)
	assert.Equal(t, collectionName, mongoDB.CollectionName)
}

