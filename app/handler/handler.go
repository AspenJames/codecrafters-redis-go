package handler

import (
	"fmt"
	"log"
	"strings"
)

type Handler interface {
	execute() CommandResponse
}

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

// fmtArrayLen formats an array of length `l`
// https://redis.io/docs/latest/develop/reference/protocol-spec/#arrays
func (b *baseHandler) fmtArrayLen(l int) CommandResponse {
	return []byte(fmt.Sprintf("*%d\r\n", l))
}

// fmtBulkString formats `s` as a bulk string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func (b *baseHandler) fmtBulkString(s string) CommandResponse {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

// fmtErr formats `s` as a simple error.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors
func (b *baseHandler) fmtErr(s string) CommandResponse {
	return []byte(fmt.Sprintf("-ERR %s\r\n", s))
}

// fmtSimpleString formats `s` as a simple string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings
func (b *baseHandler) fmtSimpleString(s string) CommandResponse {
	return []byte(fmt.Sprintf("+%s\r\n", s))
}

// fmtNullString returns a null bulk string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func (b *baseHandler) fmtNullString() CommandResponse {
	return []byte("$-1\r\n")
}

// Default handler for unrecognized commands
type defaultHandler struct {
	baseHandler
}

func newDefaultHandler(_ CommandArgs) Handler {
	return &defaultHandler{}
}

func (d *defaultHandler) execute() CommandResponse {
	return d.fmtErr("unrecognized command")
}

type HandlerFunc = func(CommandArgs) Handler

// Main command handler
func Handle(command CommandArgs) CommandResponse {
	handlers := map[string]HandlerFunc{
		"CONFIG": newConfigHandler,
		"ECHO":   newEchoHandler,
		"GET":    newGetHandler,
		"KEYS":   newKeysHander,
		"PING":   newPingHandler,
		"SET":    newSetHandler,
	}
	cmd, args := command[0], command[1:]
	handler, ok := handlers[strings.ToUpper(fmt.Sprint(cmd))]
	if !ok {
		log.Printf("[Handle] Unexpected command: '%s' with args: '%v'\n", cmd, args)
		return newDefaultHandler(args).execute()
	}
	return handler(args).execute()
}