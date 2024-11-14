package handler

import (
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type ConfigHandler = Handler

// CONFIG GET parameter [parameter ...]
// The CONFIG GET command is used to read the configuration parameters of a
// running Redis server.
func newConfigHandler(ctx *Ctx) ConfigHandler {
	args := ctx.GetArgs()
	return &configHandler{baseHandler{args: args}}
}

type configHandler struct {
	baseHandler
}

func (c *configHandler) execute() CommandResponse {
	// CONFIG expects at least two arguments
	if !c.argsAtLeast(2) {
		return c.fmtErr("wrong number of arguments for command")
	}
	cmd, args := c.args[0], c.args[1:]
	switch strings.ToUpper(cmd.(string)) {
	case "GET":
		// Answer will be an array of pairs, making its length twice the number
		// of parameters requested
		resp := c.fmtArrayLen(2 * len(args))
		for len(args) > 0 {
			key, ok := args[0].(string)
			if !ok {
				return c.fmtErr("syntax error")
			}
			resp = append(resp, c.fmtBulkString(key)...)
			val, ok := config.Get(key)
			if !ok {
				resp = append(resp, c.fmtNullString()...)
			} else {
				resp = append(resp, c.fmtBulkString(val)...)
			}
			args = args[1:]
		}
		return resp
	default:
		log.Println("[ConfigHandler] Unrecognized command: ", cmd)
		return c.fmtErr("unrecognized command")
	}
}
