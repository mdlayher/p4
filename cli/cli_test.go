package cli_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/mdlayher/p4/cli"
)

func TestExecutor(t *testing.T) {
	tests := []struct {
		name string
		exec []cli.Executor
		in   string
		out  string
	}{
		{
			name: "only preprocessor",
			exec: []cli.Executor{cli.NewPreprocessor()},
			in: `
#define FOO 1
FOO
`,
			out: `
1
`,
		},
		{
			name: "preprocessor ran twice",
			exec: []cli.Executor{
				cli.NewPreprocessor(),
				cli.NewPreprocessor(),
			},
			in: `
#define FOO 2
FOO
`,
			out: `
2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			r := bytes.NewBufferString(tt.in)

			for _, e := range tt.exec {
				w.Reset()
				if err := e.Execute(w, r); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Use output of last stage as input for the next stage.  w is not used
				// directly as io.Reader so its internal position pointer remains in the
				// same place, but we can still read its contents.
				r.Reset()
				if _, err := io.Copy(r, bytes.NewReader(w.Bytes())); err != nil {
					t.Fatalf("failed to copy: %v", err)
				}
			}

			if want, got := tt.out, w.String(); want != got {
				t.Fatalf("unexpected output:\n- want:\n%v\n-  got:\n%s", want, got)
			}
		})
	}
}
