package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (t *telnetClient) Connect() error {
	var err error
	t.conn, err = net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT)
		<-sigChan
		t.conn.Close()
	}()

	return nil
}

func (t *telnetClient) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *telnetClient) Send() error {
	scanner := bufio.NewScanner(t.in)
	if scanner.Scan() {
		text := scanner.Text() + "\n"
		_, err := t.conn.Write([]byte(text))
		return err
	}
	return scanner.Err()
}

func (t *telnetClient) Receive() error {
	reader := bufio.NewReader(t.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	_, err = t.out.Write([]byte(line))
	return err
}
