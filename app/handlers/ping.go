package handlers

// PING -- return PONG.
type pingHandler struct {
	baseHandler
}

// PING -- return PONG.
// For this implementation, we ignore arguments and just return PONG. The actual
// implementation accepts an optional message and will return it if given,
// similar to ECHO.
func newPingHandler() *pingHandler {
	return &pingHandler{}
}

func (p *pingHandler) execute() CommandResponse {
	return p.simpleString("PONG")
}
