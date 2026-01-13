package main

import (
	"bufio"
	"io"
	"net"
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
	for scanner.Scan() {
		text := scanner.Text() + "\n"
		t.conn.SetWriteDeadline(time.Now().Add(t.timeout))
		if _, err := t.conn.Write([]byte(text)); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (t *telnetClient) Receive() error {
	reader := bufio.NewReader(t.conn)
	for {
		t.conn.SetReadDeadline(time.Now().Add(t.timeout))
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if _, err := t.out.Write([]byte(line)); err != nil {
			return err
		}
	}
}
