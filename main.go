package main

import (
	"flag"
	"fmt"
	"github.com/gragas/jabberwock-server/game"
)

const (
	ipString    = "specifies the IP address this jabberwock server will bind to"
	portString  = "specifies the port this jabberwock server will bind to"
	quietString = "specifies whether to be quiet"
	debugString = "specifies whether to print debug info"
)

func main() {
	var ip string
	var port int
	var quiet, debug bool
	flag.StringVar(&ip, "ip", "127.0.0.1", ipString)
	flag.IntVar(&port, "port", 5000, portString)
	flag.BoolVar(&quiet, "quiet", false, quietString)
	flag.BoolVar(&debug, "debug", false, debugString)
	flag.Parse()
	if !quiet {
		fmt.Printf("Starting server on \033[0;31m")
		fmt.Printf("%v\033[0m:\033[0;34m%v\033[0m\n", ip, port)
	}
	game.StartGame(ip, port, quiet, debug)
}
