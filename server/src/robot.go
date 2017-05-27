package server

import (
	"errors"
	"net"
	. "Minger/common"
)

type RobotServer struct {
	// nowadays there is only one robotserver
	// connMap map[string]*net.TCPConn
	com      Communication
	messages *MessageQueue
	chat     *WebChat
}

// Serve open a listener to listen for client. addr is the address:port such as 127.0.0.1:4500
func (s *RobotServer) Serve(addr string, chat *WebChat) {
	// s.writeLock = &sync.Mutex{}
	// s.readLock = &sync.Mutex{}
	s.chat = chat
	//	s.chat.SetMessageHandler(s.handleMessages)
	var tcpAddr *net.TCPAddr
	s.messages = NewMessageQueue(100)
	// set message handler
	//s.messages.SetHandler(s.handleMessages)
	tcpAddr, _ = net.ResolveTCPAddr("tcp", addr)

	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	defer tcpListener.Close()
	// nowadays there is only one robotserver.
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
        s.com = NewCommunication(tcpConn)
		//fmt.Println("A client connected : " + s.conn.RemoteAddr().String())
		go s.read()
	}
	//fmt.Println(s.conn)
	// 新连接加入map

	// for multiple robots
	//for {
	//    tcpConn, err := tcpListener.AcceptTCP()
	//    if err != nil {
	//        continue
	//    }

	//    fmt.Println("A client connected : " + tcpConn.RemoteAddr().String())
	//    // 新连接加入map
	//    s.connMap[tcpConn.RemoteAddr().String()] = tcpConn
	//    tcpPipe(tcpConn)
	//}
}

func (s *RobotServer) Close() {
    s.com.Close()
	s.messages = nil
}

func (s *RobotServer) read() {
	var message *Message
	// read data
	for {
		message = s.com.Read()
		if message != nil {
			s.messages.Push(message)
		}
	}
}

func (s *RobotServer) Write(message *Message) error {
	if !s.com.IsConnectionAlive() {
		return errors.New("The connection is already closed.")
	}
	return s.com.Write(message)
}

// func (s *RobotServer) handleMessages() {
// 	// if webchat is not online (for debug)
// 	if s.chat == nil {
// 		return
// 	}
// 	if s.messages.Size() > 0 {
// 		s.chat.sendMessage(s.messages.Pop())
// 	}
// 	// return to user a default message if the connection is closed
// 	if s.conn == nil && s.chat.messages.Size() > 0 {
// 		message := s.chat.messages.Pop()
// 		s.chat.sendMessage(&Message{User: message.User, CreateTime: time.Now().Unix(),
// 			Format:                       "str", Content: []byte("小豆儿出去玩了，请过一会再来找Ta吧!")})
// 	} else if s.chat.messages.Size() > 0 {
// 		s.Write(s.chat.messages.Pop())
// 	}
// }
