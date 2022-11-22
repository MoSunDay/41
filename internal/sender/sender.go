package sender

import "41/internal/stype"

type Sender interface {
	Send(*stype.RequestResponseRecord) error
}
