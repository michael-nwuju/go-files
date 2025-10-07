package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(rpc)
}

type DefaultDecoder struct {
}

func (dec DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	peekBuf := make([]byte, 1)

	if _, err := r.Read(peekBuf); err != nil {
		return err
	}

	// In case of a stream, we are not decoding what is being sent over the network.
	// We are just setting Stream true so we can handle that in a logic
	stream := peekBuf[0] == IncomingStream

	if stream {
		rpc.Stream = true

		return nil
	}

	// // For regular messages, read the payload
	// buf := make([]byte, 4096) // Increased buffer size for gob messages

	buf := make([]byte, 1028)

	n, err := r.Read(buf)

	if err != nil {
		return err
	}

	// fmt.Println(string(buf[:n]))

	rpc.Payload = buf[:n]

	return nil

}
