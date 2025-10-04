package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTranspor(t *testing.T) {
	listenAddr := ":4000"

	options := TCPTransportOptions{
		ListenAddr: listenAddr,
		Handshake:  NOPHandshakeFunc,
		Decoder:    DefaultDecoder{},
	}

	tr := NewTCPTransport(options)

	assert.Equal(t, tr.ListenAddr, listenAddr)

	// Server

	assert.Nil(t, tr.ListenAndAccept())

	// select {}
}
