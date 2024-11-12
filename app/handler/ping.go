package handler

type PingHandler = Handler

// PING -- return PONG.
// For this implementation, we ignore arguments and just return PONG. The actual
// implementation accepts an optional message and will return it if given,
// similar to ECHO.
func newPingHandler() PingHandler {
	return &pingHandler{}
}

type pingHandler struct {
	baseHandler
}

func (p *pingHandler) execute() CommandResponse {
	return p.fmtSimpleString("PONG")
}
