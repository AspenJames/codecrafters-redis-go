package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/cache"
	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

type ReplicationClient = Client

// ReplicationClient initializes the replica following for a master database
// accessible at `addr`, in the format "host:port".
func NewReplicationClient(addr string) ReplicationClient {
	return &replicationClient{addr, nil, baseHandler{}}
}

type replicationClient struct {
	addr string
	conn net.Conn
	baseHandler
}

// Ping handshake.
func (r *replicationClient) Init() error {
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	r.conn = conn

	respParser := parser.NewRESPParser(bufio.NewScanner(conn))
	readSingleByte := func() (byte, error) {
		buf := make([]byte, 1)
		if _, err := conn.Read(buf); err != nil {
			return byte(0), err
		}
		return buf[0], nil
	}

	sendCmd := func(cmd []string) (string, error) {
		// Build request.
		req := []byte{}
		req = append(req, r.fmtArrayLen(len(cmd))[0]...)
		for _, str := range cmd {
			req = append(req, r.fmtBulkString(str)[0]...)
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
	// Send PSYNC ? -1
	if _, err := sendCmd([]string{"PSYNC", "?", "-1"}); err != nil {
		return err
	}
	// Assume FULLRESYNC, read rdb data
	b, err := readSingleByte()
	if err != nil {
		// Sometimes we don't get data, check for EOF
		if err == io.EOF {
			log.Println("[ReplicationClient] EOF read; no RDB data sent")
			return nil
		}
		return err
	}
	// RDB data is encoded as '$<size>\r\n<content>'.
	if b != byte('$') {
		return fmt.Errorf("unexpected byte flag response from PSYNC: %b", b)
	}
	// Read until \r
	sizeBytes := []byte{}
	for {
		b, err = readSingleByte()
		if err != nil {
			return err
		}
		if b == byte('\r') {
			// Discard next byte, '\n'.
			readSingleByte()
			break
		}
		sizeBytes = append(sizeBytes, b)
	}
	// Parse size; read RDB data.
	size, err := strconv.Atoi(string(sizeBytes))
	if err != nil {
		return err
	}
	rdbBytes := make([]byte, size)
	if _, err := conn.Read(rdbBytes); err != nil {
		return err
	}
	rdbData := parser.NewRDBParser(bytes.NewBuffer(rdbBytes)).Parse()
	// Clear local data and load rdbData
	cache.GetDefaultCache().LoadRDB(rdbData)
	return nil
}

func (r *replicationClient) Handle() {
	defer r.conn.Close()
	scanner := bufio.NewScanner(r.conn)

	for {
		cmdCtx := &Ctx{}
		parsed := parser.NewRESPParser(scanner).Parse()
		// Assert `parsed` is of form CommandArgs
		command, ok := parsed.(CommandArgs)
		if !ok || len(command) == 0 {
			break
		}
		cmdCtx.SetCmd(command[0].(string))
		cmdCtx.SetArgs(command[1:])
		// We don't write responses in replication.
		_ = Handle(cmdCtx)
	}
}
