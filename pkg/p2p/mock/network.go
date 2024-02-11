package mock

type Network interface {
	Listen(addr string) (Socket, error)
	Dial(addr string) (Socket, error)
}
