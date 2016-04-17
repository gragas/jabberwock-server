package game

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gragas/jabberwock-lib/consts"
	"github.com/gragas/jabberwock-lib/entity"
	"github.com/gragas/jabberwock-lib/player"
	"github.com/gragas/jabberwock-lib/protocol"
	"net"
	"strconv"
	"sync"
	"time"
)

var entityID uint64
var entityIDMutex *sync.Mutex
var players []player.Player

func generateEntityID(debug bool) uint64 {
	entityIDMutex.Lock()
	entityID++
	if debug {
		fmt.Printf("SERVER: Generated entity ID %v.\n", entityID)
	}
	entityIDMutex.Unlock()
	return entityID
}

func StartGame(ip string, port int, quiet bool, debug bool, done chan<- string) {
	entityID = uint64(protocol.GenerateEntityID)
	entityIDMutex = &sync.Mutex{}
	ch := make(chan string)
	go loop(ch, debug)
	bindAndListen(ip, port, ch, quiet, done)
}

func loop(ch chan string, debug bool) {
	for {
		startTime := time.Now()

		select {
		case msg := <-ch:
			handleMessage(msg, ch, debug)
		default:
			/* if debug {
			 *	fmt.Printf("Nothing received!\n")
			 * }
			 */
		}

		update(debug)

		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)
		if elapsedTime < consts.TicksPerFrame {
			time.Sleep(consts.TicksPerFrame - elapsedTime)
		}
	}
}

func update(debug bool) {
	if debug {
		
	}
}

func registerClient(ch chan string, player player.Player, debug bool) {
	if debug {
		fmt.Printf("SERVER: Registering client...\n")
	}
	player.ID = generateEntityID(debug)
	players = append(players, player)
	if debug {
		fmt.Printf("SERVER: New players state: %v\n", players)
	}
	ch <- string(protocol.Success) + string(protocol.EndOfMessage)
}

func handleMessage(msg string, ch chan string, debug bool) {
	malformedMessage := func(ch chan string) {
		fmt.Printf("SERVER: Received malformed message.\n")
		ch <- msg
	}

	var str string
	var long bool
	if len(msg) < 1 {
		malformedMessage(ch)
		return
	} else if len(msg) == 1 {
		str = protocol.Code(msg[0]).String()
	} else {
		str = protocol.Code(msg[0]).String() + msg[1:]
		long = true
	}
	if debug {
		fmt.Printf("SERVER: Received msg: %s\n", str)
	}
	switch protocol.Code(msg[0]) {
	case protocol.Register:
		if !long {
			malformedMessage(ch)
			return
		}
		entity, err := entity.FromBytes([]byte(msg[1:]))
		if err != nil {
			malformedMessage(ch)
			return
		}
		player := player.Player{Entity: *entity}
		registerClient(ch, player, debug)
	default:
		fmt.Printf("SERVER: Received unknown command: %s\n", str)
		ch <- msg
	}
}

func handleConnection(conn net.Conn, ch chan string, handled chan int, quiet bool) {
	if !quiet {
		fmt.Printf("SERVER: Accepted connection from %v\n", conn.RemoteAddr())
	}
	msg, err := bufio.NewReader(conn).ReadString(byte(protocol.EndOfMessage))
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
				protocol.Code(serverResponse[0]).String()+serverResponse[1:len(serverResponse)-1])
		}
		fmt.Fprintf(conn, serverResponse)
	}
}
