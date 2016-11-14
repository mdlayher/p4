package preprocessor_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mdlayher/p4/preprocessor"
)

func TestNoProcessing(t *testing.T) {
	const in = `
header_type foo_t {
	fields {
		foo : 8;
	}
}
`

	p := preprocessor.New(strings.NewReader(in))
	out, err := p.Process()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want, got := in, string(out); want != got {
		t.Fatalf("unexpected output:\n- want:\n%v\n-  got:%v\n", want, got)
	}
}

func TestDefine(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		out     string
		invalid bool
	}{
		{
			name: "no name for define",
			in: `
#define
header_type foo_t {
	fields {
		foo : 8;
	}
}
`,
			invalid: true,
		},
		{
			name: "no value for define",
			in: `
#define FOO_BITS
header_type foo_t {
	fields {
		foo : FOO_BITS;
	}
}
`,
			invalid: true,
		},
		{
			name: "one define",
			in: `
#define FOO_BITS 8
header_type foo_t {
	fields {
		foo : FOO_BITS;
	}
}
`,
			out: `
header_type foo_t {
	fields {
		foo : 8;
	}
}
`,
		},
		{
			name: "multiple defines",
			in: `
#define FOO_BITS 8
#define BAR_BITS 16
#define FOO_LEN 24
header_type foo_t {
	fields {
		foo : FOO_BITS;
		bar : BAR_BITS;
	}
	length : FOO_LEN;
}
`,
			out: `
header_type foo_t {
	fields {
		foo : 8;
		bar : 16;
	}
	length : 24;
}
`,
		},
		{
			name: "FOO_BITS defined twice",
			in: `
#define FOO_BITS 8
#define FOO_BITS 16
header_type foo_t {
	fields {
		foo : FOO_BITS;
	}
}
`,
			out: `
header_type foo_t {
	fields {
		foo : 16;
	}
}
`,
		},
		{
			name: "define uses another defined value",
			in: `
#define FOO_BITS 8
#define BAR_BITS FOO_BITS
header_type foo_t {
	fields {
		foo : BAR_BITS;
	}
}
`,
			out: `
header_type foo_t {
	fields {
		foo : 8;
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := preprocessor.New(strings.NewReader(tt.in))
			out, err := p.Process()
			if err != nil && !tt.invalid {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.invalid {
				return
			}

			if want, got := tt.out, string(out); want != got {
				t.Fatalf("unexpected output:\n- want:\n%v\n-  got:%v\n", want, got)
			}
		})
	}
}

func TestInclude(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		include map[string]string
		out     string
		invalid bool
	}{
		{
			name: "no value for include",
			in: `
#include
header_type foo_t {
	fields {
		foo : FOO_BITS;
	}
}
`,
			invalid: true,
		},
		{
			name: "include non-existent file",
			in: `
#include "bar.p4"

header_type foo_t {
	fields {
		foo : FOO_BITS;
	}
}
`,
			invalid: true,
		},
		{
			name: "one include",
			in: `
#include "bar.p4"

header_type foo_t {
	fields {
		foo : 8;
	}
}
`,
			include: map[string]string{
				"bar.p4": `
header_type bar_t {
	fields {
		bar : 16;
	}
}
`,
			},
			out: `

header_type bar_t {
	fields {
		bar : 16;
	}
}

header_type foo_t {
	fields {
		foo : 8;
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := preprocessor.New(strings.NewReader(tt.in))
			p.Include = func(name string) ([]byte, error) {
				f, ok := tt.include[name]
				if !ok {
					return nil, os.ErrNotExist
				}

				return []byte(f), nil
			}

			out, err := p.Process()
			if err != nil && !tt.invalid {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.invalid {
				return
			}

			if want, got := tt.out, string(out); want != got {
				t.Fatalf("unexpected output:\n- want:\n%v\n-  got:%v\n", want, got)
			}
		})
	}
}
