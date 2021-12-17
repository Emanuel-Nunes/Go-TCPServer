package dataServer

import (
	"fmt"
	"net"
	"store"
	"testing"
)

func TestGetDigits(t *testing.T) {

	t.Run("getDigitsSuccessful", func(t *testing.T) {
		expected := 5

		actual := getDigits(50000)

		if expected != actual {
			t.Error(fmt.Sprintf("Expected: %d, Actual: %d", expected, actual))
		}
	})

	t.Run("getDigitsZeroPassedIn", func(t *testing.T) {
		expected := 0

		actual := getDigits(0)

		if expected != actual {
			t.Error(fmt.Sprintf("Expected: %d, Actual: %d", expected, actual))
		}
	})

	t.Run("getDigitsNegativePassedIn", func(t *testing.T) {
		expected := -1337

		actual := getDigits(-1)

		if expected != actual {
			t.Error(fmt.Sprintf("Expected: %d, Actual: %d", expected, actual))
		}
	})
}

func TestInitClientListener(t *testing.T) {

	t.Run("InitClientListenerSuccessful", func(t *testing.T) {

		expectedResponse := "ack"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")
		go tcpServer.InitClientListener("localhost:1234")

		conn, err := net.Dial("tcp", "localhost:1234")
		if err != nil {
			t.Error("Failed to connect to server")
		}

		_, _ = conn.Write([]byte("del11k"))
		buffer := make([]byte, 3)
		_, _ = conn.Read(buffer)

		actualResponse := string(buffer)

		if actualResponse != expectedResponse {
			t.Error(fmt.Sprintf("Expected: %s, Actual: %s", expectedResponse, actualResponse))
		}

		_, _ = conn.Write([]byte("bye"))
	})

	t.Run("InitClientListenerInvalidAddress", func(t *testing.T) {

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")
		go tcpServer.InitClientListener("localhost:")

		_, err := net.Dial("tcp", "localhost:1234")
		if err == nil {
			t.Error("Expected the connection to fail")
		}
	})

	t.Run("InitClientListenerConnectAfterListenerClosed", func(t *testing.T) {

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")
		go tcpServer.InitClientListener("localhost:1234")

		conn, err := net.Dial("tcp", "localhost:1234")
		if err != nil {
			t.Error("Failed to connect to server")
		}

		_, _ = conn.Write([]byte("bye"))

		_, err = net.Dial("tcp", "localhost:1234")
		if err == nil {
			t.Error("Expected the connection to fail")
		}
	})
}

