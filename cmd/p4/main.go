// Command p4 is a tool for managing P4 source code.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/mdlayher/p4/cli"
)

// List of subcommands available for this binary.
const (
	cmdPP = "pp"
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	if len(flag.Args()) == 0 {
		fatalf("no arguments specified")
	}

	var exec cli.Executor

	cmd := flag.Arg(0)
	switch cmd {
	case cmdPP:
		exec = cli.NewPreprocessor()
	default:
		fatalf("unknown command: %q", cmd)
	}

	// Read from stdin, write to stdout
	// TODO(mdlayher): perhaps allow a file as input instead of stdin
	w, r := os.Stdout, os.Stdin

	if err := exec.Execute(w, r); err != nil {
		log.Fatal(err)
	}
}

// usage prints a usage message for this binary.
func usage() {
	lf := log.Printf

	lf("p4 is a tool for managing P4 source code.")
	lf("")
	lf("Usage:")
	lf("  p4 command [arguments]")
	lf("")
	lf("Commands:")
	lf("  p4 %s      invoke a preprocessor on P4 source code read from stdin", cmdPP)
	lf("")
}

// fatalf prints the binary's usage and a formatted error message, exiting
// the program.
func fatalf(format string, a ...interface{}) {
	usage()
	log.Fatalf("p4: "+format, a...)
}
