package p2p

// type Handshaker interface {
// 	Handshake() error
// }

// HandshakeFunc is ...?
type HandshakeFunc func(Peer) error

// type DefaultHandshaker struct {

// }

func NOPHandshakeFunc(Peer) error {
	return nil
}

// // ErrorInvalidHandshake is returned when the handshakes
// // between local & remote nodes could not be established
// var ErrorInvalidHandshake = errors.New("invalid handshake")
