package handler

import "net"

// Ctx is the context for a handler to carry the current command, args, and
// client address.
type Ctx struct {
	args       CommandArgs
	clientAddr string
	cmd        string
	conn       net.Conn
}

// Getters

func (c *Ctx) GetArgs() CommandArgs {
	return c.args
}

func (c *Ctx) GetCmd() string {
	return c.cmd
}

func (c *Ctx) GetClientAddr() string {
	return c.clientAddr
}

func (c *Ctx) GetConn() net.Conn {
	return c.conn
}

// Setters

func (c *Ctx) SetArgs(args CommandArgs) {
	c.args = args
}

func (c *Ctx) SetCmd(cmd string) {
	c.cmd = cmd
}

func (c *Ctx) SetClientAddr(clientAddr string) {
	c.clientAddr = clientAddr
}

func (c *Ctx) SetConn(conn net.Conn) {
	c.conn = conn
}
