package main

import (
	"fmt"
	"log"

	"github.com/michael-nwuju/go-files/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Println("doing some logic with the peer outside of TCPTransport")

	if err := peer.Close(); err != nil {
		return err
	}

	return nil
	// return fmt.Errorf("failed the onpeer func")
}

func main() {
	tcpOptions := p2p.TCPTransportOptions{
		ListenAddr: ":3000",
		Decoder:    p2p.DefaultDecoder{},
		Handshake:  p2p.NOPHandshakeFunc,
		OnPeer:     OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOptions)

	go func() {
		for {
			msg := <-tr.Consume()

			fmt.Printf("Consuming info here - %+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("We up!")

	select {}
}
