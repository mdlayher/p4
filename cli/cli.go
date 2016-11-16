// Package cli provides subcommands for the p4 tool.
package cli

import (
	"fmt"
	"io"

	"github.com/mdlayher/p4/preprocessor"
)

// An Executor is a type that can execute an action on P4 source code by
// processing an input stream and writing to an output stream.
type Executor interface {
	Execute(w io.Writer, r io.Reader) error
}

// A preprocessorCmd is an Executor that executes a P4 preprocessor.
type preprocessorCmd struct{}

// NewPreprocessor creates an Executor that executes a P4 preprocessor.
func NewPreprocessor() Executor {
	return &preprocessorCmd{}
}

// Execute implements Executor.
func (*preprocessorCmd) Execute(w io.Writer, r io.Reader) error {
	pp := preprocessor.New(r)

	out, err := pp.Process()
	if err != nil {
		return fmt.Errorf("failed to run preprocessor: %v", err)
	}

	n, err := w.Write(out)
	if err != nil {
		return fmt.Errorf("failed to write output source code: %v", err)
	}
	if n != len(out) {
		return fmt.Errorf("output was %d bytes, but wrote %d bytes", len(out), n)
	}

	return nil
}
