package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"time"
)

const MAXLEN = 10 * 1024 * 1024 // 10 MB

// You need to implement this interface in your fixer Look at fixer_simple.go
// for an example.
type Fixer interface {
	Fix([]byte) ([]byte, error)
	Banner() string
	BenchString() string
	TestMap() map[string]string
}

// scanNetStrings is a bufio.ScanFunc that can be used with bufio.Scanner to
// Scan() netstrings from a Reader.
// Netstring information here: http://cr.yp.to/proto/netstrings.txt
func scanNetStrings(data []byte, atEOF bool) (advance int, token []byte, err error) {

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// find out how much to read
	dataLen := 0
	lenLen := bytes.IndexByte(data, ':')
	if lenLen <= 0 {
		err = fmt.Errorf("Missing Length")
		return 0, nil, err
	}

	// No leading zeroes allowed, per spec
	lenStr := string(data[:lenLen])
	if len(lenStr) > 1 && lenStr[0] == '0' {
		err = fmt.Errorf("Leading zero in length")
		return 0, nil, err
	}

	i, err := strconv.ParseInt(lenStr, 10, 0)
	dataLen = int(i) // 0 on error, safe to check this first
	if dataLen > MAXLEN {
		err = fmt.Errorf("Proposed Length > MAXLEN")
		return 0, nil, err
	}
	if err != nil {
		err = fmt.Errorf("Error Parsing Length")
		return 0, nil, err
	}

	// one extra byte for ':'' and ',''
	if len(data) >= lenLen+1+dataLen+1 {
		// possibly complete!
		idxComma := lenLen + 1 + dataLen
		if data[idxComma] != ',' {
			err = fmt.Errorf("Missing Terminator (,)")
			return len(data), nil, err
		}
		// return the data between the : and ,
		return idxComma + 1, data[lenLen+1 : idxComma], nil
	}

	if atEOF {
		err = fmt.Errorf("Unexpected EOF")
		return len(data), nil, err
	}

	// No complete netstring in buffer yet - request more data.
	return 0, nil, nil
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

		scanner := bufio.NewScanner(conn)
		scanner.Split(scanNetStrings)
		for scanner.Scan() {
			// read a netstring
			in := scanner.Bytes()
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

		if scanner.Err() != nil {
			log.Printf("Error during scanning: %s\n", scanner.Err().Error())
		} else {
			log.Printf("Interrupted. Waiting for new connection (^C to abort)\n")
		}
	}
}
