package main

import (
	"bufio"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

func main() {
	config.ParseCLIFlags()

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection: ", err.Error())
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// The data we receive is a command in the form of an array, where the first
	// element is the command and the rest are optional args.
	for {
		parsed := parser.NewRESPParser(scanner).Parse()
		// Assert `parsed` is of form CommandArgs
		command, ok := parsed.(handler.CommandArgs)
		if !ok {
			break
		}
		conn.Write(handler.Handle(command))
	}
}
