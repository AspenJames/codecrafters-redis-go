package handlers

import (
	"fmt"
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
		return e.err("wrong number of arguments for command")
	}
	cmd, args := e.args[0], e.args[1:]
	switch strings.ToUpper(cmd.(string)) {
	case "GET":
		// Answer will be an array of pairs, making its length twice the number
		// of parameters requested
		resp := fmt.Sprintf("*%d\r\n", 2*len(args))
		for len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return e.err("syntax error")
			}
			resp += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
			val, ok := config.Get(key)
			if !ok {
				resp += string(e.nullString())
			} else {
				resp += fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
			}
			args = args[1:]
		}
		return []byte(resp)
	default:
		log.Println("[ConfigHandler] Unrecognized command: ", cmd)
		return e.err("unrecognized command")
	}
}
