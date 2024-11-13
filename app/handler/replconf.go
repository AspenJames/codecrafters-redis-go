package handler

type ReplconfHandler = Handler

// REPLCONF
// Internal command used to configure replication. Just reply OK for now
func newReplconfHandler(_ CommandArgs) ReplconfHandler {
	return &replconfHandler{}
}

type replconfHandler struct {
	baseHandler
}

func (p *replconfHandler) execute() CommandResponse {
	return p.fmtSimpleString("OK")
}
