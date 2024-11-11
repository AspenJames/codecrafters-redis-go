package handlers

import (
	"fmt"
	"log"
	"strings"
)

// baseHandler holds args, validation logic, and formatting.
// All handlers should inhert from baseHandler.
//
// A handler is a type that includes baseHandler as an anonymous field and
// implements `execute() CommandResponse`
type baseHandler struct {
	args CommandArgs
}

// Array of arguments of any type.
type CommandArgs = []interface{}

// Return response as []byte to pass to net.Conn.Write().
type CommandResponse = []byte

// Utils
/// [Utils] Validation

// argsExactly returns true if there are at least `n` args.
func (b *baseHandler) argsAtLeast(n int) bool {
	return len(b.args) >= n
}

// argsExactly returns true if there are exactly `n` args.
func (b *baseHandler) argsExactly(n int) bool {
	return len(b.args) == n
}

/// [Utils] Formatting

// err formats `s` as a simple error.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors
func (b *baseHandler) err(s string) CommandResponse {
	return []byte(fmt.Sprintf("-ERR %s\r\n", s))
}

// simpleString formats `s` as a simple string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings
func (b *baseHandler) simpleString(s string) CommandResponse {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

// nullString returns a null bulk string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func (b *baseHandler) nullString() CommandResponse {
	return []byte("$-1\r\n")
}

// Default handler for unrecognized commands
type defaultHandler struct {
	baseHandler
}

func newDefaultHandler() *defaultHandler {
	return &defaultHandler{}
}

func (d *defaultHandler) execute() CommandResponse {
	return d.err("unrecognized command")
}

// Main command handler
func Handle(command CommandArgs) CommandResponse {
	cmd, args := command[0], command[1:]
	switch strings.ToUpper(fmt.Sprint(cmd)) {
	case "PING":
		return newPingHandler().execute()
	case "ECHO":
		return newEchoHandler(args).execute()
	case "SET":
		return newSetHandler(args).execute()
	case "GET":
		return newGetHandler(args).execute()
	case "CONFIG":
		return newConfigHandler(args).execute()
	default:
		log.Printf("[Handle] Unexpected command: '%s' with args: '%v'\n", cmd, args)
		return newDefaultHandler().execute()
	}
}
