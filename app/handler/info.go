package handler

import (
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type InfoHandler = Handler

// INFO [section [section ...]]
// The INFO command returns information and statistics about the server in a
// format that is simple to parse by computers and easy to read by humans.
func newInfoHandler(args CommandArgs) InfoHandler {
	return &infoHandler{baseHandler{args: args}}
}

type infoHandler struct {
	baseHandler
}

func (i *infoHandler) execute() CommandResponse {
	sections := []string{"replication"}
	responseLines := []string{}

	if len(i.args) == 0 {
		// Respond with all sections [that we handle]
		for _, section := range sections {
			i.args = append(i.args, section)
		}
	}

	for _, arg := range i.args {
		section, ok := arg.(string)
		if !ok {
			log.Printf("[InfoHandler] Non-string arg: %#v\n", arg)
			return i.fmtErr("syntax error")
		}
		switch strings.ToLower(section) {
		case "replication":
			responseLines = append(responseLines, "# Replication")
			// Get replica status from config
			replicaof, ok := config.Get("replicaof")
			if !ok {
				log.Println("[InfoHandler] Unexpected error: unable to retrieve 'replicaof' from config")
				return i.fmtErr("unexpected server error")
			}
			if replicaof == "" {
				responseLines = append(responseLines, "role:master")
			} else {
				responseLines = append(responseLines, "role:slave")
			}
		default:
			// Ignore unrecognized section
			log.Printf("[InfoHandler] Unrecognized INFO section %q\n", section)
		}
	}
	resp := ""
	for _, l := range responseLines {
		resp += l + "\n"
	}
	return i.fmtBulkString(resp)
}
