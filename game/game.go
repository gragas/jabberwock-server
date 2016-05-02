package game

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gragas/jabberwock-lib/consts"
	"github.com/gragas/jabberwock-lib/entity"
	"github.com/gragas/jabberwock-lib/player"
	"github.com/gragas/jabberwock-lib/protocol"
	"github.com/gragas/jabberwock-server/serverutils"
	"net"
	"strconv"
	"sync"
	"time"
)

var entityID uint64
var entityIDMutex *sync.Mutex
var updateCond  *sync.Cond
var modifyingEntities, updatingEntities bool
var entities map[uint64]entity.Entity
var players map[uint64]*player.Player
var jsonPlayers map[string]*player.Player
var readersToPlayers map[*bufio.Reader]*player.Player
var connections []net.Conn

func StartGame(ip string, port int, quiet bool, debug bool, dedicated bool, done chan bool) {
	/* initialize "global" variables */
	entities = make(map[uint64]entity.Entity)
	players = make(map[uint64]*player.Player)
	jsonPlayers = make(map[string]*player.Player)
	readersToPlayers = make(map[*bufio.Reader]*player.Player)
	entityID = uint64(protocol.GenerateEntityID)
	entityIDMutex = new(sync.Mutex)
	updateCond = sync.NewCond(&sync.Mutex{})
	updatingEntities = true
	/*********************************/

	go loop(debug)
	bindAndListen(ip, port, debug, quiet, dedicated, done)
}

func loop(debug bool) {
	for {
		startTime := time.Now()
		update(debug)
		broadcast(generateBroadcastString(), debug)
		endTime := time.Now()
		serverutils.ElapsedTime = endTime.Sub(startTime)
		serverutils.Delta = float32(serverutils.ElapsedTime) * 0.001
		if serverutils.ElapsedTime < consts.TicksPerFrame {
			time.Sleep(consts.TicksPerFrame - serverutils.ElapsedTime)
		}
	}
}

func update(debug bool) {
	updateCond.L.Lock()
	for modifyingEntities { updateCond.Wait() }
	updatingEntities = true
	
	for _, e := range entities {
		e.Update()
	}
	
	updatingEntities = false
	updateCond.L.Unlock()
	updateCond.Signal()
}

func broadcast(msg string, debug bool) {
	for _, c := range connections {
		fmt.Fprintf(c, msg)
	}
}

func listenTo(reader *bufio.Reader, conn net.Conn, debug bool) {
	for {
		msg, err := reader.ReadString(byte(protocol.EndOfMessage))
		if err != nil {
			fmt.Printf("SERVER: Disconnecting from %v\n", conn.RemoteAddr())
			// remove the connection
			for i, c := range connections {
				if c == conn {
					updateCond.L.Lock()
					for updatingEntities { updateCond.Wait() }
					modifyingEntities = true

					p := readersToPlayers[reader]
					if p != nil {
						delete(players, p.GetID())
						delete(jsonPlayers, strconv.FormatUint(p.GetID(), 10))
						delete(entities, p.GetID())
						delete(readersToPlayers, reader)
					}
					connections = append(connections[:i], connections[i+1:]...)
					broadcast(string(protocol.Disconnect) + strconv.FormatUint(p.GetID(), 10) + string(protocol.EndOfMessage), debug)
					
					modifyingEntities = false
					updateCond.L.Unlock()
					updateCond.Signal()
					break
				}
			}
			// stop listening
			return
		}
		if len(msg) < 2 {
			fmt.Printf("SERVER: Received msg that was too short from %v\n", conn.RemoteAddr())
			continue
		}
		contents := msg[1:len(msg)-1]
		// augment the state of the game based on this message
		switch protocol.Code(msg[0]) {
		case protocol.EntityStartMove:
			moveLocal(conn, msg, contents, true)
		case protocol.EntityStopMove:
			moveLocal(conn, msg, contents, false)
		default:
			fmt.Println(contents)
		}
	}
}

func bindAndListen(ip string, port int, debug bool, quiet bool, dedicated bool, done chan bool) {
	/* bind */
	binding := ip + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", binding)
	if err != nil {
		panic(err)
	}
	if !dedicated {
		done <- true
	}
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
			fmt.Fprintf(conn, "SERVER ERROR: Cannot unmarshal player from client." + string(protocol.EndOfMessage))
			continue
		}
		// success! generate an id for the player and handshake the newly connected client!
		p.SetID(generateEntityID(debug))
		marshalledPlayer := p.String()
		unmarshalledPlayer := new(player.Player)
		unmarshalledPlayer.FromBytes([]byte(marshalledPlayer))
		if *unmarshalledPlayer != *p {
			fmt.Printf("SERVER ERROR: Marshalled and unmarshalled players do not match\n")
			fmt.Fprintf(conn, "SERVER ERROR: Cannot register client." + string(protocol.EndOfMessage))
			continue
		}
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
		jsonPlayers[strconv.FormatUint(p.GetID(), 10)] = p
		entities[p.GetID()] = p
		readersToPlayers[reader] = p
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

func moveLocal(conn net.Conn, msg string, contents string, start bool) {
	l := len(contents)
	if l < 21 {
		fmt.Printf("SERVER: Received short move msg from %v: %v\n", conn.RemoteAddr(), msg)
		fmt.Printf("len(contents): %v; contents: %v\n", l, contents)
		return
	}
	id, err := entity.FromIDString(contents[1:])
	if err != nil {
		fmt.Printf("SERVER: Received unparseable move msg from %v: %v\n", conn.RemoteAddr(), msg)
		return
	}
	if entities[id] == nil || players[id] == nil {
		fmt.Printf("SERVER: No such entity or player with id %v\n", id)
		return
	}
	dir := entity.Direction(contents[0])
	if start {
		entity.StartMoveLocal(entities[id], dir)
	} else {
		entity.StopMoveLocal(entities[id], dir)
	}
}

func generateBroadcastString() string {
	if len(connections) == 0 {
		return ""
	}
	bytes, err := json.Marshal(jsonPlayers)
	if err != nil {
		panic(err)
	}
	return string(protocol.UpdatePlayers) + string(bytes) + string(protocol.EndOfMessage)
}
