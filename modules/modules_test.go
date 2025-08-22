package modules_test

import (
	"os"
	"testing"

	"github.com/demula/mono/modules"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestDirHash(t *testing.T) {
	tests := []struct {
		version  string
		license  bool
		expected string
	}{
		{
			version:  "v0.1.0-alpha.3",
			license:  true,
			expected: "h1:pCqCciz65SYm9OAahIWOm2vKmoLPlsqe7iQfpMT7+XQ=",
		},
		{
			version:  "v0.1.0-alpha.2",
			license:  true,
			expected: "h1:DmFUhwzq4loJyq5UPKESKNmfp1bHB5eqBhbJE9Hv2sA=",
		},
		{
			version:  "v0.1.0-alpha.3",
			license:  false,
			expected: "h1:+hpncckGAguCjXrO9TRRLcamwX7JxBoilTYMrqOQFzU=",
		},
		{
			version:  "v0.1.0-alpha.2",
			license:  false,
			expected: "h1:RACcIHWGv2w4rFhFCV7RE6KBHq9qwKSPsoIY4Nz23og=",
		},
	}

	for _, tt := range tests {
		name := "with "
		if tt.license {
			name = "without "
		}
		t.Run(name+"license/"+tt.version, func(t *testing.T) {
			os.Chdir("../testdata/")
			prefix := "./golden/"
			license := ""
			if tt.license {
				prefix = "./golden-license/"
				license = "./golden-license/LICENSE"
			}
			m := &modules.Module{
				Prefix:   prefix,
				FileName: "api",
				License:  license,
				File: &modfile.File{
					Module: &modfile.Module{
						Mod: module.Version{
							Path:    "github.com/demula/mono-example/api",
							Version: tt.version,
						},
					},
				},
			}

			actual, err := modules.DirHash(m)
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			if actual != tt.expected {
				t.Errorf("hashes do not match. expected: %s, got: %s",
					tt.expected,
					actual,
				)
			}
		})
	}
}
