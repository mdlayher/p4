// Package preprocessor provides a basic source code preprocessor, emulating
// the C preprocessor.
package preprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// A Preprocessor processes directives in source code to perform certain
// actions before any other processing stages occur.
type Preprocessor struct {
	// Define is a function which indicates how "#define" directives should
	// be handled.
	//
	// This function will be invoked with the name of a preprocessor constant
	// and its value.  The appropriate name and value should be returned to
	// register them with the Preprocessor.  If an error is returned by Define,
	// Preprocessor.Process is halted and the error is returned.
	//
	// If nil, the name and value are left unchanged and registered with the
	// Preprocessor.
	Define func(name string, value string) (outName string, outValue string, err error)

	// Include is a function which indicates how "#include" directives should
	// be handled.
	//
	// This function will be invoked with the name of a source file for
	// inclusion in the code.  The source code of the file should be returned
	// to replace the include directive with the source code.  If an error is
	// returned by Include, Preprocessor.Process is halted and the error is
	// returned.
	//
	// If nil, name attempts to open a file in the filesystem, returning its
	// contents for inclusion in the output source code.
	Include func(name string) (src []byte, err error)

	s       *bufio.Scanner
	defines map[string]string
}

// New creates a new Preprocessor using the input io.Reader.
func New(r io.Reader) *Preprocessor {
	return &Preprocessor{
		s:       bufio.NewScanner(r),
		defines: make(map[string]string, 0),
	}
}

var (
	bufDefine  = []byte("#define")
	bufInclude = []byte("#include")
)

// Process invokes the Preprocessor, handling directives as necessary, and
// returning the modified source code.
func (p *Preprocessor) Process() ([]byte, error) {
	var src []byte
	for p.s.Scan() {
		b := p.s.Bytes()
		switch {
		case bytes.HasPrefix(b, bufDefine):
			// Syntax:
			//   - #define FOO_BITS 8
			//   - #define FOO_BAR "foo bar"
			f := bytes.Fields(b)
			if len(f) < 3 {
				return nil, fmt.Errorf("invalid define preprocessor directive: %q", string(b))
			}

			name := string(f[1])
			// TODO(mdlayher): don't mangle spaces in define value
			value := string(bytes.Join(f[2:], []byte(" ")))

			// Apply any nested definitions before defining the new value
			for k, v := range p.defines {
				value = strings.Replace(value, k, v, -1)
			}

			define := p.Define
			if define == nil {
				define = defineNoChanges
			}

			outName, outValue, err := define(name, value)
			if err != nil {
				return nil, fmt.Errorf("preprocessor error while defining %q as %q: %v", outName, outValue, err)
			}

			p.defines[outName] = outValue
		case bytes.HasPrefix(b, bufInclude):
			// Syntax:
			//   - #include "foo.p4"
			//   - #include "foo/bar.p4"
			f := bytes.Fields(b)
			if len(f) != 2 {
				return nil, fmt.Errorf("invalid include preprocessor directive: %q", string(b))
			}

			include := p.Include
			if include == nil {
				include = ioutil.ReadFile
			}

			name := strings.Trim(string(f[1]), `"`)
			inc, err := include(name)
			if err != nil {
				return nil, fmt.Errorf("preprocessor error while including %q: %v", name, err)
			}

			// Apply definitions on included source
			for k, v := range p.defines {
				inc = bytes.Replace(inc, []byte(k), []byte(v), -1)
			}

			src = append(src, inc...)
		default:
			for k, v := range p.defines {
				b = bytes.Replace(b, []byte(k), []byte(v), -1)
			}

			// bufio.Scanner removes trailing newlines so we need to put it back
			src = append(src, append(b, '\n')...)
		}
	}

	if err := p.s.Err(); err != nil {
		return nil, err
	}

	return src, nil
}

// defineNoChanges is the default Preprocessor.Define function.
func defineNoChanges(name string, value string) (string, string, error) {
	return name, value, nil
}
