package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// filename => clown.jpg
// path => transform(key) => ROOT/[id]/path

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

func (p Pathkey) FullPath(root string, id string) string {
	return fmt.Sprintf("%s/%s/%s/%s", root, id, p.Pathname, p.Filename)
}

func (p Pathkey) PathnameWithRoot(root string, id string) string {
	return fmt.Sprintf("%s/%s/%s", root, id, p.Pathname)
}

func (p Pathkey) FirstPathname(root string, id string) string {
	paths := strings.Split(p.Pathname, "/")

	if len(paths) == 0 {
		return ""
	}

	return fmt.Sprintf("%s/%s/%s", root, id, paths[0])
}

var DefaultPathTransformer = func(key string) Pathkey {
	return Pathkey{
		Pathname: key,
		Filename: key,
	}
}

type StoreOptions struct {
	// Folder name of the root containing all the files/folders of the system
	Root string

	PathTransformer PathTransformer

	// ID of the owner of the storage, which will be used to store all files in the location
	// so we can sync all the files needed
	ID string
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

	if len(options.ID) == 0 {
		options.ID = generateID()
	}

	return &Store{
		StoreOptions: options,
	}
}

func (s *Store) Has(key string) bool {
	pathkey := s.PathTransformer(key)

	// println(pathkey.FullPath(s.Root, s.ID))

	// println(s.ID)

	_, err := os.Stat(pathkey.FullPath(s.Root, s.ID))

	// return err != fs.ErrNotExist
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformer(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FirstPathname(s.Root, s.ID))
	}()

	return os.RemoveAll(pathKey.FirstPathname(s.Root, s.ID))
}

func (s *Store) Read(key string) (int64, io.Reader, error) {
	return s.readStream(key)
}

func (s *Store) readStream(key string) (int64, io.ReadCloser, error) {
	pathkey := s.PathTransformer(key)

	file, err := os.Open(pathkey.FullPath(s.Root, s.ID))

	if err != nil {
		return 0, nil, err
	}

	stat, err := file.Stat()

	if err != nil {
		return 0, nil, err
	}

	return stat.Size(), file, nil

}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) WriteDecrypt(encryptionKey []byte, key string, r io.Reader) (int64, error) {
	file, err := s.openFileForWriting(key)

	if err != nil {
		return 0, err
	}

	n, err := copyDecrypt(encryptionKey, r, file)

	return int64(n), err
}

func (s *Store) openFileForWriting(key string) (*os.File, error) {
	pathkey := s.PathTransformer(key)

	if err := os.MkdirAll(pathkey.PathnameWithRoot(s.Root, s.ID), os.ModePerm); err != nil {
		return nil, err
	}

	fullPath := pathkey.FullPath(s.Root, s.ID)

	return os.Create(fullPath)
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {

	file, err := s.openFileForWriting(key)

	if err != nil {
		return 0, err
	}

	return io.Copy(file, r)
}
