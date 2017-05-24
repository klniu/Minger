package main

import (
	"time"
	"fmt"
    . "minger/client/src"
)

func main() {
	a := NewAudioPlayer()
	a.AddFile("/home/klniu/Downloads/chongerfei.mp3")
	a.Play()
	fmt.Println("test")
	time.Sleep(5*time.Second)
	fmt.Println("test1")
	a.Pause()
	time.Sleep(5*time.Second)
	a.Play()
	time.Sleep(5*time.Second)
	fmt.Println("test2")
	a.Resume()
	time.Sleep(5*time.Second)
}