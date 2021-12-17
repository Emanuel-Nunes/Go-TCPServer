package store_test

import (
	"store"
	"testing"
)


func TestAddEntry(t *testing.T) {

	t.Run("AddNewSuccessful", func(t *testing.T) {
		dataStore := store.NewDataStore() // setup

		testAdd(t, dataStore, []string{"1", "Apple"}, nil) // test
		
		dataStore = nil // cleanup
	})

	t.Run("AddUpdateAuthorized", func(t *testing.T) {
		dataStore := store.NewDataStore()
		
		testAdd(t, dataStore, []string{"1", "Apple"}, nil)
		testAdd(t, dataStore, []string{"1", "Banana"}, nil)
		
		dataStore = nil
	})
}

func TestDeleteEntry(t *testing.T) {

	t.Run("DeleteSuccessful", func(t *testing.T) {
		dataStore := store.NewDataStore()
		
		testAdd(t, dataStore, []string{"1", "Apple"}, nil)
		testDelete(t, dataStore, "1", nil)
		
		dataStore = nil
	})

	t.Run("DeleteKeyNotFound", func(t *testing.T) {
		dataStore := store.NewDataStore()
		
		testDelete(t, dataStore, "1", store.ErrKeyNotFound)
		
		dataStore = nil
	})
}

func TestGetEntry(t *testing.T) {

	t.Run("GetKeyExists", func(t *testing.T) {
		expectedResults := store.GetContents{Value: "Apple", Err: nil}
		dataStore := store.NewDataStore()
		
		testAdd(t, dataStore, []string{"1", "Apple"}, nil)
		testGet(t, dataStore, "1", expectedResults)
		
		dataStore = nil
	})

	t.Run("GetKeyNotFound", func(t *testing.T) {
		expectedResults := store.GetContents{Value: "", Err: store.ErrKeyNotFound}
		dataStore := store.NewDataStore()

		testGet(t, dataStore, "1", expectedResults)
		
		dataStore = nil
	})
}



// Helper functions

func testAdd(t *testing.T, dataStore *store.DataStore, data []string, expected error) {

	testChan := make(chan interface{})
	msg := store.NewStoreMessage(testChan, data)
	dataStore.Put(msg)
	result := <-testChan

	if result != expected {
		t.Error("Expected error : ", expected, " Actual error: ", result)
	}
}

func testDelete(t *testing.T, dataStore *store.DataStore, key string, expected error) {
	testChan := make(chan interface{})
	msg := store.NewStoreMessage(testChan, key)
	dataStore.Delete(msg)
	result := <-testChan

	if result != expected {
		t.Error("Expected error: ", expected, " Actual error: ", result)
	}
}

func testGet(t *testing.T, dataStore *store.DataStore, key string, expected store.GetContents) {
	testChan := make(chan interface{})
	msg := store.NewStoreMessage(testChan, key)
	dataStore.Get(msg)
	result := <-testChan

	actualResults, ok := result.(store.GetContents)
	if !ok {
		t.Error("Returned type is not what we expected")
		return
	}

	if actualResults != expected {
		t.Error("Expected results: ", expected, " Actual results: ", actualResults)
	}
}