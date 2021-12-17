package dataServer

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/Emanuel-Nunes/Go-TCPServer/store"
)

type DataServer struct {
	udpListenerConn *net.UDPConn
	tcpListener     net.Listener
	store           *store.DataStore
	udpOn           bool
	tcpOn           bool
	standAlone      bool
	log             *log.Logger
	udpIP           string
	udpConn         *net.UDPConn
}

func NewDataServer(store *store.DataStore, standAlone bool, logFile string, udpIP string) *DataServer {

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	dataServer := DataServer{
		store:      store,
		udpOn:      false,
		tcpOn:      false,
		standAlone: standAlone,
		log:        log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		udpIP:      udpIP,
	}

	return &dataServer
}

// TCP listener for client requests
func (ds *DataServer) InitClientListener(address string) {
	log.Println("Starting Server")

	var err error
	ds.tcpListener, err = net.Listen("tcp4", address)
	if err != nil {
		log.Println("TCP Listener failed.")
		return
	}

	ds.tcpOn = true

	for {
		log.Println("Waiting for client connection")
		connection, err := ds.tcpListener.Accept()
		if err != nil {
			break
		}
		log.Println("Connection accepted")
		log.Println("Handling client request")
		go ds.handleTCP(connection)
	}
}

func (ds *DataServer) handleTCP(c net.Conn) {

	for {
		buffer := make([]byte, 2048)
		_, err := c.Read(buffer)

		if err != nil {
			fmt.Fprintf(c, "err")
			return
		}

		index := 0
		upperBound := 3
		commandString := string(buffer[index:upperBound])

		switch commandString {
		case "get":
			// get key
			key, _ := ds.parseArg(buffer[upperBound:])

			if key == "" {
				// Failure to retrieve arg
				fmt.Fprintf(c, "err")
				continue
			}

			fmt.Fprintf(c, ds.get(key))
		case "del":
			// get key
			key, _ := ds.parseArg(buffer[upperBound:])

			if key == "" {
				// Failure to retrieve arg
				fmt.Fprintf(c, "err")
				continue
			}

			fmt.Fprintf(c, ds.delete(key))

			if !ds.standAlone {
				// Notify cluster leader
				ds.broadcast(strings.TrimSpace(string(buffer)))
			}

		case "put":
			// get key
			key, pos := ds.parseArg(buffer[upperBound:])

			if key == "" || pos == -1 {
				// Failure to retrieve arg
				fmt.Fprintf(c, "err")
			}

			// get value
			value, _ := ds.parseArg(buffer[pos+3:])

			if value == "" {
				// We expect a value with a put request...
				fmt.Fprintf(c, "err")
			}

			fmt.Fprintf(c, ds.put(key, value))

			if !ds.standAlone {
				// Notify cluster leader
				log.Println("Notifying Cluster")
				ds.broadcast(strings.Trim(string(buffer), "\x00"))
			}
		case "bye":
			// Shutdown
			if ds.udpOn {
				ds.udpListenerConn.Close()
				ds.udpConn.Close()
			}
			if ds.tcpOn {
				ds.tcpListener.Close()
			}
		}
	}
}

// UDP listener for distributed store cluster communication
func (ds *DataServer) InitClusterListener() {

	//local address
	la, err := net.ResolveUDPAddr("udp4", ds.udpIP)
	if err != nil {
		log.Println(err)
		return
	}

	ds.udpListenerConn, err = net.ListenUDP("udp4", la)
	if err != nil {
		log.Println("UDP Listener failed.", ds.udpIP)
		log.Println(err)
		return
	}

	log.Printf("Server listening %s\n", ds.udpListenerConn.LocalAddr().String())
	ds.udpOn = true

	for {
		ds.handleUDP(ds.udpListenerConn)
	}
}

