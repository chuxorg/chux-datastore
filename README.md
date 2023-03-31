[![Build and Test](https://github.com/csailer/chux-datastore/actions/workflows/build_and_test.yml/badge.svg?branch=master)](https://github.com/csailer/chux-datastore/actions/workflows/build_and_test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/csailer/chux-mongo)](https://goreportcard.com/report/github.com/csailer/chux-mongo)
# Chux MongoDB Go Library - chux-mongo

chux-datastore is a simple and easy-to-use Go library for performing basic CRUD operations on MongoDB. The library provides a clean and straightforward interface for connecting to MongoDB, as well as creating, retrieving, updating, and deleting documents.

## Driving Forces

Ah, the treacherous realm of software development! A chaotic wasteland where arbitrary deadlines circle like vultures, ever ready to swoop down upon the unsuspecting developer from the fantastical realms of project plans, conjured by oxygen thieves bearing titles like Project Manager and Scrum Master. Amidst this silly world, one must conquer the ferocious beast of repetitive code. But fret not, for chux-mongo rides to the rescue, charging heroically into the melee!

Envision this: You are a Software Engineer with A.D.D. battling the banal and the humdrum of avoiding the "Crotch Punching Gnomes" that lurk beneath your desk and ponce, without warning, at every unit test failure. One fateful day, your gaze falls upon your jumbled codebase, only to discover that the ominous MongoDB CRUD code is the recurring nemesis in this grand saga. True, it functions, but its unsightly presence is a thorn in your side. If only something could emerge to liberate you from this ordeal...

Enter `chux-mongo`, a formidable Excalibur forged amidst the inferno of ingenuity and practicality. In one swift motion, it beheads the monstrous hydra of boilerplate, emancipating you from its tyrannical clutches. With chux-mongo as your steadfast ally, your codebase transforms into a pristine sanctuary where CRUD operations frolic freely like untamed mustangs.

And what of the days yet to come? Shall this valiant library withstand the sands of time, lending its might to other endeavors in their pursuit of greatness? The answer, dear compatriot, resounds with a thunderous affirmation. For chux-mongo transcends the fleeting present, safeguarding the myriad services that shall voyage through the perilous landscape of reading and writing to and from Mongo for countless eons.

In closing, chux-mongo emerges as the mythical hero we never knew we yearned for, restoring harmony to the pandemonium, severing the chains of monotony, and banishing the specter of boilerplate Mongo code to the annals of distant memory. 

The Repo is Dark and Full of Terrors.

Nah, Seriously. Very simple library that encapsulates CRUD operations for [Mongo DB](https://www.mongodb.com) with Golang >= 1.19.
## Features
- Simple MongoDB connection setup
- Basic CRUD operations: Create, Read, Update, and Delete
- Supports custom MongoDB document structures
## Getting Started

**Prerequisites**
- Go 1.19 or later
- MongoDB server
- Installation

To install chux-mongo, run the following command:

```sh
go get github.com/csailer/chux-datastore
```

## Usage
This example demonstrates how to perform CRUD operations using the provided code with a 
custom struct called `MyMongoDocument` that implements the `IMongoDocument`interface. 
The `main()` function initializes a `MongoDB` instance, and then it creates, retrieves, 
updates, and deletes a document in a `MongoDB` collection.

First, download `chux-mongo`:

```shell
$ go get github.com/csailer/chux-datastore
```

```go
package main

import (
	"fmt"
	"log"

	"github.com/csailer/chux-mongo/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Define a custom struct that implements the IMongoDocument interface
type MyMongoDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	FirstName string             `bson:"firstName"`
	LastName  string             `bson:"lastName"`
}

func (m *MyMongoDocument) GetCollectionName() string {
	return "myCollection"
}

func (m *MyMongoDocument) GetDatabaseName() string {
	return "myDatabase"
}

func (m *MyMongoDocument) GetURI() string {
	return "mongodb://localhost:27017"
}

func (m *MyMongoDocument) GetID() primitive.ObjectID {
	return m.ID
}

func main() {
	// Initialize the MongoDB instance
	mongoDB := db.New(
		db.WithURI("mongodb://localhost:27017"),
		db.WithTimeout(30),
	)

	// Create a new document
	doc := &MyMongoDocument{
		FirstName: "John",
		LastName:  "Doe",
	}
	err := mongoDB.Create(doc)
	if err != nil {
		log.Fatalf("Error creating document: %v", err)
	}
	fmt.Printf("Created document: %+v\n", doc)

	// Get document by ID
	foundDoc := &MyMongoDocument{}
	id := doc.ID.Hex()
	_, err = mongoDB.GetByID(foundDoc, id)
	if err != nil {
		log.Fatalf("Error getting document by ID: %v", err)
	}
	fmt.Printf("Found document by ID: %+v\n", foundDoc)

	// Get all documents
	allDocs, err := mongoDB.GetAll(&MyMongoDocument{})
	if err != nil {
		log.Fatalf("Error getting all documents: %v", err)
	}
	fmt.Println("All documents:")
	for _, d := range allDocs {
		fmt.Printf("%+v\n", d)
	}

	// Update a document
	updatedDoc := &MyMongoDocument{
		ID:        doc.ID,
		FirstName: "Jane",
		LastName:  "Doe",
	}
	err = mongoDB.Update(updatedDoc, id)
	if err != nil {
		log.Fatalf("Error updating document: %v", err)
	}
	fmt.Printf("Updated document: %+v\n", updatedDoc)

	// Delete a document
	err = mongoDB.Delete(updatedDoc, id)
	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
	}
	fmt.Println("Deleted document")
}


```

# Makefile

- `make test` - Runs all tests in `chux-mongo`.
- `make test-release` - Commits, tags, and releases `chux-mongo`  
   &nbsp; 
   To release and version, pass in major and minor values on the command line. If either major or minor has a value, the patch number is set to zero. If neither major nor minor has a value, the patch number is incremented by 1.
   You can run the target with or without the MAJOR_VALUE and MINOR_VALUE variables, like this:
   &nbsp; 
   To bump the patch version:
   ```shell
   make release-version
   ```
   &nbsp; 
   To set a new major version and reset minor and patch to zero:
   ```shell
   MAJOR_VALUE=2 make release-version
   ```
   &nbsp; 
   To set a new minor version and reset patch to zero:
   ```shell
   MINOR_VALUE=3 make release-version 
   ```
   &nbsp; 
   To set both new major and minor versions and reset patch to zero:
   ```shell
   MAJOR_VALUE=2 MINOR_VALUE=3 make release-version 
   ``` 
## Unit Tests
The unit tests use an in-memory MongoDB from Ben Weissmann named [memongo](https://github.com/benweissmann/memongo).
I really like it and, to me, it beats Mocking MongoDB. It works very well locally but I haven't had much luck in GitHub
Pipelines.
## Contributing
Contributions are welcome! Please feel free to submit issues and pull requests.

## Projects Using this Library

- [chux-bizobj](github.com/csailer/chux-mongo)

## License
chux-mongo is released under the [GNU General Public License v3](https://www.gnu.org/licenses/gpl-3.0.en.html)
