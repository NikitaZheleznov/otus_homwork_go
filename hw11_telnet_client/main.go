package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	timeoutFlag = flag.Duration("timeout", 10*time.Second, "connection timeout")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	address := args[0] + ":" + args[1]

	client := NewTelnetClient(
		address,
		*timeoutFlag,
		os.Stdin,
		os.Stdout,
	)

	if err := client.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer client.Close()

	go func() {
		if err := client.Send(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	if err := client.Receive(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		client.Close()
		os.Exit(1)
	}
}
