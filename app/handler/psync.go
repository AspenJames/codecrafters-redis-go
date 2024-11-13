package handler

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type PsyncHandler = Handler

// PSYNC replicationid offset
// The PSYNC command is called by Redis replicas for initiating a replication
// stream from the master. Just reply +FULLRESYNC <REPL_ID> 0\r\n for now.
func newPsyncHandler(args CommandArgs) PsyncHandler {
	return &psyncHandler{args, baseHandler{}}
}

type psyncHandler struct {
	args CommandArgs
	baseHandler
}

func (p *psyncHandler) execute() CommandResponse {
	repl_id, _ := config.Get("master_replid")
	return p.fmtSimpleString(fmt.Sprintf("FULLRESYNC %s 0", repl_id))
}