func (ds *DataServer) handleUDP(conn *net.UDPConn) {

	buffer := make([]byte, 2048)
	length, remote, err := conn.ReadFromUDP(buffer[:])

	if strings.Split(remote.String(), ":")[0] == strings.Split(conn.LocalAddr().String(), ":")[0] {
		log.Println("Local addr: ", conn.LocalAddr().String(), ", Remote addr: ", remote.String(), "ignoring data")
		return
	}

	if err != nil {
		log.Println("Failed to read:", err)
		return
	}

	data := strings.Trim(string(buffer[:length]), "\x00")
	log.Printf("received: %s from %s\n", data, remote)

	commandString := string(buffer[:3])

	switch commandString {
	case "del":
		key, _ := ds.parseArg(buffer[3:])

		if key == "" {
			// Failure to retrieve arg
			log.Println("Get failed key")
			return
		}

		ds.delete(key)
	case "put":
		key, pos := ds.parseArg(buffer[3:])

		if key == "" {
			log.Println("Put failed key")
			return
		}

		value, pos1 := ds.parseArg(buffer[pos+3:])

		if value == "" || pos1 == -1 {
			log.Println("Put failed value")

		}

		ds.put(key, value)
	default:
		log.Println("Default case")
	}
}

func (ds *DataServer) broadcast(msg string) {
	log.Println("In broadcast")
	n, err := ds.udpConn.Write([]byte(msg))
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("%d bytes written\n", n)
}

// rename later
func (ds *DataServer) SetupUDPConn() {

	//local address
	t := strings.Split(ds.udpIP, ":")[0] + ":8001"

	la, err := net.ResolveUDPAddr("udp4", t)
	if err != nil {
		log.Println(err)
		log.Println(ds.udpIP)
		return
	}

	ra, err := net.ResolveUDPAddr("udp4", "255.255.255.255:8000")
	if err != nil {
		log.Println(err)
		return
	}

	//dial
	ds.udpConn, err = net.DialUDP("udp4", la, ra)
	if err != nil {
		log.Println(err)
		return
	}
}

// Helpers

// return key and position in buffer so we can retrieve next args value easier
func (ds *DataServer) parseArg(buffer []byte) (string, int) {

	if len(buffer) <= 0 {
		// Invalid buffer
		log.Println("Invalid buffer")
		return "", -1
	}

	index := 0
	upperBound := 0

	lengthBytes, _ := strconv.Atoi(string(buffer[index]))

	if lengthBytes < 1 || lengthBytes > 9 {
		// error with length
		log.Println("Invalid first part of arg")
		return "", -1
	}

	index++
	upperBound = (index + lengthBytes)

	argLength, _ := strconv.Atoi(string(buffer[index:upperBound]))

	// need to check digits are correct
	if lengthBytes != getDigits(argLength) {
		// arg length is 0, fail!
		log.Println("Arg length is invalid")
		return "", -1
	}

	index += lengthBytes
	upperBound += argLength

	argValue := string(buffer[index:upperBound])

	if len(argValue) != argLength {
		// string isn't the length we expected...
		log.Println("Arg length is not what we expected")
		return "", -1
	}

	index += upperBound - index

	return argValue, index
}

func getDigits(n int) int {

	if n < 0 {
		return -1337
	}

	digits := 0
	for n != 0 {
		n /= 10
		digits += 1
	}

	return digits
}

// Data store functions
func (ds *DataServer) put(key, value string) string {
	responseChannel := make(chan interface{})
	msg := store.NewStoreMessage(responseChannel, []string{key, value})
	ds.store.Put(msg)
	<-responseChannel
	return "ack"
}

func (ds *DataServer) get(key string) string {
	responseChannel := make(chan interface{})
	msg := store.NewStoreMessage(responseChannel, key)
	ds.store.Get(msg)
	result := <-responseChannel

	convertedResults, ok := result.(store.GetContents)
	if !ok || convertedResults.Err != nil || convertedResults.Value == "" {
		return "nil"
	}

	digits := getDigits(len(convertedResults.Value))

	return fmt.Sprintf("val%d%d%s", digits, len(convertedResults.Value), convertedResults.Value)
}

func (ds *DataServer) delete(key string) string {
	responseChannel := make(chan interface{})
	msg := store.NewStoreMessage(responseChannel, key)
	ds.store.Delete(msg)
	<-responseChannel
	return "ack"
}
