package game

import (
	"fmt"
//	"github.com/gragas/jabberwock-server/inventory"
//	"github.com/gragas/jabberwock-server/entity"
	"net"
	"strconv"
	"sync"
)

func StartGame(ip string, port int) {
	var wg sync.WaitGroup
	wg.Add(1)
	
	go bindAndListen(ip, port)
	wg.Wait()
}

func handleConnection(conn net.Conn) {
	fmt.Printf("Accepted connection from %v\n", conn.RemoteAddr())
}

func bindAndListen(ip string, port int) {
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
		go handleConnection(conn)
	}
}
