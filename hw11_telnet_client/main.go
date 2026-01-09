package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var timeoutFlag = flag.Duration("timeout", 10*time.Second, "connection timeout")

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
	errCh := make(chan error, 2)

	go func() {
		if err := client.Send(); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := client.Receive(); err != nil {
			errCh <- err
		}
	}()

	err := <-errCh
	client.Close()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
