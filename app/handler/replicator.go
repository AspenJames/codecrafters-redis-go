package handler

import (
	"log"
	"net"
)

// Map of address => conn
var replicas = make(map[string]net.Conn)

// For access to formatting helpers.
var b = &baseHandler{}

func registerReplica(addr string, conn net.Conn) {
	replicas[addr] = conn
}

// Not currently used.
// func dregisterReplica(addr string) {
// 	delete(replicas, addr)
// }

func notifyReplicas(cmd []string) {
	command := []byte{}
	command = append(command, b.fmtArrayLen(len(cmd))[0]...)
	for _, el := range cmd {
		command = append(command, b.fmtBulkString(el)[0]...)
	}

	for addr, conn := range replicas {
		if _, err := conn.Write(command); err != nil {
			log.Printf("[notifyReplicas] Error response from %q: %s\n", addr, err)
		}
	}
}
