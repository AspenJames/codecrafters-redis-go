package handler

// Client is a public interface, used when we need to connect out.
type Client interface {
	Init() error
	Handle()
}
