package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/michael-nwuju/go-files/p2p"
)

func makeServer(StoreID, listenAddr string, nodes []string) *FileServer {
	fileServerOptions := FileServerOptions{
		StorageRoot:     fmt.Sprintf("%s_network", listenAddr[1:]),
		PathTransformer: ContentAddressiblePathTransformer,
		BootstrapNodes:  nodes,
		EncryptionKey:   newEncryptionKey(),
		StoreID:         StoreID,
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
	storeID := generateID()

	s1 := makeServer(storeID, ":3000", []string{})

	s2 := makeServer(storeID, ":4000", []string{":3000"})

	s3 := makeServer(storeID, ":5001", []string{":3000", ":4000"})

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)

	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(1 * time.Second)

	go func() {
		log.Fatal(s3.Start())
	}()

	time.Sleep(2 * time.Second)

	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("picture_%d.png", i)

		data := bytes.NewReader([]byte("my big data file here!"))

		s3.Store(key, data)

		if err := s3.store.Delete(key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)

		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))

		time.Sleep(5 * time.Millisecond)
	}

	fmt.Println("")
	fmt.Println("Sleeping before deleting file")

	time.Sleep(2 * time.Second)

	fmt.Println("")

	fmt.Println("Deleting middle file")

	s3.Delete("picture_1.png")

}
