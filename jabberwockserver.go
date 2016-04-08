package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
)

var ip string
var port int
var quiet bool

func main() {
	parseFlags()
	if !quiet {
		// ANSI red text: \033[0;31m
		// ANSI blue text: \033[0;34m
		// ANSI black text: \033[0m
		// This will print something like:
		// Starting server on 127.0.0.1:5000
		fmt.Printf("Starting server on \033[0;31m")
		fmt.Printf("%v\033[0m:\033[0;34m%v\033[0m\n",
			ip, port)
	}
	bindAndListen()
}

func handleConnection(conn net.Conn) {
	fmt.Printf("Accepted connection: %v\n", conn)
}

func bindAndListen() {
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

func parseFlags() {
	flag.StringVar(&ip, "ip", "127.0.0.1",
		"specifies the IP address this jabberwock server will bind to")
	flag.IntVar(&port, "port", 5000,
		"specifies the port this jabberwock server will bind to")
	flag.BoolVar(&quiet, "quiet", false,
		"specifies whether to be quiet")
	flag.Parse()
}
