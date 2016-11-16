// Command p4pp is a basic C preprocessor, meant for use with P4 source code.
package main

import (
	"log"
	"os"

	"github.com/mdlayher/p4/preprocessor"
)

func main() {
	log.SetOutput(os.Stderr)

	pp := preprocessor.New(os.Stdin)

	out, err := pp.Process()
	if err != nil {
		log.Fatalf("failed to run preprocessor: %v", err)
	}

	n, err := os.Stdout.Write(out)
	if err != nil {
		log.Fatalf("failed to write output source code: %v", err)
	}
	if n != len(out) {
		log.Fatalf("output was %d bytes, but wrote %d bytes", len(out), n)
	}
}
