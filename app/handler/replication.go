package handler

import "net"

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
	pingReq := []byte{}
	pingReq = append(pingReq, r.fmtArrayLen(1)...)
	pingReq = append(pingReq, r.fmtBulkString("PING")...)
	if _, err = conn.Write(pingReq); err != nil {
		return err
	}
	return nil
}
