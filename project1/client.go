package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultPort = 27998
	sslEnabled  = false
)

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
