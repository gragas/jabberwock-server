package main

import (
	"flag"
	"fmt"
	"github.com/gragas/jabberwock-server/game"
)

func main() {
	var ip string
	var port int
	var quiet, debug bool
	flag.StringVar(&ip, "ip", "127.0.0.1",
		"specifies the IP address this jabberwock server will bind to")
	flag.IntVar(&port, "port", 5000,
		"specifies the port this jabberwock server will bind to")
	flag.BoolVar(&quiet, "quiet", false,
		"specifies whether to be quiet")
	flag.BoolVar(&debug, "debug", false,
		"specifies whether to print debug info")
	flag.Parse()
	if !quiet {
		fmt.Printf("Starting server on \033[0;31m")
		fmt.Printf("%v\033[0m:\033[0;34m%v\033[0m\n",
			ip, port)
	}
	game.StartGame(ip, port, quiet, debug)
}
