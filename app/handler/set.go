package handler

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
)

type SetHandler = Handler

// SET key value [NX | XX] [EX seconds | PX milliseconds ]
//
// Set key to hold the string value. If key already holds a value, it is
// overwritten, regardless of its type.
// Options:
// The SET command supports a (sub)set of options that modify its behavior:
// * EX seconds -- Set the specified expire time, in seconds (a positive integer).
// * PX milliseconds -- Set the specified expire time, in milliseconds (a positive integer).
// * NX -- Only set the key if it does not already exist.
// * XX -- Only set the key if it already exists.
func newSetHandler(ctx *Ctx) SetHandler {
	args := ctx.GetArgs()
	return &setHandler{cache.GetDefaultCache(), baseHandler{args: args}}
}

type setHandler struct {
	cache *cache.Cache
	baseHandler
}

func (s *setHandler) execute() CommandResponse {
	// SET expects at least two arguments
	if !s.argsAtLeast(2) {
		return s.fmtErr("wrong number of arguments for command")
	}
	// Parse key and value
	key, ok := s.args[0].(string)
	if !ok {
		return s.fmtErr("syntax error")
	}
	value, ok := s.args[1].(string)
	if !ok {
		return s.fmtErr("syntax error")
	}
	// Set default expiry
	// time.Time is not nilable in Go, but the zero value works
	// https://pkg.go.dev/time#Time
	expiry := time.Time{}

	// Parse options
	options := s.args[2:]
	for len(options) > 0 {
		opt, rest := options[0], options[1:]
		switch strings.ToUpper(opt.(string)) {
		case "NX": // Only set key if it does not exist
			if s.cache.KeyExists(key) {
				return s.fmtNullString()
			}
			// Set options for next loop
			options = rest
		case "XX": // Only set key if it already exists
			if !s.cache.KeyExists(key) {
				return s.fmtNullString()
			}
			// Set options for next loop
			options = rest
		case "PX": // Set expiry in +PX milliseconds
			if len(rest) == 0 {
				return s.fmtErr("syntax error")
			}
			if !expiry.IsZero() {
				return s.fmtErr("syntax error")
			}
			// Set expiry
			ms, err := strconv.Atoi(rest[0].(string))
			if err != nil {
				log.Println("[SetHandler] Error formatting timeout: ", err)
				return s.fmtErr("syntax error")
			}
			expiry = time.Now().Add(time.Millisecond * time.Duration(ms))
			// Set options for next loop
			options = rest[1:]
		case "EX":
			if len(rest) == 0 {
				return s.fmtErr("syntax error")
			}
			if !expiry.IsZero() {
				return s.fmtErr("syntax error")
			}
			// Set expiry
			sec, err := strconv.Atoi(rest[0].(string))
			if err != nil {
				log.Println("[SetHandler] Error formatting timeout: ", err)
				return s.fmtErr("syntax error")
			}
			expiry = time.Now().Add(time.Second * time.Duration(sec))
			// Set options for next loop
			options = rest[1:]
		default:
			log.Println("[SetHandler] Unrecognized option for SET: ", options)
			// Set options for next loop
			options = rest
		}
	}
	s.cache.Set(key, value, expiry)
	return s.fmtSimpleString("OK")
}
