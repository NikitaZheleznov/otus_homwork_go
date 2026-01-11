package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	exitChan := make(chan struct{})
	errorChan := make(chan error, 2)

	go func() {
		select {
		case <-sigChan:
			fmt.Fprintln(os.Stderr, "...Interrupted")
			close(exitChan)
		case <-errorChan:
			return
		}
	}()

	go func() {
		if err := client.Send(); err != nil {
			errorChan <- err
		} else {
			close(exitChan)
		}
	}()

	go func() {
		if err := client.Receive(); err != nil {
			errorChan <- err
		} else {
			close(exitChan)
		}
	}()

	select {
	case err := <-errorChan:
		fmt.Fprintln(os.Stderr, err)
		client.Close()
		os.Exit(1)
	case <-exitChan:
	}
}
