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

// Return response as list of []byte to pass to net.Conn.Write().
type CommandResponse = [][]byte

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
	return CommandResponse{[]byte(fmt.Sprintf("*%d\r\n", l))}
}

// fmtBulkString formats `s` as a bulk string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func (b *baseHandler) fmtBulkString(s string) CommandResponse {
	return CommandResponse{[]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))}
}

// fmtErr formats `s` as a simple error.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-errors
func (b *baseHandler) fmtErr(s string) CommandResponse {
	return CommandResponse{[]byte(fmt.Sprintf("-ERR %s\r\n", s))}
}

// fmtSimpleString formats `s` as a simple string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#simple-strings
func (b *baseHandler) fmtSimpleString(s string) CommandResponse {
	return CommandResponse{[]byte(fmt.Sprintf("+%s\r\n", s))}
}

// fmtNullString returns a null bulk string.
// https://redis.io/docs/latest/develop/reference/protocol-spec/#bulk-strings
func (b *baseHandler) fmtNullString() CommandResponse {
	return CommandResponse{[]byte("$-1\r\n")}
}

// Default handler for unrecognized commands
type defaultHandler struct {
	baseHandler
}

func newDefaultHandler(_ *Ctx) Handler {
	return &defaultHandler{}
}

func (d *defaultHandler) execute() CommandResponse {
	return d.fmtErr("unrecognized command")
}

type HandlerFunc = func(*Ctx) Handler

var handlers = map[string]HandlerFunc{
	"CONFIG":   newConfigHandler,
	"ECHO":     newEchoHandler,
	"GET":      newGetHandler,
	"INFO":     newInfoHandler,
	"KEYS":     newKeysHander,
	"PING":     newPingHandler,
	"SET":      newSetHandler,
	"PSYNC":    newPsyncHandler,
	"REPLCONF": newReplconfHandler,
}

var replicatingCmds = []string{
	"SET",
}

func isReplicatingCmd(cmd string) bool {
	for _, c := range replicatingCmds {
		if c == cmd {
			return true
		}
	}
	return false
}

// Main command handler
func Handle(ctx *Ctx) CommandResponse {
	cmd := ctx.GetCmd()
	handler, ok := handlers[strings.ToUpper(fmt.Sprint(cmd))]
	if !ok {
		log.Printf("[Handle] Unexpected command: %q\n", cmd)
		return newDefaultHandler(ctx).execute()
	}
	if isReplicatingCmd(cmd) {
		command := []string{cmd}
		args := ctx.GetArgs()
		for _, a := range args {
			command = append(command, fmt.Sprint(a))
		}
		go notifyReplicas(command)
	}
	return handler(ctx).execute()
}
