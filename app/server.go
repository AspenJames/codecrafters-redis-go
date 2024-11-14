package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handler"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

func main() {
	config.ParseCLIFlags()
	// It's safe to ignore the `ok` flag on these config values, since we know
	// they're set in config.ParseCLIFlags().
	dir, _ := config.Get("dir")
	dbfilename, _ := config.Get("dbfilename")
	replicaof, _ := config.Get("replicaof")
	port, _ := config.Get("port")

	// Load from rdb file.
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

	// Initialize replication.
	if replicaof != "" {
		// replicaof is formatted like "<MASTER HOST> <MASTER PORT>"
		// we want an address like "<MASTER HOST>:<MASTER PORT>"
		address := strings.Replace(replicaof, " ", ":", 1)

		replClient := handler.NewReplicationClient(address)
		err := replClient.Init()
		if err != nil {
			log.Fatal("[main] Replication error:", err)
		}
		go replClient.Handle()
	}

	// Run listener.
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

	// Get the clientAddr from conn, with the port section removed.
	split := strings.Split(conn.LocalAddr().String(), ":")
	clientAddr := strings.Join(split[:len(split)-1], ":")

	// The data we receive is a command in the form of an array, where the first
	// element is the command and the rest are optional args.
	for {
		cmdCtx := &handler.Ctx{}
		parsed := parser.NewRESPParser(scanner).Parse()
		// Assert `parsed` is of form CommandArgs
		command, ok := parsed.(handler.CommandArgs)
		if !ok || len(command) == 0 {
			break
		}
		cmdCtx.SetCmd(command[0].(string))
		cmdCtx.SetArgs(command[1:])
		cmdCtx.SetClientAddr(clientAddr)
		cmdCtx.SetConn(conn)
		for _, resp := range handler.Handle(cmdCtx) {
			// Without _something_ here reading resp though some sort of
			// formatting, the RDB data from PSYNC won't actually write out.
			// ¯\_(ツ)_/¯
			_ = fmt.Sprint(resp)
			_, err := conn.Write(resp)
			if err != nil {
				log.Println("[main] Error writing command response: ", err.Error())
			}
		}
	}
}
