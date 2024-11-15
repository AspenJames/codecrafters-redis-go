package parser

import (
	"bufio"
	"log"
	"strconv"
)

// RESPParser parses incoming data on `scanner` as RESP data.
type RESPParser = Parser

type respParser struct {
	scanner *bufio.Scanner
}

// RespParser parses incoming data on `scanner` as RESP data.
func NewRESPParser(scanner *bufio.Scanner) RESPParser {
	return &respParser{scanner}
}

func (r *respParser) Parse() ParseResponse {
	if !r.scanner.Scan() {
		if r.scanner.Err() != nil {
			log.Println("[RESPParser] Scan() encountered error: ", r.scanner.Err())
		}
		return nil
	}

	// Read the first byte to determine how to process the data
	// https://redis.io/docs/latest/develop/reference/protocol-spec/#resp-protocol-description
	///// At this point in the exercise, we only need to handle Array and String
	data := r.scanner.Bytes()
	if len(data) == 0 {
		log.Println("[RESPParser] no bytes received from Scan()")
		return nil
	}
	switch fb := data[0]; fb {
	case byte('*'):
		// Array, data[1:] contains the length
		arrLen, err := strconv.Atoi(string(data[1:]))
		if err != nil {
			log.Println("[RESPParser] Error while processing array length: ", err.Error())
			return nil
		}
		arr := []interface{}{}
		// Read `arrLen` number of elements into `arr`
		for arrLen > 0 {
			arr = append(arr, r.Parse())
			arrLen--
		}
		return arr

	case byte('+'):
		// Simple string, data[1:] contains string
		return data[1:]

	case byte('$'):
		// Bulk string, data[1:] contains length
		// Length is unnecessary, as `r.scanner` is just going to read untiil
		// the next CLRF.
		if !r.scanner.Scan() {
			log.Println("[RESPParser] Expected to Scan() a string byte; received error: ", r.scanner.Err())
			return nil
		}
		return r.scanner.Text()
	default:
		log.Printf("[RESPParser] Unexpected first byte: %#v\n", data)
		return nil
	}
}
