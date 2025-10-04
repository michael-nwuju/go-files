package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/michael-nwuju/go-files/p2p"
)

type FileServerOptions struct {
	StorageRoot string

	PathTransformer PathTransformer

	TCPTransportOptions p2p.TCPTransportOptions

	BootstrapNodes []string
}

type FileServer struct {
	FileServerOptions

	store *Store

	Transport p2p.Transport

	quitChannel chan struct{}

	peerLock sync.Mutex

	peers map[string]p2p.Peer
}

func NewFileServer(options FileServerOptions) *FileServer {
	return &FileServer{
		FileServerOptions: options,
		quitChannel:       make(chan struct{}),
		peers:             map[string]p2p.Peer{},
		Transport:         p2p.NewTCPTransport(options.TCPTransportOptions),
		store:             NewStore(StoreOptions{Root: options.StorageRoot, PathTransformer: options.PathTransformer}),
	}
}

func (s *FileServer) Stop() {
	close(s.quitChannel)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()

	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote peer %s", p.RemoteAddr().String())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("File Server stopped due to user quit action")

		s.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitChannel:
			return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		fmt.Printf("bootstrapping address %s\n", addr)

		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("dial error: %v", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	if len(s.BootstrapNodes) > 0 {
		s.bootstrapNetwork()
	}

	s.loop()

	return nil
}

// func (s *FileServer) Store(key string, r io.Reader) error {
// 	return s.store.Write(key, r)
// }
