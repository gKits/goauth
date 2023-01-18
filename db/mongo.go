package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
connectToDB returns a database object of the MongoDB using the configured MONGO_URI and MOGNO_DB from the environment
In the case of a connection failure this function returns an error
*/
func connectToDB() (*mongo.Database, error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set connection options
	opt := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	// Connect client
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return nil, err
	}

	return client.Database(os.Getenv("MONGO_DB")), nil
}

/*
Update updates one doc in the given collection found from the given filter
If the update process fails at any point an error is returned
*/
func Update(collection string, filter, update interface{}) error {
	// Connect to database
	database, err := connectToDB()
	if err != nil {
		return err
	}

	// Find and update the doc by _id
	_, err = database.Collection(collection).UpdateOne(context.TODO(), filter, update)
	if err != nil {
	}

	return nil
}

/*
Insert inserts the given object as an doc inside the given collection
If the insertion fails at any point an error is returned
*/
func Insert(collection string, object interface{}) error {
	// Connect to database
	database, err := connectToDB()
	if err != nil {
		return err
	}

	// Insert object into collection
	_, err = database.Collection(collection).InsertOne(context.TODO(), object)
	if err != nil {
		return err
	}

	return nil
}

/*
Delete deletes one doc from the given collection found by the given filter
If the deletion fails at any point an error is returned
*/
func Delete(collection string, filter interface{}) error {
	// Connect to database
	database, err := connectToDB()
	if err != nil {
		return err
	}

	// Delete object by filtering for it in the given collection
	_, err = database.Collection(collection).DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

/*
Find returns a SingleResult object that is found in the given collection through the given filter
If there is no object found fitting to the filter an error is returned
*/
func Find(collection string, filter interface{}) (*mongo.SingleResult, error) {
	// Connect to database
	database, err := connectToDB()
	if err != nil {
		return &mongo.SingleResult{}, err
	}

	// Find object by filtering for it in the given collection
	res := database.Collection(collection).FindOne(context.TODO(), filter)
	if res.Err() != nil {
		return &mongo.SingleResult{}, err
	}

	return res, nil
}
