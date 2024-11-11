package handlers

import "log"

// ECHO message -- returns message.
type echoHandler struct {
	baseHandler
}

// ECHO message -- returns message.
func newEchoHandler(args CommandArgs) *echoHandler {
	return &echoHandler{baseHandler{args: args}}
}

func (e *echoHandler) execute() CommandResponse {
	// ECHO expects exactly one argument
	if !e.argsExactly(1) {
		return e.err("wrong number of arguments for command")
	}
	// Respond with argument as simple string
	msg, ok := e.args[0].(string)
	if !ok {
		log.Printf("[EchoHandler] Non-string argument: %#v\n", e.args[0])
		return e.err("syntax error")
	}
	return e.simpleString(msg)
}
