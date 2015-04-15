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

	r := bufio.NewReader(s)
	out := []byte(fmt.Sprintf("%d:%s,", len(srv.BenchString()), srv.BenchString()))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Write(out)
		readNetString(r)
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

	r := bufio.NewReader(s)

	for k, want := range srv.TestMap() {
		s.Write([]byte(fmt.Sprintf("%d:%s,", len(k), k)))
		in, err := readNetString(r)
		if string(in) != want {
			t.Fatalf("Sent %s, got %s, want %s", k, string(in), want)
		}
		if err != nil {
			t.Fatalf("Unexpected error %s", err)
		}
	}

}
