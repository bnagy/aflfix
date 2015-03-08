package main

// Fixers must define the "tests" map, the "bench" string varible as well as
// the Fix() method etc. See fix_simple.go for an example.

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"testing"
	"time"
)

func sleepyConnect(dest string) (s net.Conn, err error) {
	zzz := 1 * time.Millisecond
	for {
		if zzz > 1*time.Second {
			err = fmt.Errorf("Failed to connect to fix server")
			return
		}
		time.Sleep(zzz)
		s, err = net.Dial("unix", dest)
		if err == nil {
			return
		}
		zzz *= 2
	}
}

func BenchmarkFixup(b *testing.B) {

	sock := path.Join(os.TempDir(), "aflfix.sock")
	srv := server{NewFixer()}
	go srv.Run(sock)

	s, err := sleepyConnect(sock)
	if err != nil {
		b.Fatalf("Failed to connect to fix server")
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)
	scanner.Split(scanNetStrings)
	out := []byte(fmt.Sprintf("%d:%s,", len(srv.BenchString()), srv.BenchString()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Write(out)
		scanner.Scan()
	}
}

func TestRoundTrip(t *testing.T) {

	sock := path.Join(os.TempDir(), "aflfix.sock")
	srv := server{NewFixer()}
	go srv.Run(sock)

	s, err := sleepyConnect(sock)
	if err != nil {
		t.Fatalf("Failed to connect to fix server")
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)
	scanner.Split(scanNetStrings)

	for k, want := range srv.TestMap() {
		s.Write([]byte(fmt.Sprintf("%d:%s,", len(k), k)))
		scanner.Scan()
		if scanner.Text() != want {
			t.Fatalf("Sent %s, got %s, want %s", k, scanner.Text(), want)
		}
	}

}
