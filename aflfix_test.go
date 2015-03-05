package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

// change these!
var tests = map[string]string{
	"Blah Hello World": "Blah A MUCH LONGER THING World",
	"Blah":             "Blah",
	"Blah\xff\xfe\xaa\x00\x00Hello World": "Blah\xff\xfe\xaa\x00\x00A MUCH LONGER THING World",
	"": "",
}

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
	os.Setenv("AFL_FIX_SOCK", sock)
	server := exec.Command("./aflfix")
	err := server.Start()
	if err != nil {
		b.Fatalf("Unable to launch server: %s", err)
	}
	// When we exit the sock will EOF and the server will exit cleanly.
	s, err := sleepyConnect(sock)
	if err != nil {
		b.Fatalf("Failed to connect to fix server")
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)
	scanner.Split(ScanNetStrings)
	k := "Blah\xff\xfe\xaa\x00\x00Hello World"
	out := []byte(fmt.Sprintf("%d:%s,", len(k), k))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Write(out)
		scanner.Scan()
	}
}

func TestRoundTrip(t *testing.T) {

	sock := path.Join(os.TempDir(), "aflfix.sock")
	os.Setenv("AFL_FIX_SOCK", sock)
	server := exec.Command("./aflfix")
	err := server.Start()
	if err != nil {
		t.Fatalf("Unable to launch server: %s", err)
	}
	// Don't defer Kill() - when we exit the sock will EOF at the server end
	// and it will exit cleanly.
	s, err := sleepyConnect(sock)
	if err != nil {
		t.Fatalf("Failed to connect to fix server")
	}
	defer s.Close()

	scanner := bufio.NewScanner(s)
	scanner.Split(ScanNetStrings)

	// Change the tests if you change the core logic, obviously.
	for k, want := range tests {
		s.Write([]byte(fmt.Sprintf("%d:%s,", len(k), k)))
		scanner.Scan()
		if scanner.Text() != want {
			t.Fatalf("Sent %s, got %s, want %s", k, scanner.Text(), want)
		}
	}
}
