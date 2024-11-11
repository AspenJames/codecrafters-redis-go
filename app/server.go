package main

import (
	"bufio"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handlers"
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

// processData is a recursive function to parse the incoming data on `scanner`
// as RESP data. Any time we'd encounter an error, we instead return `nil` --
// area for improvement, but should work for our purposes.
func processData(scanner *bufio.Scanner) interface{} {
	if !scanner.Scan() {
		if scanner.Err() != nil {
			log.Println("[processData] Scan() encountered error: ", scanner.Err())
		}
		return nil
	}

	// Read the first byte to determine how to process the data
	// https://redis.io/docs/latest/develop/reference/protocol-spec/#resp-protocol-description
	///// At this point in the exercise, we only need to handle Array and String
	data := scanner.Bytes()
	switch fb := data[0]; fb {
	case byte('*'):
		// Array, data[1:] contains the length
		arrLen, err := strconv.Atoi(string(data[1:]))
		if err != nil {
			log.Println("[processData] Error while processing array length: ", err.Error())
			return nil
		}
		arr := []interface{}{}
		// Read `arrLen` number of elements into `arr`
		for arrLen > 0 {
			arr = append(arr, processData(scanner))
			arrLen--
		}
		return arr

	case byte('$'):
		// Bulk string, data[1:] contains length
		// Length is unnecessary, as `scanner` is just going to read untiil
		// the next CLRF.
		if !scanner.Scan() {
			log.Println("[processData] Expected to Scan() a string byte; received error: ", scanner.Err())
			return nil
		}
		return scanner.Text()
	default:
		log.Printf("[processData] Unexpected first byte: %v\n", data)
		return nil
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// The data we receive is a command in the form of an array, where the first
	// element is the command and the rest are optional args.
	for {
		command, ok := processData(scanner).(handlers.CommandArgs)
		if !ok {
			break
		}
		conn.Write(handlers.Handle(command))
	}
}
