package p2p

import (
	"fmt"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn net.Conn

	// if we dial and retrieve a connection => outbound == true
	// if we accept and retrieve a connection => outbound == false
	outbound bool
}

type TCPTransportOptions struct {
	ListenAddr string

	Handshake HandshakeFunc

	Decoder Decoder

	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportOptions

	listener net.Listener

	rpcChannel chan RPC

	// mu sync.RWMutex

	// peers map[net.Addr]Peer
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// Close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func NewTCPTransport(options TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		TCPTransportOptions: options,
		rpcChannel:          make(chan RPC),
	}
}

// Consume implements the Transport Interface
// which will return the read-only channel
// for reading the incoming messages received from another peer in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChannel
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.ListenAddr)

	if err != nil {
		return err
	}

	t.listener = ln

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	var err error

	defer func() {
		fmt.Printf("Dropping Peer connection: %s\n", err)
		conn.Close()
	}()

	if err = t.Handshake(peer); err != nil {
		conn.Close()

		fmt.Printf("TCP Handshake error: %s\n", err)

		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	fmt.Printf("new incoming connection %+v\n", peer)

	rpc := RPC{}

	// Read loop
	for {
		err = t.Decoder.Decode(conn, &rpc)

		if err != nil {
			fmt.Printf("TCP Read error: %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()

		t.rpcChannel <- rpc

		fmt.Printf("message: %+v\n", rpc)
	}
}
