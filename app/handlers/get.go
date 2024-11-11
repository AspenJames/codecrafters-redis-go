package handlers

import (
	"log"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
)

type getHandler struct {
	cache *cache.Cache
	baseHandler
}

// GET key
// Get the value of key. If the key does not exist the special value nil is
// returned. An error is returned if the value stored at key is not a string,
// because GET only handles string values.
func newGetHandler(args CommandArgs) *getHandler {
	return &getHandler{cache.GetDefaultCache(), baseHandler{args: args}}
}

func (g *getHandler) execute() CommandResponse {
	// GET expects exactly one argument
	if !g.argsExactly(1) {
		return g.fmtErr("wrong number of arguments for command")
	}
	key, ok := g.args[0].(string)
	if !ok {
		log.Printf("[GetHandler] Non-string key: %#v\n", g.args[0])
		return g.fmtErr("syntax error")
	}
	val, ok := g.cache.Get(key)
	if !ok {
		return g.fmtNullString()
	}
	return g.fmtSimpleString(val)
}
