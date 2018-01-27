package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort    = 27998
	defaultSSLPort = 27999
	sslEnabled     = false
	buffer         = 256
	msgPrefix      = "cs3700spring2018"
)

// maps acceptable operands to wrapper functions that satisfy operationFunc
var acceptableOperands = map[string]operationFunc{
	"+": add,
	"-": subtract,
	"*": multiply,
	"/": divide,
}

// type that accepts two numbers and returns a number as a result.
// add, subtract, multiply, and divide below all satisfy this type.
type operationFunc func(int, int) int

func add(num1, num2 int) int      { return num1 + num2 }
func subtract(num1, num2 int) int { return num1 - num2 }
func multiply(num1, num2 int) int { return num1 * num2 }
func divide(num1, num2 int) int   { return num1 / num2 }

// creates and connects a socket based on the given host and port #.
// also sets the read and write buffer sizes to 256 bytes.
func connectSocket(host string, port int, ssl bool) (net.Conn, error) {
	strPort := strconv.Itoa(port)
	servAddr := net.JoinHostPort(host, strPort)

	var (
		conn net.Conn
		err  error
	)
	if ssl {
		conn, err = tls.Dial("tcp", servAddr, &tls.Config{
			InsecureSkipVerify: true, // typically unsafe, but ok for this project
		})
		if err != nil {
			return &tls.Conn{}, err
		}
	} else {
		tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
		if err != nil {
			return &net.TCPConn{}, fmt.Errorf("Error resolving TCP address: %s", err.Error())
		}

		conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return &net.TCPConn{}, fmt.Errorf("Error dialing TCP: %s", err.Error())
		}
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

// evaluates a response from a server and determines its validity.
// returns an error if deemed to be an unexpected response.
func evaluateResponse(resp string) (int, error) {
	fields := strings.Fields(resp)
	if fields[1] == "STATUS" {
		solution, err := evalExpr(fields[2], fields[3], fields[4]) // evaluate the given expression
		if err != nil {
			return 0, err
		}
		return solution, nil
	} else if fields[2] == "BYE" {
		secretFlag := fields[1]
		fmt.Printf("%v\n", secretFlag) // print secret flag and exit
		os.Exit(0)
		return 0, nil
	} else {
		return 0, fmt.Errorf("Unexpected response from server. %q", fields)
	}
}

// evaluates an expression returned by the server in a STATUS message
func evalExpr(num1, operand, num2 string) (int, error) {
	first, err := strconv.Atoi(num1)
	if err != nil {
		return 0, err
	}
	second, err := strconv.Atoi(num2)
	if err != nil {
		return 0, err
	}
	operation, ok := acceptableOperands[operand]
	if !ok {
		return 0, fmt.Errorf("Unrecognized operand returned")
	}
	return operation(first, second), nil
}

func main() {
	// read flags and args
	port := flag.Int("p", defaultPort, "port to use")
	ssl := flag.Bool("s", sslEnabled, "enable ssl if provided")
	flag.Parse()

	// hacky-ish way to set the port to the SSL default port
	// assuming the user did not explicitly define the port in
	// the command line args to be port 27998
	if *ssl && *port == defaultPort {
		*port = defaultSSLPort
	}

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Must only provide a hostname and NUID as arguments.")
		os.Exit(1)
	}
	hostname, nuid := args[0], args[1]

	conn, err := connectSocket(hostname, *port, *ssl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close() // close connection when finished

	// say hello
	message := fmt.Sprintf("%s HELLO %s\n", msgPrefix, nuid)
	for {
		// keep reading and writing messages until the program quits in evaluateResponse func
		if err := writeMessage(message, conn); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		resp, err := readMessage(conn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		solution, err := evaluateResponse(resp)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// generate solution message
		message = fmt.Sprintf("%s %d\n", msgPrefix, solution)
	}
}
