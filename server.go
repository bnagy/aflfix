package main

import (
	"bufio"
	//"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

const MAXLEN = 10 * 1024 * 1024 // 10 MB

// You need to implement this interface in your fixer. Look at fixer_simple.go
// for an example.
type Fixer interface {
	Fix([]byte) ([]byte, error)
	Banner() string
	BenchString() string
	TestMap() map[string]string
}

var data = make([]byte, MAXLEN, MAXLEN)

func readNetString(r *bufio.Reader) ([]byte, error) {

	lenBytes, err := r.ReadBytes(':')
	if err != nil {
		return []byte{}, err
	}
	i, err := strconv.ParseInt(string(lenBytes[:len(lenBytes)-1]), 10, 0)
	dataLen := int(i) // 0 on error, safe to check this first
	if dataLen > MAXLEN {
		err = fmt.Errorf("Proposed Length > MAXLEN")
		return []byte{}, err
	}
	if err != nil {
		err = fmt.Errorf("Error Parsing Length %#v", lenBytes)
		return []byte{}, err
	}

	data = data[:dataLen]
	_, err = io.ReadFull(r, data)
	if err != nil {
		err = fmt.Errorf("Error reading data")
		return []byte{}, err
	}
	b, err := r.ReadByte()
	if err != nil || b != ',' {
		err = fmt.Errorf("Missing terminator")
		return []byte{}, err
	}

	return data, nil
}

func usage() {
	fmt.Fprintf(
		os.Stderr,
		"  Usage: AFL_FIX_SOCK=/path/to/sockfile %s\n",
		path.Base(os.Args[0]),
	)
}

type server struct{ Fixer }

func (srv *server) Run(sockName string) error {

	os.Remove(sockName)
	s, err := net.Listen("unix", sockName)
	if err != nil {
		usage()
		return err
	}
	log.Printf("Listening on %s...", sockName)
	defer os.Remove(sockName) // not guaranteed to run
	defer s.Close()
	log.Printf("Starting up. Running Fixer: %s", srv.Banner())

	for {
		conn, err := s.Accept()
		if err != nil {
			usage()
			return err
		}
		log.Printf("Accepted connection!")

		r := bufio.NewReader(conn)
		for {

			in, err := readNetString(r)
			if err != nil {
				log.Printf("Error during scanning: %s\n", err)
				log.Printf("Disconnected.")
				conn.Close()
				break
			}

			fixed, err := srv.Fix(in)
			if err != nil {
				log.Printf("WARNING: Error %s from Fixer!", err)
				// Don't spam the user with 100+ warnings per second. They're
				// not going to want to run for long if their fixer is broken.
				time.Sleep(1 * time.Second)
				conn.Write(in)
				continue
			}
			// write a netstring
			conn.Write([]byte(fmt.Sprintf("%d:%s,", len(fixed), string(fixed))))

		}
	}
}
