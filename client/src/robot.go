package client

import (
	"net"
	"errors"
    . "minger/common"
)

type Robot struct {
	com Communication
	messages *MessageQueue
}

func (r *Robot) Connect(addr string) (err error) {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", addr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	r.com = NewCommunication(conn)
	return err
}

func (r *Robot) Read() {
	var message *Message
	// read data
	for {
		message = r.com.Read()
		if message != nil {
			r.messages.Push(message)
		}
	}
}

func (r *Robot) Write(message *Message) error {
	if !r.com.IsConnectionAlive() {
		return errors.New("The connection is already closed.")
	}
	return r.com.Write(message)
}
