package game

import (
	"bufio"
	"encoding/json"
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
var entities []entity.Entity
var players []*player.Player
var connections []net.Conn

func generateEntityID(debug bool) uint64 {
	entityIDMutex.Lock()
	entityID++
	if debug {
		fmt.Printf("SERVER: Generated entity ID %v.\n", entityID)
	}
	entityIDMutex.Unlock()
	return entityID
}

func StartGame(ip string, port int, quiet bool, debug bool, done chan bool) {
	entityID = uint64(protocol.GenerateEntityID)
	entityIDMutex = &sync.Mutex{}
	listenToLoop := make(chan string)
	loopToListen := make(chan string)
	go loop(listenToLoop, loopToListen, debug)
	bindAndListen(ip, port, listenToLoop, loopToListen, quiet, done)
}

func loop(listenToLoop <-chan string, loopToListen chan<- string, debug bool) {
	for {
		startTime := time.Now()

		select {
		case msg := <-listenToLoop:
			handleMessage(msg, loopToListen, debug)
		default:
			/* if debug {
			 *	fmt.Printf("Nothing received!\n")
			 * }
			 */
		}

		update(debug)
		broadcast(generateBroadcastString())

		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)
		if elapsedTime < consts.TicksPerFrame {
			time.Sleep(consts.TicksPerFrame - elapsedTime)
		}
	}
}

func update(debug bool) {
	for _, e := range entities {
		e.Update()
	}
}

func broadcast(msg string) {
	for _, conn := range connections {
		// fmt.Printf("SERVER: Sending data to %v\n", conn.RemoteAddr())
		fmt.Fprintf(conn, msg)
	}
}

func generateBroadcastString() string {
	pBytes, err := json.Marshal(players)
	if err != nil {
		panic(err)
	}
	return string(protocol.UpdatePlayers) + string(pBytes) + string(protocol.EndOfMessage)
}

func registerClient(loopToListen chan<- string, p *player.Player, debug bool) {
	if debug {
		fmt.Printf("SERVER: Registering client...\n")
	}
	p.SetID(generateEntityID(debug))
	entities = append(entities, p)
	players = append(players, p)
	if debug {
		fmt.Printf("SERVER: New entities state: %v\n", entities)
	}
	loopToListen <- string(protocol.Success) + p.String() + string(protocol.EndOfMessage)
}

func extractID(msg string) uint64 {
	var index int
	var ok bool
	for i, c := range msg[1:] {
		if byte(c) != '_' {
			index = i+1
			ok = true
			break
		}
	}
	if !ok || index >= len(msg)-1 {
		fmt.Println("msg:", msg)
		panic(errors.New("msg does not contain an ID.\n"))
	}
	ID, err := strconv.ParseUint(msg[index:len(msg)-1], 10, 64)
	if err != nil {
		panic(err)
	}
	return ID
}

func handleMessage(msg string, loopToListen chan<- string, debug bool) {
	malformedMessage := func(loopToListen chan<- string) {
		fmt.Printf("SERVER: Received malformed message.\n")
		loopToListen <- msg
	}

	var str string
	var long bool
	if len(msg) < 1 {
		malformedMessage(loopToListen)
		return
	} else if len(msg) == 1 {
		str = protocol.Code(msg[0]).String()
	} else {
		str = protocol.Code(msg[0]).String() + msg[1:]
		long = true
	}
	if debug {
		// fmt.Printf("SERVER: Received msg: %s\n", str)
	}
	if !long {
		malformedMessage(loopToListen)
		return
	}
	switch protocol.Code(msg[0]) {
	case protocol.Register:
		var p player.Player
		err := p.FromBytes([]byte(msg[1:]))
		if err != nil {
			malformedMessage(loopToListen)
			return
		}
		registerClient(loopToListen, &p, debug)
	case protocol.EntityStartMoveRight:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StartMoveLocal(e, entity.Right)
				break
			}
		}
	case protocol.EntityStopMoveRight:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StopMoveLocal(e, entity.Right)
				break
			}
		}
	case protocol.EntityStartMoveLeft:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StartMoveLocal(e, entity.Left)
				break
			}
		}
	case protocol.EntityStopMoveLeft:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StopMoveLocal(e, entity.Left)
				break
			}
		}
	case protocol.EntityStartMoveUp:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StartMoveLocal(e, entity.Up)
				break
			}
		}
	case protocol.EntityStopMoveUp:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StopMoveLocal(e, entity.Up)
				break
			}
		}
	case protocol.EntityStartMoveDown:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StartMoveLocal(e, entity.Down)
				break
			}
		}
	case protocol.EntityStopMoveDown:
		ID := extractID(msg)
		// Make this a hashmap
		for _, e := range entities {
			if e.GetID() == ID {
				entity.StopMoveLocal(e, entity.Down)
				break
			}
		}
	default:
		fmt.Printf("SERVER: Received unknown command: %s\n", str)
		loopToListen <- msg
	}
}

func handleConnection(conn net.Conn, listenToLoop chan<- string, quiet bool) (string, error) {
	if !quiet {
		fmt.Printf("SERVER: Accepted connection from %v\n", conn.RemoteAddr())
	}
	msg, err := bufio.NewReader(conn).ReadString(byte(protocol.EndOfMessage))
	if err != nil {
		fmt.Printf("SERVER: Bad message from %v\n", conn.RemoteAddr())
		fmt.Printf("msg: %v\n", msg)
		return "", err
	}
	go func() {
		for {
			msg, err := bufio.NewReader(conn).ReadString(byte(protocol.EndOfMessage))
			if err != nil {
				fmt.Printf("SERVER: Client %v disconnected.\n", conn.RemoteAddr())
				// remove the connection
				var index int
				for i, c := range connections {
					if c == conn {
						index = i
						break
					}
				}
				if index+1 < len(connections) {
					connections = append(connections[:index], connections[index+1:]...)
				} else {
					connections = connections[:len(connections)-1]
				}
				return
			}
			listenToLoop <- msg
		}
	}()
	return msg[:len(msg)-1], nil
}

func bindAndListen(ip string, port int, listenToLoop chan<- string, loopToListen <-chan string,
	quiet bool, done chan bool) {

	binding := ip + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", binding)
	if err != nil {
		panic(err)
	}
	done <- true
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		connections = append(connections, conn)
		clientMessage, err := handleConnection(conn, listenToLoop, quiet)
		listenToLoop <- clientMessage
		serverResponse := <-loopToListen
		if err != nil || len(serverResponse) < 2 {
			fmt.Fprintf(conn, string(protocol.BadMessageError)+string(protocol.EndOfMessage))
			return
		}
		if !quiet {
			fmt.Printf("SERVER: Writing '%s' in response\n",
				protocol.Code(serverResponse[0]).String()+serverResponse[1:len(serverResponse)-1])
		}
		fmt.Fprintf(conn, serverResponse)
	}
}
