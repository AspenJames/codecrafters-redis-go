package handler

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
)

type KeysHander = Handler

// KEYS pattern
// Returns all keys matching pattern.
func newKeysHander(args CommandArgs) KeysHander {
	return &keysHander{cache.GetDefaultCache(), baseHandler{args: args}}
}

type keysHander struct {
	cache *cache.Cache
	baseHandler
}

func (k *keysHander) execute() CommandResponse {
	// KEYS expects exactly one argument
	if !k.argsExactly(1) {
		return k.fmtErr("wrong number of arguments for command")
	}
	search, ok := k.args[0].(string)
	if !ok {
		return k.fmtErr("syntax error")
	}

	// Convert basic search pattern to regexp syntax.
	// Yes, this is silly and potentially buggy.
	// * -> .* none or any characters.
	updatedSearch := strings.ReplaceAll(search, "*", ".*")
	// ? -> \w single character.
	updatedSearch = strings.ReplaceAll(updatedSearch, "?", "\\w")
	// Add terminators to the search.
	updatedSearch = fmt.Sprintf("^%s$", updatedSearch)
	pattern, err := regexp.Compile(updatedSearch)
	if err != nil {
		log.Printf("[KeysHandler] Error compiling regexp for pattern %q: %s\n", updatedSearch, err)
		return k.fmtErr("syntax error")
	}

	keys := k.cache.GetKeys(pattern)
	resp := k.fmtArrayLen(len(keys))
	for _, key := range keys {
		resp = append(resp, k.fmtBulkString(key)...)
	}
	return resp
}
