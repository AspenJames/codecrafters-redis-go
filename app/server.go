package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Simplest of k/v stores
var cache map[string]cacheVal = map[string]cacheVal{}

type cacheVal struct {
	value  string
	expiry time.Time
}

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

	writeString := func(str string) {
		conn.Write([]byte(fmt.Sprintf("+%s\r\n", str)))
	}

	writeNullString := func() {
		conn.Write([]byte("$-1\r\n"))
	}

	writeErr := func(errMsg string) {
		conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", errMsg)))
	}

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
				writeErr("wrong number of arguments for command")
				return false
			}
			return true
		}

		// Coerce `cmd` into an uppercase string
		switch strings.ToUpper(fmt.Sprint(cmd)) {
		case "PING":
			writeString("PONG")
		case "ECHO":
			// Echo back message in first arg as simple string
			if isCorrectArgLength(1) {
				writeString(args[0].(string))
			}
		case "SET":
			// Set key=val
			if len(args) < 2 {
				writeErr("wrong number of arguments for command")
				return
			} else {
				key := args[0].(string)
				val := cacheVal{
					value: args[1].(string),
				}

				optVals := args[2:]
				for len(optVals) > 0 {
					opt, rest := optVals[0], optVals[1:]
					switch strings.ToUpper(opt.(string)) {
					case "NX": // Only set key if it does not exist
						if _, ok := cache[key]; ok {
							fmt.Println("NX -- Key exists")
							writeNullString()
							return
						}
						optVals = rest
					case "XX": // Only set key if it already exists
						if _, ok := cache[key]; !ok {
							writeNullString()
							return
						}
						optVals = rest
					case "PX": // Set expiry in +PX milliseconds
						if len(rest) == 0 {
							writeErr("syntax error")
							return
						}
						if !val.expiry.IsZero() {
							fmt.Println("[execute SET PX] Duplicate values provided for expiry")
							writeErr("syntax error")
							return
						}
						// Set expiry
						ms, err := strconv.Atoi(rest[0].(string))
						if err != nil {
							fmt.Println("[execute SET PX] Error formatting timeout: ", err)
							writeErr("syntax error")
							return
						}
						val.expiry = time.Now().Add(time.Millisecond * time.Duration(ms))
						// Set optVals for next loop
						optVals = rest[1:]
					case "EX":
						if len(rest) == 0 {
							fmt.Println("[execute SET EX] No value provided")
							writeErr("syntax error")
							return
						}
						if !val.expiry.IsZero() {
							fmt.Println("[execute SET EX] Duplicate values provided for expiry")
							writeErr("syntax error")
							return
						}
						// Set expiry
						sec, err := strconv.Atoi(rest[0].(string))
						if err != nil {
							fmt.Println("[execute SET EX] Error formatting timeout: ", err)
							writeErr("syntax error")
							return
						}
						val.expiry = time.Now().Add(time.Second * time.Duration(sec))
						// Set optVals for next loop
						optVals = rest[1:]
					default:
						fmt.Println("[execute SET] Unrecognized option for SET: ", optVals)
						optVals = rest
					}
				}
				cache[key] = val
				writeString("OK")
			}
		case "GET":
			// GET key
			if isCorrectArgLength(1) {
				key := args[0].(string)
				val, ok := cache[key]
				if !ok {
					// Not found
					writeNullString()
				} else if !val.expiry.IsZero() && time.Now().After(val.expiry) {
					// Expired
					delete(cache, key)
					writeNullString()
				} else {
					writeString(val.value)
				}
			}
		default:
			fmt.Printf("Unexpected command: '%s' with args: '%v'\n", cmd, args)
			writeErr("unrecognized command")
		}
	}
}
