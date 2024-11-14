package handler

import (
	"fmt"
	"log"
	"net"
)

type ReplconfHandler = Handler

// REPLCONF
// Internal command used to configure replication. Just reply OK for now
func newReplconfHandler(ctx *Ctx) ReplconfHandler {
	args := ctx.GetArgs()
	clientAddr := ctx.GetClientAddr()
	conn := ctx.GetConn()
	return &replconfHandler{clientAddr, conn, baseHandler{args: args}}
}

type replconfHandler struct {
	clientAddr string
	conn       net.Conn
	baseHandler
}

func (r *replconfHandler) execute() CommandResponse {
	for len(r.args) > 0 {
		// We should be getting pairs of values.
		if len(r.args) < 2 {
			log.Printf("[ReplconfHandler] Oh no! Incorrect # of args -- %#v\n", r.args)
		}
		switch r.args[0] {
		case "listening-port":
			port := r.args[1]
			addr := fmt.Sprintf("%s:%s", r.clientAddr, port)
			registerReplica(addr, r.conn)
		default:
			log.Printf("[ReplconfHandler] %q => %q\n", r.args[0], r.args[1])
		}
		r.args = r.args[2:]
	}
	return r.fmtSimpleString("OK")
}
