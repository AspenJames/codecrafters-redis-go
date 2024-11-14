package handler

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/app/config"
)

type PsyncHandler = Handler

// PSYNC replicationid offset
// The PSYNC command is called by Redis replicas for initiating a replication
// stream from the master. Just reply +FULLRESYNC <REPL_ID> 0\r\n for now.
func newPsyncHandler(ctx *Ctx) PsyncHandler {
	args := ctx.GetArgs()
	return &psyncHandler{baseHandler{args: args}}
}

type psyncHandler struct {
	baseHandler
}

// RDB file with no data, hex encoded.
const emptyRDBhex = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (p *psyncHandler) execute() CommandResponse {
	resp := make(CommandResponse, 2)
	// Send FULLRESYNC cmd response
	repl_id, _ := config.Get("master_replid")
	resp = append(resp, p.fmtSimpleString(fmt.Sprintf("FULLRESYNC %s 0", repl_id))...)
	// Send rdb file response
	// Attempt to read configured RDB file; fallback to empty
	dir, _ := config.Get("dir")
	dbfilename, _ := config.Get("dbfilename")
	dbfilepath := filepath.Join(dir, dbfilename)
	dbFile, err := os.Open(dbfilepath)
	var filebytes []byte
	if os.IsNotExist(err) {
		// Fallback
		filebytes, err = hex.DecodeString(emptyRDBhex)
		if err != nil {
			log.Fatal("Error decoding empty RDB file: ", err.Error())
		}
	} else if err != nil {
		log.Printf("[PsyncHandler] Error opening file %s: %s\n", dbfilepath, err)
		return p.fmtErr("Unexpected server error")
	} else {
		defer dbFile.Close()
		filebytes, err = io.ReadAll(dbFile)
		if err != nil {
			log.Printf("[PsyncHandler] Error reading file %s: %s\n", dbfilepath, err)
			return p.fmtErr("Unexpected server error")
		}
	}
	filedata := fmt.Sprintf("$%d\r\n%s", len(filebytes), filebytes)
	resp = append(resp, CommandResponse{[]byte(filedata)}...)
	return resp
}
