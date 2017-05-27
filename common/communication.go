package common

import (
	"net"
	"bufio"
	"encoding/binary"
	"errors"
	"sync"
)

type Communication struct {
	conn      *net.TCPConn
	writeLock *sync.Mutex
	readLock  *sync.Mutex
}

func NewCommunication(conn *net.TCPConn) Communication {
	var com Communication
	com.conn = conn
	com.writeLock = &sync.Mutex{}
	com.readLock = &sync.Mutex{}
	return com
}

func (c *Communication) IsConnectionAlive() bool {
	if c.conn == nil {
		return false
	}
	return true
}

func (c *Communication) Close() {
    if c.IsConnectionAlive() {
		c.conn.Close()
	}
}

func (c *Communication) Read() *Message {
	// exit when connection is nil
	if c.conn == nil {
		return nil
	}
	c.readLock.Lock()
	defer c.readLock.Unlock()
	reader := bufio.NewReader(c.conn)
	var userSize int16
	var contentSize int32
	var createTime int64
	// user size
	if err := binary.Read(reader, binary.LittleEndian, &userSize); err != nil {
		return nil
	}
	// user
	user := string(readNBytes(reader, int32(userSize)))
	// create time
	if err := binary.Read(reader, binary.BigEndian, &createTime); err != nil {
		return nil
	}

	// format
	format := string(readNBytes(reader, 3))
	// heart beat
	if format == "bat" {
		return nil
	}

	// content size
	if err := binary.Read(reader, binary.LittleEndian, &contentSize); err != nil {
		return nil
	}

	// content
	content := readNBytes(reader, contentSize)
	return &Message{User: user, CreateTime: createTime, Format: format, Content: content}
}

func (c *Communication) Write(message *Message) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	writer := bufio.NewWriter(c.conn)

	// write user size
	if err := binary.Write(writer, binary.LittleEndian, int16(len(message.User))); err != nil {
		return err
	}
	// write user
	if nn, err := writer.Write([]byte(message.User)); nn != len(message.User) || err != nil {
		return errors.New("write username failure")
	}
	// write create time
	if err := binary.Write(writer, binary.BigEndian, message.CreateTime); err != nil {
		return err
	}
	// write format
	if nn, err := writer.Write([]byte(message.Format)); nn != 3 || err != nil {
		return errors.New("write format failure")
	}
	// write content size
	if err := binary.Write(writer, binary.LittleEndian, int32(len(message.Content))); err != nil {
		return err
	}
	// write content
	if nn, err := writer.Write(message.Content); nn != len(message.Content) || err != nil {
		return errors.New("write data failure")
	}
	return writer.Flush()
}

func readNBytes(reader *bufio.Reader, size int32) []byte {
	bytes := make([]byte, 0, size)
	for len(bytes) < cap(bytes) {
		needs := cap(bytes) - len(bytes)
		temp := make([]byte, needs, needs)
		if _, err := reader.Read(temp); err != nil {
			continue
		}
		bytes = append(bytes, temp...)
	}
	return bytes
}
