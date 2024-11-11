package handlers

import (
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

// CONFIG GET parameter [parameter ...]
// The CONFIG GET command is used to read the configuration parameters of a
// running Redis server.
type configHandler struct {
	baseHandler
}

// CONFIG GET parameter [parameter ...]
// The CONFIG GET command is used to read the configuration parameters of a
// running Redis server.
func newConfigHandler(args CommandArgs) *configHandler {
	return &configHandler{baseHandler{args: args}}
}

func (e *configHandler) execute() CommandResponse {
	// CONFIG expects at least two arguments
	if !e.argsAtLeast(2) {
		return e.fmtErr("wrong number of arguments for command")
	}
	cmd, args := e.args[0], e.args[1:]
	switch strings.ToUpper(cmd.(string)) {
	case "GET":
		// Answer will be an array of pairs, making its length twice the number
		// of parameters requested
		resp := e.fmtArrayLen(2 * len(args))
		for len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return e.fmtErr("syntax error")
			}
			resp = append(resp, e.fmtBulkString(key)...)
			val, ok := config.Get(key)
			if !ok {
				resp = append(resp, e.fmtNullString()...)
			} else {
				resp = append(resp, e.fmtBulkString(val)...)
			}
			args = args[1:]
		}
		return resp
	default:
		log.Println("[ConfigHandler] Unrecognized command: ", cmd)
		return e.fmtErr("unrecognized command")
	}
}
