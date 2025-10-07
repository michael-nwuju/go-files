package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

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
		peers:             make(map[string]p2p.Peer),
		Transport:         p2p.NewTCPTransport(options.TCPTransportOptions),
		store:             NewStore(StoreOptions{Root: options.StorageRoot, PathTransformer: options.PathTransformer}),
	}
}

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buffer := new(bytes.Buffer)

	if err := gob.NewEncoder(buffer).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})

		if err := peer.Send(buffer.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	// From string

	Payload any
}

type MessageStoreFile struct {
	Key string

	Size int64
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(key)

		return r, err
	}

	fmt.Printf("[%s] don't have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(500 * time.Millisecond)

	for _, peer := range s.peers {
		// First, read the file size so we can limit the amount of bytes that we read
		// from the connection, so it will not keep hanging

		var fileSize int64

		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := s.store.Write(key, io.LimitReader(peer, fileSize))

		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr().String())

		peer.CloseStream()
	}

	_, reader, err := s.store.Read(key)

	return reader, err
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. Store this file to the disk
	// 2. Broadcast this file to all known peers in the network

	var (
		fileBuffer = new(bytes.Buffer)

		tee = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(key, tee)

	if err != nil {
		return err
	}

	message := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&message); err != nil {
		return err
	}

	time.Sleep(5 * time.Millisecond)

	// #TODO: Use a multiwriter here.
	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingStream})

		n, err := io.Copy(peer, fileBuffer)

		if err != nil {
			return err
		}

		fmt.Printf("received and written bytes to disk: %d\n", n)
	}

	return nil
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
		log.Println("File Server stopped due to error or user quit action")

		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var message Message

			// Skip decoding for stream control messages
			if rpc.Stream {
				continue
			}

			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&message); err != nil {
				log.Println("decoding error: ", err)
			}

			if err := s.handleMessage(rpc.From.String(), &message); err != nil {
				log.Println("handle message error: ", err)
			}

		case <-s.quitChannel:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist in disk", s.Transport.Addr(), msg.Key)
	}

	// fmt.Println("need to get a file from disk and send it over the wire")

	fmt.Printf("[%s] serving (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.Key)

	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing ReadCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	// First send the "incomingstream" byte to the peer and then we can send
	// the file size as an int64
	peer.Send([]byte{p2p.IncomingStream})

	binary.Write(peer, binary.LittleEndian, fileSize)

	n, err := io.Copy(peer, r)

	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	// fmt.Printf("received store file %+v\n", msg)

	peer, ok := s.peers[from]

	if !ok {
		return fmt.Errorf("peer (%s) could not be found in peer list", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, (msg.Size)))

	if err != nil {
		return err
	}

	log.Printf("[%s] written (%d) bytes to disk\n", s.Transport.Addr(), n)

	peer.CloseStream()

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		fmt.Printf("bootstrapping address %s\n", addr)

		go func(addr string) {
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("dial error: %v", err)
			}
		}(addr)
	}

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}
