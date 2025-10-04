package main

import (
	"fmt"
	"log"

	"github.com/michael-nwuju/go-files/p2p"
)

// func OnPeer(peer p2p.Peer) error {
// 	fmt.Println("doing some logic with the peer outside of TCPTransport")

// 	if err := peer.Close(); err != nil {
// 		return err
// 	}

// 	return nil
// 	// return fmt.Errorf("failed the onpeer func")
// }

// func main() {
// 	tcpOptions := p2p.TCPTransportOptions{
// 		ListenAddr: ":3000",
// 		Decoder:    p2p.DefaultDecoder{},
// 		Handshake:  p2p.NOPHandshakeFunc,
// 		OnPeer:     OnPeer,
// 	}

// 	tr := p2p.NewTCPTransport(tcpOptions)

// 	go func() {
// 		for {
// 			msg := <-tr.Consume()

// 			fmt.Printf("Consuming info here - %+v\n", msg)
// 		}
// 	}()

// 	if err := tr.ListenAndAccept(); err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("We up!")

// 	select {}
// }

func makeServer(listenAddr string, nodes []string) *FileServer {
	tcpTransportOptions := p2p.TCPTransportOptions{
		ListenAddr: listenAddr,
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		// #TODO: add on peer function
	}

	fileServerOptions := FileServerOptions{
		BootstrapNodes:      nodes,
		TCPTransportOptions: tcpTransportOptions,
		PathTransformer:     ContentAddressiblePathTransformer,
		StorageRoot:         fmt.Sprintf("%s_network", listenAddr[1:]),
	}

	server := NewFileServer(fileServerOptions)

	tcpTransportOptions.OnPeer = server.OnPeer

	fileServerOptions.TCPTransportOptions = tcpTransportOptions

	server = NewFileServer(fileServerOptions)

	return server
}

func main() {
	s1 := makeServer(":3000", []string{})

	s2 := makeServer(":4000", []string{":3000"})

	go func() {
		if err := s1.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	s2.Start()

	// go func() {
	// 	time.Sleep(time.Second * 3)

	// 	s.Stop()
	// }()

	// if err := s.Start(); err != nil {
	// 	log.Fatal(err)
	// }

	// select {}
}
