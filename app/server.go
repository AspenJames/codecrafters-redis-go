package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Simplest of k/v stores
var cache map[string]string = map[string]string{}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
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
			fmt.Println("[processData] Scan() encountered error: ", scanner.Err())
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
			fmt.Println("[processData] Error while processing array length: ", err.Error())
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
			fmt.Println("[processData] Expected to Scan() a string byte; received error: ", scanner.Err())
			return nil
		}
		return scanner.Text()
	default:
		fmt.Printf("[processData] Unexpected first byte: %v\n", data)
		return nil
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// The data we receive is a command in the form of an array, where the first
	// element is the command and the rest are optional args.
	for {
		command, ok := processData(scanner).([]interface{})
		if !ok {
			break
		}
		cmd, args := command[0], command[1:]

		// Checks if `args` is proper length, writes error message if not.
		isCorrectArgLength := func(expectedLen int) bool {
			if len(args) != expectedLen {
				// Return an error
				conn.Write([]byte("-ERR wrong number of arguments for command\r\n"))
				return false
			}
			return true
		}

		// Coerce `cmd` into an uppercase string
		switch strings.ToUpper(fmt.Sprint(cmd)) {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			// Echo back message in first arg as simple string
			if isCorrectArgLength(1) {
				conn.Write([]byte(fmt.Sprintf("+%s\r\n", args[0])))
			}
		case "SET":
			// Set key=val
			if isCorrectArgLength(2) {
				// We're just assuming these type casts work
				key := args[0].(string)
				cache[key] = args[1].(string)
				conn.Write([]byte("+OK\r\n"))
			}
		case "GET":
			// GET key
			if isCorrectArgLength(1) {
				key := args[0].(string)
				val, ok := cache[key]
				if !ok {
					conn.Write([]byte("$-1\r\n"))
				} else {
					conn.Write([]byte(fmt.Sprintf("+%s\r\n", val)))
				}
			}
		default:
			fmt.Printf("Unexpected command: '%s' with args: '%v'\n", cmd, args)
			conn.Write([]byte("-ERR unrecognized command\r\n"))
		}
	}
}
