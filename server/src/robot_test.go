package server

import (
	. "gopkg.in/check.v1"
	"testing"
)
// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type ServerSuite struct {
	addr   string
	server RobotServer
}

var _ = Suite(&ServerSuite{})

// func (s *ServerSuite) SetUpTest(c *C) {
// 	s.addr = "127.0.0.1:4500"
// 	s.server = RobotServer{}
// 	wc :=NewWebChat()
// 	go s.server.Serve(s.addr, wc)
// }
//
// func (s *ServerSuite) TearDownSuite(c *C) {
// 	s.server.Close()
// }
//
// func (s *ServerSuite) TestReadAndWrite(c *C) {
// 	var tcpAddr *net.TCPAddr
// 	tcpAddr, _ = net.ResolveTCPAddr("tcp", s.addr)
// 	com := NewCommunication()
// 	var err error
// 	com.conn, err = net.DialTCP("tcp", nil, tcpAddr)
// 	c.Assert(err, IsNil)
// 	//time.Sleep(1000 * time.Second)
//
// 	// read
// 	for i := 0; i < 2; i++ {
// 		message := &Message{User:  "username" + strconv.Itoa(i),
// 			CreateTime:            1348831860 + int64(i),
// 			Format:                "str",
// 			Content:               []byte("test" + strconv.Itoa(i)),
// 		}
// 		go com.Write(message)
// 	}
//
// 	// wait two seconds for server to respond
// 	time.Sleep(5 * time.Second)
//
// 	c.Assert(s.server.messages.Size(), Equals, 2)
// 	mes := s.server.messages.Pop()
// 	c.Assert(mes.User, Equals, "username0")
// 	c.Assert(mes.Format, Equals, "str")
// 	c.Assert(string(mes.Content), Equals, "test0")
//
// 	// write
// 	message := &Message{User:  "username",
// 		CreateTime:            1348831860,
// 		Format:                "str",
//		Content:               []byte("test"),
//	}
//	s.server.Write(message)
//	message1 := com.Read()
//	c.Assert(*message1, DeepEquals, *message)
// }
