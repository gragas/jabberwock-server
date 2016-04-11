package game

import (
	"bytes"
	"fmt"
//	"github.com/gragas/jabberwock-server/inventory"
//	"github.com/gragas/jabberwock-server/entity"
	"github.com/gragas/jabberwock-lib/consts"
	"io"
	"net"
	"strconv"
	"time"
)

func StartGame(ip string, port int, quiet bool, debug bool) {
	ch := make(chan string)
	go loop(ch, debug)
	bindAndListen(ip, port, ch, quiet)
}

func loop(ch <-chan string, debug bool) {
	for {
		startTime := time.Now()
		
		select {
		case msg := <- ch:
			handleMessage(msg, debug)
		default:
			if debug {
				fmt.Printf("Nothing received!\n")
			}
		}

		// do stuff

		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)
		if elapsedTime < consts.TicksPerFrame {
			time.Sleep(consts.TicksPerFrame - elapsedTime)
		}
	}
}

func handleMessage(msg string, debug bool) {
	if debug {
		fmt.Printf("%s\n", msg)
	}
}

func handleConnection(conn net.Conn, ch chan<- string, quiet bool) {
	if !quiet {
		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
	}
	var buf bytes.Buffer
	io.Copy(&buf, conn)
	ch <- buf.String()
}

func bindAndListen(ip string, port int, ch chan<- string, quiet bool) {
	binding := ip + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", binding)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn, ch, quiet)
	}
}
