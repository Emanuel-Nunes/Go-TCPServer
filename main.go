package main

import (
	"flag"
	"fmt"

	"github.com/Emanuel-Nunes/Go-TCPServer/dataServer"
	"github.com/Emanuel-Nunes/Go-TCPServer/store"
)

func main() {
	var (
		tcpListenIP string
		udpListenIP string
		standAlone  bool
		logFile     string
	)

	flag.StringVar(&tcpListenIP, "tcpListenIP", "127.0.0.1:1234", "port for tcp listener")
	flag.StringVar(&udpListenIP, "udpListenIP", "127.0.0.1:8000", "ip:port for udp listener")
	flag.BoolVar(&standAlone, "standalone", false, "set if using server outside of a cluster")
	flag.StringVar(&logFile, "log", "server.log", "log file name")
	flag.Parse()

	fmt.Println(standAlone)

	dataServer := dataServer.NewDataServer(store.NewDataStore(), standAlone, logFile, udpListenIP)

	if !standAlone {
		dataServer.SetupUDPConn()
		go dataServer.InitClusterListener()
	}

	dataServer.InitClientListener(tcpListenIP)
}