func TestParseArg(t *testing.T) {

	const commandOffset = 3

	t.Run("parseArgKVSuccessful", func(t *testing.T) {
		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		expectedKey := "k"
		expectedValue := "v"
		buffer := []byte("put11k11v")

		actualKey, pos := tcpServer.parseArg(buffer[commandOffset:])

		if actualKey == "" || pos == -1 {
			t.Error("Failed to retrieve the key!")
		}

		actualValue, _ := tcpServer.parseArg(buffer[pos+commandOffset:])

		if actualValue == "" {
			t.Error("Failed to retrieve the value!")
		}

		if expectedKey != actualKey {
			t.Error(fmt.Sprintf("Expected key: %s, Actual key: %s", expectedKey, actualKey))
		}

		if expectedValue != actualValue {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedValue, actualValue))
		}
	})

	t.Run("parseArgKVInvalidBuffer", func(t *testing.T) {
		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")
		arg, pos := tcpServer.parseArg([]byte{})
		if arg != "" || pos != -1 {
			t.Error("Expected this to fail as invalid buffer was passed in")
		}
	})

	t.Run("parseArgKVValueMissing", func(t *testing.T) {
		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		expectedKey := "k"
		expectedValue := ""
		buffer := []byte("put11k")

		// get key
		actualKey, pos := tcpServer.parseArg(buffer[commandOffset:])

		if actualKey == "" || pos == -1 {
			t.Error("Failed to retrieve the key!")
		}

		// get value
		actualValue, pos1 := tcpServer.parseArg(buffer[pos+commandOffset:])

		if actualValue != "" || pos1 != -1 {
			t.Error("Expected retrieving the value to fail")
		}

		if expectedKey != actualKey {
			t.Error(fmt.Sprintf("Expected key: %s, Actual key: %s", expectedKey, actualKey))
		}

		if expectedValue != actualValue {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedValue, actualValue))
		}
	})

	t.Run("parseArgKUnexpectedValue", func(t *testing.T) {
		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		expectedKey := "k"
		buffer := []byte("del11k")

		actualKey, pos := tcpServer.parseArg(buffer[commandOffset:])

		if actualKey == "" || pos == -1 {
			t.Error("Failed to retrieve the key!")
		}

		secondArg, pos1 := tcpServer.parseArg(buffer[pos+commandOffset:])

		if secondArg != "" || pos1 != -1 {
			t.Error("Expected to fail retrieving a second arg")
		}

		if expectedKey != actualKey {
			t.Error(fmt.Sprintf("Expected key: %s, Actual key: %s", expectedKey, actualKey))
		}
	})

	t.Run("parseArgMultipleArgsStress", func(t *testing.T) {
		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		buffer := []byte("put11k11v14test15value15lorem15ipsum13del16value114nine13ten17destiny14halo15seven15eight15maria212dominicxasgd15Apple15jesus183m4aNu3L16GoLang")

		expectedResults := [20]string{
			"k", "v", "test", "value", "lorem", "ipsum", "del", "value1", "nine", "ten",
			"destiny", "halo", "seven", "eight", "maria", "dominicxasgd", "Apple", "jesus", "3m4aNu3L", "GoLang",
		}

		actualResults := [20]string{}
		pos := 0
		for i := 0; i < 20; i++ {
			temp := 0
			actualResults[i], temp = tcpServer.parseArg(buffer[pos+commandOffset:])
			pos += temp

			if pos == -1 || actualResults[i] == "" {
				t.Error(fmt.Sprintf("Failed to parse arg. Expected: %s, Actual: %s", expectedResults[i], actualResults[i]))
			}
		}

		if len(actualResults) == 0 {
			t.Error("Results are empty, failed to parse any args")
		}

		if expectedResults != actualResults {
			t.Error(fmt.Sprintf("Expected results: %v, Actual results: %v", expectedResults, actualResults))
		}
	})
}

func TestPut(t *testing.T) {

	t.Run("putCreate", func(t *testing.T) {
		expected := "ack"
		expectedVal := "val11v"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		actual1 := tcpServer.put("k", "v")

		if actual1 != expected {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expected, actual1))
		}

		actualVal := tcpServer.get("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}

	})

	t.Run("putUpdate", func(t *testing.T) {
		expected := "ack"
		expectedVal := "val11v"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		actual1 := tcpServer.put("k", "v")

		if actual1 != expected {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expected, actual1))
		}

		actualVal := tcpServer.get("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}

		actual2 := tcpServer.put("k", "v")

		if actual2 != expected {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expected, actual2))
		}

		actualVal = tcpServer.get("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}
	})
}

func TestGet(t *testing.T) {

	t.Run("getSuccessfull", func(t *testing.T) {
		expected := "ack"
		expectedVal := "val11v"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		actual1 := tcpServer.put("k", "v")

		if actual1 != expected {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expected, actual1))
		}

		actualVal := tcpServer.get("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}
	})

	t.Run("getKeyNotFound", func(t *testing.T) {
		expectedVal := "nil"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		actualVal := tcpServer.get("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}
	})
}

func TestDelete(t *testing.T) {

	t.Run("deleteSuccessful", func(t *testing.T) {
		expectedVal := "ack"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		tcpServer.put("k", "v")
		actualVal := tcpServer.delete("k")

		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}
	})

	t.Run("deleteKeyNotFound", func(t *testing.T) {
		expectedVal := "ack"

		tcpServer := NewDataServer(store.NewDataStore(), true, "server.log", "")

		actualVal := tcpServer.delete("k")
		if actualVal != expectedVal {
			t.Error(fmt.Sprintf("Expected Value: %s, Actual value: %s", expectedVal, actualVal))
		}
	})
}
