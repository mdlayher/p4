package preprocessor

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestIncludeLargeFile(t *testing.T) {
	// File twice as big as limit
	file := bytes.Repeat([]byte{0}, maxFileSize*2)

	pp := New(strings.NewReader(`#include "big.p4"`))
	pp.fs = &memoryFilesystem{
		file: file,
	}

	out, err := pp.Process()
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}

	if want, got := maxFileSize, len(out); want != got {
		t.Fatalf("unexpected output file size:\n- want: %d\n-  got: %d", want, got)
	}
}

type memoryFilesystem struct {
	file []byte
}

func (fs *memoryFilesystem) Open(name string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(fs.file)), nil
}
