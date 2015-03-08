package main

import (
	"fmt"
	"os"
)

func main() {

	sockName := os.Getenv("AFL_FIX_SOCK")
	if sockName == "" {
		fmt.Fprintf(os.Stderr, "No socket variable in ENV")
		usage()
		os.Exit(1)
	}

	srv := server{NewFixer()}
	err := srv.Run(sockName)
	fmt.Fprintf(os.Stderr, "Server exited: %s", err)

}
