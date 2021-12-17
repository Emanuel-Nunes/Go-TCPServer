package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	key   = "an expert from lorem ipsum"
	value = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Sed elementum mi et faucibus sollicitudin. Mauris ac ex sapien. " +
		"Vivamus lacinia posuere sem vitae venenatis. Aliquam erat volutpat. " +
		"Aliquam erat volutpat. In imperdiet velit sit amet sem lacinia " +
		"eleifend. Curabitur ac ex ut magna vehicula mollis sit amet sed " +
		"massa. Nullam auctor nunc elit, a consequat quam tristique non. " +
		"Fusce ut imperdiet dolor. Duis posuere luctus efficitur. Sed " +
		"facilisis massa sit amet leo dignissim consectetur. Aenean vehicula " +
		"est."
)

func main() {
	if len(os.Args) != 2 {
		logrus.Fatal("Please provide an address!")
	}

	c, err := net.Dial("tcp", os.Args[1])

	if err != nil {
		logrus.WithError(err).Fatal("Failed to startup")
	}

	write(c, "put11k11v")
	assert(c, "ack")

	write(c, "get11k")
	assert(c, "val11v")

	write(c, "get11v")
	assert(c, "nil")

	write(c, "get21v")
	assert(c, "err")

	write(c, "del11k")
	assert(c, "ack")

	write(c, "get11k")
	assert(c, "nil")

	_ = c.Close()

	c, err = net.Dial("tcp", os.Args[1])

	if err != nil {
		logrus.WithError(err).Fatal("Failed to reconnect")
	}

	write(c, "del11v")
	assert(c, "ack")
	write(c, "put226"+key+"3513"+value)
	assert(c, "ack")
	write(c, "get226"+key)
	assert(c, "val3513"+value)
	write(c, "del226"+key)
	assert(c, "ack")
	write(c, "bye")
	fmt.Println("DONE")
}

func write(c net.Conn, s string) {
	fmt.Printf("%s ... ", s[:3])

	if _, err := c.Write([]byte(s)); err != nil {
		logrus.WithError(err).Fatal("Write error")
	}
}

func assert(c net.Conn, expect string) {
	b := make([]byte, len(expect))

	if _, err := c.Read(b); err != nil {
		logrus.WithError(err).Fatal("Read error")
	}

	if string(b) != expect {
		fmt.Printf("FAIL (got %s, want %s)\n", string(b), expect)
		os.Exit(-1)
	} else {
		fmt.Println("PASS")
	}
}
