package main

// To get tests to work properly you will need to `go build -tags [FIXERNAME]`
// because the test code runs the aflfix server out of the current directory,
// which is not modified by the test invocation ( so it needs to be explicitly
// rebuilt including the desired fixer )

// Fixers must define the "tests" map as well as the Fix() method etc. See
// fix_simple.go for an example.

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

	s, err := sleepyConnect(sock)
	if err != nil {
		t.Fatalf("Failed to connect to fix server")
	}
	defer server.Process.Kill()
	defer s.Close()

	scanner := bufio.NewScanner(s)
	scanner.Split(ScanNetStrings)

	for k, want := range tests {
		s.Write([]byte(fmt.Sprintf("%d:%s,", len(k), k)))
		scanner.Scan()
		if scanner.Text() != want {
			t.Fatalf("Sent %s, got %s, want %s", k, scanner.Text(), want)
		}
	}
}
