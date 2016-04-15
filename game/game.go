package game

import (
	"bufio"
	"errors"
	"fmt"
//	"github.com/gragas/jabberwock-server/inventory"
//	"github.com/gragas/jabberwock-server/entity"
	"github.com/gragas/jabberwock-lib/consts"
	"github.com/gragas/jabberwock-lib/protocol"
	"net"
	"strconv"
	"time"
)

func StartGame(ip string, port int, quiet bool, debug bool, done chan<- string) {
	ch := make(chan string)
	go loop(ch, debug)
	bindAndListen(ip, port, ch, quiet, done)
}

func loop(ch chan string, debug bool) {
	for {
		startTime := time.Now()
		
		select {
		case msg := <- ch:
			handleMessage(msg, ch, debug)
		default:
			if debug {
				// fmt.Printf("Nothing received!\n")
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

func registerClient(ch chan string, debug bool) {
	if debug {
		fmt.Printf("SERVER: Registering client...\n")
	}
	ch <- string(protocol.Success) + "\n"
}

func handleMessage(msg string, ch chan string, debug bool) {
	if len(msg) < 1 {
		fmt.Printf("SERVER: Received malformed message.\n")
		ch <- msg
		return
	}
	var str string
	if len(msg) == 1 {
		str = protocol.Code(msg[0]).String()
	} else {
		str = protocol.Code(msg[0]).String() + msg[1:]
	}
	if debug {
		fmt.Printf("SERVER: Received msg: %s\n", str)
	}
	switch protocol.Code(msg[0]) {
	case protocol.Register:
		registerClient(ch, debug)
	default:
		fmt.Printf("SERVER: Received unknown command: %s\n", str)
		ch <- msg
	}
}

func handleConnection(conn net.Conn, ch chan string, handled chan int, quiet bool) {
	if !quiet {
		fmt.Printf("SERVER: Accepted connection from %v\n", conn.RemoteAddr())
	}
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("SERVER: Bad message from %v\n", conn.RemoteAddr())
		return
	}
	ch <- msg[:len(msg)-1]
	handled <- 0
}

func bindAndListen(ip string,
	port int,
	ch chan string,
	quiet bool,
	done chan<- string) {
	
	binding := ip + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", binding)
	if err != nil {
		panic(err)
	}
	done <- "done"
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		handled := make(chan int)
		go handleConnection(conn, ch, handled, quiet)
		<-handled
		serverResponse := <-ch
		if len(serverResponse) < 2 {
			panic(errors.New("SERVER: Malformed server response.\n"))
		}
		if !quiet {
			fmt.Printf("SERVER: Writing '%s' in response\n",
				protocol.Code(serverResponse[0]).String() + serverResponse[1:len(serverResponse)-1])
		}
		fmt.Fprintf(conn, serverResponse)
	}
}
