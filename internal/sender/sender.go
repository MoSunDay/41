package sender

type Sender interface {
	Send([]byte) error
}
