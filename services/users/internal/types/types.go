package types

type IDatabase interface {
	Register(username, password string) error
	Login(username, password string) error
}

type IBroker interface {
	Send(message string) error
	Close()
}
