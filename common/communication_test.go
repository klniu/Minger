package common

import (
    "net"

    . "gopkg.in/check.v1"
    "time"
)

type CommunicationSuite struct {
    tcpAddr *net.TCPAddr
    server  Communication
    client  Communication
}

var _ = Suite(&CommunicationSuite{})

func (s *CommunicationSuite) SetUpTest(c *C) {
    addr := "127.0.0.1:4500"
    s.tcpAddr, _ = net.ResolveTCPAddr("tcp", addr)
    // server start
}

func (s *CommunicationSuite) TearDownSuite(c *C) {
    s.client.Close()
    s.server.Close()
}

func (s *CommunicationSuite) TestReadAndWrite(c *C) {
    // server start
    go func() {
        tcpListener, _ := net.ListenTCP("tcp", s.tcpAddr)
        tcpConn, err := tcpListener.AcceptTCP()
        if err != nil {
            panic(err)
        }
        s.server = NewCommunication(tcpConn)
    }()
    // wait for server
    time.Sleep(5 * time.Second)

    // client dial
    conn, err := net.DialTCP("tcp", nil, s.tcpAddr)
    c.Assert(err, IsNil)
    s.client = NewCommunication(conn)

    // server read
    var messages []*Message
    go func() {
        for {
            message := s.server.Read()
            if message != nil {
                messages = append(messages, message)
            }
        }
    }()
    // client write
        message := &Message{User: "username",
            CreateTime:           1348831860,
            Format:               "str",
            Content:              []byte("test"),
        }
        c.Assert(s.client.Write(message), IsNil)
    // wait for server
    time.Sleep(5 * time.Second)


    // server read
    c.Assert(len(messages), Equals, 1)
    c.Assert(messages[0].User, Equals, "username")
    c.Assert(messages[0].Format, Equals, "str")
    c.Assert(string(messages[0].Content), Equals, "test")

    // write
    message = &Message{User: "username",
        CreateTime:           1348831860,
        Format:               "str",
        Content:              []byte("test"),
    }
    c.Assert(s.server.Write(message),IsNil)
    message1 := s.client.Read()
    c.Assert(*message1, DeepEquals, *message)
}
