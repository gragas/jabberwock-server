package game

import (
	"bufio"
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
var entities map[uint64]entity.Entity
var players map[uint64]*player.Player
var connections []net.Conn

func StartGame(ip string, port int, quiet bool, debug bool, done chan bool) {
	/* initialize "global" variables */
	entities = make(map[uint64]entity.Entity)
	players = make(map[uint64]*player.Player)
	entityID = uint64(protocol.GenerateEntityID)
	entityIDMutex = &sync.Mutex{}
	/*********************************/

	go loop(debug)
	bindAndListen(ip, port, debug, quiet, done)
}

func loop(debug bool) {
	for {
		startTime := time.Now()
		update(debug)
		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)
		if elapsedTime < consts.TicksPerFrame {
			time.Sleep(consts.TicksPerFrame - elapsedTime)
		}
	}
}

func update(debug bool) {
	
}

func listenTo(reader *bufio.Reader, conn net.Conn, debug bool) {
	for {
		msg, err := reader.ReadString(byte(protocol.EndOfMessage))
		if err != nil {
			fmt.Printf("SERVER: Disconnecting from %v\n", conn.RemoteAddr())
			for i, c := range connections {
				if c == conn {
					connections = append(connections[:i], connections[i+1:]...)
					break
				}
			}
			return
		}
		if len(msg) < 2 {
			fmt.Printf("SERVER: Received msg that was too short from %v\n", conn.RemoteAddr())
			continue
		}
		contents := msg[1:len(msg)-1]
		// augment the state of the game based on this message
		switch protocol.Code(msg[0]) {
		default:
			fmt.Println(contents)
		}
	}
}

func bindAndListen(ip string, port int, debug bool, quiet bool, done chan bool) {
	/* bind */
	binding := ip + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", binding)
	if err != nil {
		panic(err)
	}
	done <- true
	/*********/

	/* continuously accept connections */
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("SERVER ERROR: %v\n", err)
			continue
		}
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString(byte(protocol.EndOfMessage))
		if err != nil {
			fmt.Printf("SERVER ERROR: %v\n", err)
			continue
		}
		if len(msg) < 2 {
			fmt.Printf("SERVER: New client sent short msg: %v\n", msg)
			continue
		}
		if protocol.Code(msg[0]) != protocol.Register {
			fmt.Printf("SERVER: New client (%v) sent invalid code: %v\n", conn.RemoteAddr(), err)
			continue
		}
		contents := msg[1:len(msg)-1]
		p := new(player.Player)
		err = p.FromBytes([]byte(contents))
		if err != nil {
			fmt.Printf("SERVER: Could not unmarshall player from new client (%v).\n", conn.RemoteAddr())
			continue
		}
		// success! generate an id for the player and handshake the newly connected client!
		p.SetID(generateEntityID(debug))
		fmt.Fprintf(conn, string(protocol.Success) + p.String() + string(protocol.EndOfMessage))
		// make sure that the client is on the same page...
		msg, err = reader.ReadString(byte(protocol.EndOfMessage))
		if err != nil {
			fmt.Printf("SERVER ERROR: %v\n", err)
			continue
		}
		if len(msg) != 2 || protocol.Code(msg[0]) != protocol.Handshake || protocol.Code(msg[1]) != protocol.EndOfMessage {
			fmt.Printf("SERVER: New client sent invalid handshake msg: %v\n", msg)
			continue			
		}
		// successful handshake! add the connection and the player
		players[p.GetID()] = p
		entities[p.GetID()] = p
		connections = append(connections, conn)
		go listenTo(reader, conn, debug)
	}
	/***********************************/
}

func generateEntityID(debug bool) uint64 {
	entityIDMutex.Lock()
	entityID++
	if debug {
		fmt.Printf("SERVER: Generated entity ID %v.\n", entityID)
	}
	entityIDMutex.Unlock()
	return entityID
}
