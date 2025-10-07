package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/michael-nwuju/go-files/p2p"
)

func makeServer(listenAddr string, nodes []string) *FileServer {
	fileServerOptions := FileServerOptions{
		StorageRoot:     fmt.Sprintf("%s_network", listenAddr[1:]),
		PathTransformer: ContentAddressiblePathTransformer,
		BootstrapNodes:  nodes,
	}

	server := NewFileServer(fileServerOptions)

	tcpTransportOptions := p2p.TCPTransportOptions{
		ListenAddr: listenAddr,
		Handshake:  p2p.NOPHandshakeFunc,
		Decoder:    p2p.DefaultDecoder{},
		OnPeer:     server.OnPeer,
	}

	server.TCPTransportOptions = tcpTransportOptions
	server.Transport = p2p.NewTCPTransport(tcpTransportOptions)

	return server
}

func main() {
	s1 := makeServer(":3000", []string{})

	s2 := makeServer(":4000", []string{":3000"})

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go s2.Start()

	time.Sleep(2 * time.Second)

	// for i := 0; i < 10; i++ {
	// 	data := bytes.NewReader([]byte("my big data file here"))

	// 	s2.Store(fmt.Sprintf("myprivatedata_%d", i), data)

	// 	time.Sleep(5 * time.Millisecond)
	// }

	// data := bytes.NewReader([]byte("my big data file here"))

	// s2.Store("coolPicture.jpg", data)

	// time.Sleep(5 * time.Millisecond)

	r, err := s2.Get("coolPicture.jpg")

	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)

	if err != nil {
		log.Fatal(err)
	}

	println(string(b))
}
