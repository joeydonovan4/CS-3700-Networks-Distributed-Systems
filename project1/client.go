package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

const (
	defaultPort = 27998
	sslEnabled  = false
	buffer      = 256
)

// creates and connects a socket based on the given host and port #.
// also sets the read and write buffer sizes to 256 bytes.
func connectSocket(host string, port int) (net.Conn, error) {
	strPort := strconv.Itoa(port)
	servAddr := net.JoinHostPort(host, strPort)

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		return &net.TCPConn{}, fmt.Errorf("Error resolving TCP address: %s", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return &net.TCPConn{}, fmt.Errorf("Error dialing TCP: %s", err.Error())
	}

	return conn, nil
}

// writes a message to the server and returns an error if the write fails.
func writeMessage(message string, conn net.Conn) error {
	_, err := conn.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("Error writing message %s to server: %s", message, err.Error())
	}
	return nil
}

// reads a message from the server and returns an error if the read fails.
// the response is converted to a string.
func readMessage(conn net.Conn) (string, error) {
	resp := make([]byte, buffer)
	_, err := conn.Read(resp)
	if err != nil {
		return "", fmt.Errorf("Error reading message from server: %s", err.Error())
	}
	return string(resp), nil
}

func main() {
	// read flags and args
	port := flag.Int("p", defaultPort, "port to use")
	ssl := flag.Bool("s", sslEnabled, "enable ssl if provided")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Must only provide a hostname and NUID as arguments.")
		os.Exit(1)
	}
	hostname, nuid := args[0], args[1]

	fmt.Printf("port: %d, ssl: %v, hostname: %s, nuid: %s", port, ssl, hostname, nuid)
}
