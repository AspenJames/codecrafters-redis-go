package handler

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

type ReplicationClient = Client

// ReplicationClient initializes the replica following for a master database
// accessible at `addr`, in the format "host:port".
func NewReplicationClient(addr string) ReplicationClient {
	return &replicationClient{addr, baseHandler{}}
}

type replicationClient struct {
	addr string
	baseHandler
}

// Ping handshake.
func (r *replicationClient) Init() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	respParser := parser.NewRESPParser(bufio.NewScanner(conn))

	sendCmd := func(cmd []string) (string, error) {
		// Build request.
		req := []byte{}
		req = append(req, r.fmtArrayLen(len(cmd))...)
		for _, str := range cmd {
			req = append(req, r.fmtBulkString(str)...)
		}
		// Write request.
		_, err := conn.Write(req)
		if err != nil {
			return "", err
		}
		// Parse, log, and return response
		resp := respParser.Parse()
		if resp == nil {
			log.Printf("[ReplicationClient] No response from %s\n", cmd[0])
			return "", fmt.Errorf("no response from %s", cmd[0])
		}
		str, ok := resp.([]byte)
		if !ok {
			log.Printf("[ReplicationClient] Invalid response from %q: %#v\n", cmd[0], resp)
			return "", fmt.Errorf("invalid respose from %q: %#v", cmd[0], resp)
		}
		log.Printf("[ReplicationClient] Response from %q: %s\n", cmd[0], str)
		return string(str), nil
	}

	// Send PING request
	if _, err := sendCmd([]string{"PING"}); err != nil {
		return err
	}
	// Send REPLCONF listening-port
	listeningPort, _ := config.Get("port")
	if _, err = sendCmd([]string{"REPLCONF", "listening-port", listeningPort}); err != nil {
		return err
	}
	// Send REPLCONF capa psync2
	if _, err = sendCmd([]string{"REPLCONF", "capa", "psync2"}); err != nil {
		return err
	}
	return nil
}
