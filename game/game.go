package game

import (
	"bufio"
	"fmt"
//	"github.com/gragas/jabberwock-server/inventory"
//	"github.com/gragas/jabberwock-server/entity"
	"github.com/gragas/jabberwock-lib/consts"
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

func registerClient(ch chan string, debug bool) {
	if debug {
		fmt.Printf("Registering client...\n")
	}
	ch <- "SUCCESS_\n"
}

func handleMessage(msg string, ch chan string, debug bool) {
	if debug {
		fmt.Printf("Received msg: %s\n", msg)
	}
	switch msg[:8] {
	case "REGISTER":
		registerClient(ch, debug)
	default:
		fmt.Printf("Unknown command: %s\n", msg)
	}
}

func handleConnection(conn net.Conn, ch chan string, handled chan int, quiet bool) {
	if !quiet {
		fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
	}
	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Bad message from %v\n", conn.RemoteAddr())
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
		if !quiet {
			fmt.Printf("Writing %s in response\n", serverResponse[:8])
		}
		fmt.Fprintf(conn, serverResponse)
	}
}
