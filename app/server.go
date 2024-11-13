package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

func main() {
	config.ParseCLIFlags()

	// Load from rdb file.
	dir, ok := config.Get("dir")
	if !ok {
		log.Fatal("[main] Unable to retrieve directory from config")
	}
	dbfilename, ok := config.Get("dbfilename")
	if !ok {
		log.Fatal("[main] Unable to retrieve dbfilename from config")
	}
	rdbFilepath := filepath.Join(dir, dbfilename)
	rdbFile, err := os.Open(rdbFilepath)
	if !os.IsNotExist(err) {
		if err != nil {
			log.Fatal("[main] Unable to open rdb file: ", err.Error())
		}
		defer rdbFile.Close()
		resp := parser.NewRDBParser(rdbFile).Parse()
		err = cache.GetDefaultCache().LoadRDB(resp)
		if err != nil {
			log.Fatal("[main] Unable to load RDB data into cache: ", err.Error())
		}
	}

	// Run listener.
	port, ok := config.Get("port")
	if !ok {
		log.Fatal("[main] Unable to retrieve port from config")
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("[main] Failed to bind to port ", port)
	}
	log.Println("[main] Listening on", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("[main] Error accepting connection: ", err.Error())
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
