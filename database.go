package main

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Init checks if the Database actually works
func (db *Database) Init() {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	index := mgo.Index{
		Key:        []string{"icao24"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.CollectionState).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

// Add removes and adds documents to passes collection name
// Example:
// CollectionState
// CollectionAirport
// CollectionFlight
//
// Add example
// var documents []interface{}
//	 for i := range flights {
//		 documents = append(documents, flights[i])
// 	 }
// err := DBValues.Add(documents, DBValues.CollectionState)
func (db *Database) Add(documents []interface{}, collN string) error {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	_, err = session.DB(db.DatabaseName).C(collN).RemoveAll(nil)
	if err != nil {
		return err
	}

	err = session.DB(db.DatabaseName).C(collN).Insert(documents...)

	return err
}

// Count Counts the documents in a collection
// Example:
// CollectionState
// CollectionAirport
// CollectionFlight
func (db *Database) Count(collN string) int {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(collN).Count()
	if err != nil {
		fmt.Printf("error in Count(): %v", err.Error())
		return -1
	}

	return count
}

// GetFlight accepts bson.M{} to find all flights with choosen paramaters
// Example
// findData == bson.M{"estarrivalairport": "ENFL"}
func (db *Database) GetFlight(findData bson.M) ([]Flight, error) {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	var flights []Flight
	err = session.DB(db.DatabaseName).C(db.CollectionFlight).Find(findData).All(&flights)
	if err != nil {
		return []Flight{}, err
	}

	return flights, errorCheck(flights)
}

// GetState accepts bson.M{} to find all flights with choosen paramaters
// Example
// findData == bson.M{"callsign": "<insert callsign here>"}
func (db *Database) GetState(findData bson.M) ([]State, error) {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	var state []State

	err = session.DB(db.DatabaseName).C(db.CollectionState).Find(findData).All(&state)
	if err != nil {
		return []State{}, err
	}

	return state, errorCheck(state)
}

// GetAirport accepts bson.M{} to find all Airports with choosen paramters
// Example
// FindData == bson.M{"country": "Italy"}
func (db *Database) GetAirport(findData bson.M) ([]Airport, error) {
	session, err := mgo.Dial(db.HostURL)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	var port []Airport

	err = session.DB(db.DatabaseName).C(db.CollectionAirport).Find(findData).All(&port)
	if err != nil {
		return []Airport{}, err
	}

	return port, errorCheck(port)
}

// GetPlanes accepts bson.M{} to find all flights with choosen paramaters
// Example
// findData == bson.M{"origincountry": "Italy"}
func (db *Database) GetPlanes(find bson.M) ([]Planes, error) {
	var s []State
	var f []Flight
	var err error

	if s, err = db.GetState(find); err != nil {
		return []Planes{}, err
	}
	if f, err = db.GetFlight(nil); err != nil {
		planes := mergeStatesAndFlights(s, f)
		return planes, nil
	}

	// Merges states and flight together
	planes := mergeStatesAndFlights(s, f)
	return planes, nil
}

// Checks wether or not result interface is nil
// Or if it contains nil
func errorCheck(result interface{}) error {
	if reflect.ValueOf(result).IsNil() || result == nil {
		return errors.New("Nothing returned from query")
	}
	return nil
}

func (fieldTag *State) getField(field string) (string, error) {
	// Returns the field that matches the given struct json tag
	value := reflect.ValueOf(fieldTag).Elem()
	for i := 0; i < value.NumField(); i++ {
		if value.Type().Field(i).Tag.Get("json") == field {
			return fmt.Sprint(value.Field(i)), nil
		}

	}
	return "", errors.New("")
}

func (fieldTag *Airport) getField(field string) (string, error) {
	// Returns the field that matches the given struct json tag
	value := reflect.ValueOf(fieldTag).Elem()
	for i := 0; i < value.NumField(); i++ {
		if value.Type().Field(i).Tag.Get("json") == field {
			return fmt.Sprint(value.Field(i)), nil
		}

	}
	return "", errors.New("")
}
