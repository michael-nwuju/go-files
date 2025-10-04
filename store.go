package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const DefaultRootFolderName string = "files"

func ContentAddressiblePathTransformer(key string) Pathkey {
	hash := sha1.Sum([]byte(key))

	hashString := hex.EncodeToString(hash[:])

	blockSize := 5

	sliceLength := len(hashString) / blockSize

	paths := make([]string, sliceLength)

	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize

		paths[i] = hashString[from:to]
	}

	return Pathkey{
		Filename: hashString,
		Pathname: strings.Join(paths, "/"),
	}
}

type PathTransformer func(string) Pathkey

type Pathkey struct {
	Pathname string

	Filename string
}

func (p Pathkey) FullPathWithoutRoot() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

func (p Pathkey) FullPath(root string) string {
	return fmt.Sprintf("%s/%s/%s", root, p.Pathname, p.Filename)
}

func (p Pathkey) PathnameWithRoot(root string) string {
	return fmt.Sprintf("%s/%s", root, p.Pathname)
}

func (p Pathkey) FirstPathname(root string) string {
	paths := strings.Split(p.Pathname, "/")

	if len(paths) == 0 {
		return ""
	}

	return fmt.Sprintf("%s/%s", root, paths[0])
}

var DefaultPathTransformer = func(key string) Pathkey {
	return Pathkey{
		Pathname: key,
		Filename: key,
	}
}

type StoreOptions struct {
	// Folder name of the root,
	// containing all the files/folders of the system
	Root            string
	PathTransformer PathTransformer
}

type Store struct {
	StoreOptions
}

func NewStore(options StoreOptions) *Store {
	if options.PathTransformer == nil {
		options.PathTransformer = DefaultPathTransformer
	}

	if len(options.Root) == 0 {
		options.Root = DefaultRootFolderName
	}
	return &Store{
		StoreOptions: options,
	}
}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformer(key)

	_, err := os.Stat(pathkey.FullPath(s.Root))

	// return err != fs.ErrNotExist
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformer(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FirstPathname(s.Root))
	}()

	return os.RemoveAll(pathKey.FirstPathname(s.Root))
}

func (s *Store) Read(key string) (io.Reader, error) {
	file, err := s.readStream(key)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, file)

	return buf, err
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathkey := s.PathTransformer(key)

	return os.Open(pathkey.FullPath(s.Root))
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathkey := s.PathTransformer(key)

	if err := os.MkdirAll(pathkey.PathnameWithRoot(s.Root), os.ModePerm); err != nil {
		return err
	}

	fullPath := pathkey.FullPath(s.Root)

	file, err := os.Create(fullPath)

	if err != nil {
		return err
	}

	n, err := io.Copy(file, r)

	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, fullPath)

	return nil
}
