package main

import (
	"Minger/server/src"
)

func main() {
	var robotServer server.RobotServer
	wc := server.NewWebChat()
	go wc.Serve()
	robotServer.Serve("127.0.0.1:4500", wc)
}
