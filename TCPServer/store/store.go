package store

import (
	"errors"
)

var (
	// Errors
	ErrKeyNotFound = errors.New("Unknown key")
	ErrBadData     = errors.New("Bad Data")
)

type StoreMessage struct {
	responseChannel chan interface{} // dynamic channel type for responding to messages
	data            interface{}
}

// Make a little more prettier
func NewStoreMessage(channel chan interface{}, data interface{}) StoreMessage {
	smsg := StoreMessage{
		responseChannel: channel,
		data:            data,
	}
	return smsg
}

func (ds *DataStore) T() bool {
	return true
}

type DataStore struct {
	putChannel    chan StoreMessage
	deleteChannel chan StoreMessage
	getChannel    chan StoreMessage
	doneChannel   chan bool
	data          map[string]string
}

func NewDataStore() *DataStore {

	cache := DataStore{
		putChannel:    make(chan StoreMessage),
		deleteChannel: make(chan StoreMessage),
		getChannel:    make(chan StoreMessage),
		doneChannel:   make(chan bool),
		data:          make(map[string]string),
	}

	go cache.monitor()
	return &cache
}

func (ds *DataStore) monitor() {
	process := true

	for process {
		select {
		case msg := <-ds.putChannel:
			defer close(msg.responseChannel)
			msg.responseChannel <- ds.put(msg.data)
		case msg := <-ds.deleteChannel:
			defer close(msg.responseChannel)
			msg.responseChannel <- ds.delete(msg.data)
		case msg := <-ds.getChannel:
			defer close(msg.responseChannel)
			msg.responseChannel <- ds.get(msg.data)
		case <-ds.doneChannel:
			process = false
		}
	}
}

func (ds *DataStore) Put(msg StoreMessage) {
	ds.putChannel <- msg
}

func (ds *DataStore) Delete(msg StoreMessage) {
	ds.deleteChannel <- msg
}

func (ds *DataStore) Get(msg StoreMessage) {
	ds.getChannel <- msg
}

func (ds *DataStore) put(data interface{}) error {

	kv, ok := data.([]string)
	if !ok {
		return ErrBadData
	}

	key := kv[0]
	value := kv[1]

	ds.data[key] = value
	return nil
}

type GetContents struct {
	Value string
	Err   error
}

func (ds *DataStore) get(data interface{}) GetContents {

	key, ok := data.(string)
	if !ok {
		return GetContents{Value: "", Err: ErrBadData}
	}

	value, ok := ds.data[key]
	if !ok {
		return GetContents{Value: "", Err: ErrKeyNotFound}
	}

	return GetContents{Value: value, Err: nil}
}

func (ds *DataStore) delete(data interface{}) error {

	key, ok := data.(string)
	if !ok {
		return ErrBadData
	}

	_, contains := ds.data[key]

	if !contains {
		return ErrKeyNotFound
	}

	delete(ds.data, key)
	return nil
}
