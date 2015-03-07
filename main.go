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

// You need to implement your fixer and then build it in using build tags.
// Look at fixer_test.go for an example.
type Fixer interface {
	Fix([]byte) ([]byte, error)
}

// ScanNetStrings is a bufio.ScanFunc that can be used with bufio.Scanner to
// Scan() netstrings from a Reader.
// Netstring information here: http://cr.yp.to/proto/netstrings.txt
func ScanNetStrings(data []byte, atEOF bool) (advance int, token []byte, err error) {

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

func main() {

	sockName := os.Getenv("AFL_FIX_SOCK")
	if sockName == "" {
		fmt.Fprintf(os.Stderr, "No socket variable in ENV")
		usage()
		os.Exit(1)
	}

	os.Remove(sockName)
	s, err := net.Listen("unix", sockName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create socket %s: %s", sockName, err)
		usage()
		os.Exit(1)
	}
	log.Printf("Listening on %s...", sockName)
	defer os.Remove(sockName) // not guaranteed to run
	defer s.Close()
	fixer := NewFixer()

	for {
		conn, err := s.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %s", err)
			usage()
			os.Exit(1)
		}
		log.Printf("Accepted connection!")

		scanner := bufio.NewScanner(conn)
		scanner.Split(ScanNetStrings)
		for scanner.Scan() {
			// read a netstring
			in := scanner.Bytes()
			fixed, err := fixer.Fix(in)
			if err != nil {
				log.Printf("WARNING: Error %s from Fixer!", err)
				time.Sleep(1 * time.Second)
				conn.Write(in)
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
