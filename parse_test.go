package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		expected  *TestCommand
	}{
		{
			name:      "no args",
			arguments: []string{},
			expected: &TestCommand{
				Name: "base",
			},
		},
		{
			name:      "base help",
			arguments: []string{"--help"},
			expected: &TestCommand{
				Name: "base",
				Flags: []string{
					"--help=true",
				},
			},
		},
		{
			name:      "missing subcommand",
			arguments: []string{"--context=."},
			expected: &TestCommand{
				Name: "base",
				Flags: []string{
					"--context=.",
				},
				Error: "input error. missing subcommand",
			},
		},
		{
			name:      "unknown subcommand",
			arguments: []string{"non-existing"},
			expected: &TestCommand{
				Name:  "base",
				Error: "input error. unknown subcommand \"non-existing\"",
			},
		},
		{
			name:      "invalid context",
			arguments: []string{"--context=test"},
			expected: &TestCommand{
				Name: "base",
				Error: "input error. " +
					"invalid value \"test\" for flag -context: invalid provided directory path \"test\". " +
					"stat test: no such file or directory",
			},
		},
		{
			name:      "base unknown flag",
			arguments: []string{"--unknown"},
			expected: &TestCommand{
				Name:  "base",
				Error: "input error. flag provided but not defined: -unknown",
			},
		},
		{
			name:      "release help",
			arguments: []string{"release", "--help"},
			expected: &TestCommand{
				Name: "release",
				Flags: []string{
					"--help=true",
				},
			},
		},
		{
			name:      "release help before command",
			arguments: []string{"--help", "release"},
			expected: &TestCommand{
				Name: "release",
				Flags: []string{
					"--help=true",
				},
			},
		},
		{
			name:      "release unknown flag",
			arguments: []string{"release", "--unknown"},
			expected: &TestCommand{
				Name:  "release",
				Error: "input error. flag provided but not defined: -unknown",
			},
		},
		{
			name:      "flags after arguments are recognized as arguments",
			arguments: []string{"release", "v0.1.0", "--unknown"},
			expected: &TestCommand{
				Name:  "release",
				Error: "input error. too many arguments",
			},
		},
		{
			name: "release invalid version",
			arguments: []string{
				"release",
				"--only-go-mod-sum",
				"non-valid",
			},
			expected: &TestCommand{
				Name: "release",
				Flags: []string{
					"--only-go-mod-sum=true",
				},
				Error: "input error. invalid version provided",
			},
		},
		{
			name: "release missing version argument",
			arguments: []string{
				"release",
				"--only-go-mod-sum",
			},
			expected: &TestCommand{
				Name: "release",
				Flags: []string{
					"--only-go-mod-sum=true",
				},
				Error: "input error. missing version argument",
			},
		},
		{
			name: "release missing 'only-go-mod-sum' mode",
			arguments: []string{
				"release",
				"v0.3.0",
			},
			expected: &TestCommand{
				Name:  "release",
				Error: `input error. only "--only-go-mod-sum" mode is supported`,
			},
		},
		{
			name: "release with all flags",
			arguments: []string{
				"--context=./testdata/",
				"--debug",
				"release",
				"--dry-run",
				"--only-go-mod-sum",
				"v0.0.0-20170915032832-14c0d48ead0c",
			},
			expected: &TestCommand{
				Name: "release",
				Args: []string{
					"v0.0.0-20170915032832-14c0d48ead0c",
				},
				Flags: []string{
					"--context=testdata",
					"--debug=true",
					"--dry-run=true",
					"--only-go-mod-sum=true",
				},
			},
		},
	}
	slog.SetLogLoggerLevel(slog.LevelError)
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rs := flag.NewFlagSet("root", flag.ContinueOnError)
			rs.SetOutput(io.Discard)
			actual := parse(rs, tt.arguments)
			assertEqual(t, tt.expected, actual)
		})
	}
}

func TestParseUsageSet(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		expected  string
	}{
		{
			name:      "base",
			arguments: []string{"--help"},
			expected:  baseUsage,
		},
		{
			name:      "release",
			arguments: []string{"release", "--help"},
			expected:  releaseUsage,
		},
	}

	slog.SetLogLoggerLevel(slog.LevelError)
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			out := bufio.NewWriter(buf)
			rs := flag.NewFlagSet("root", flag.ContinueOnError)
			rs.SetOutput(out)

			cmd := parse(rs, tt.arguments)
			err := cmd.Run()
			if err != nil {
				t.Fatalf("unexpected error %q", err.Error())
			}
			err = out.Flush()
			if err != nil {
				t.Fatalf("unexpected error %q while flushing content before checks", err.Error())
			}
			actual := buf.String()
			if !strings.Contains(actual, tt.expected) {
				t.Errorf("missing expected usage.\nexpected:\n%s\ngot:\n%s\n",
					tt.expected, actual)
			}
		})
	}
}

func TestParseCheckRun(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
	}{
		{
			name:      "base",
			arguments: []string{"--help", "--debug"},
		},
		{
			name: "release",
			arguments: []string{
				"--context=./testdata/prev-release/",
				"--debug",
				"release",
				"--dry-run",
				"--only-go-mod-sum",
				"v0.0.0-20170915032832-14c0d48ead0c",
			},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := flag.NewFlagSet("root", flag.ContinueOnError)
			rs.SetOutput(io.Discard)
			l := slog.New(slog.DiscardHandler)
			slog.SetDefault(l)
			cmd := parse(rs, tt.arguments)
			err := cmd.Run()
			if err != nil {
				t.Fatalf("unexpected error %q", err.Error())
			}
		})
	}
}

func assertEqual(t *testing.T, expected *TestCommand, actual *Command) {
	e := expected.String()
	a := actual.String()
	if e != a {
		t.Errorf("commands differ.\nexpected: %s\ngot: %s", e, a)
	}
}

type TestCommand struct {
	Name  string
	Args  []string
	Flags []string
	Error string
	Run   func()
}

func (c *TestCommand) String() string {
	if c == nil {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString("{ name: '")
	sb.WriteString(c.Name)
	sb.WriteString("', args: [")
	for i, a := range c.Args {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("'")
		sb.WriteString(a)
		sb.WriteString("'")
	}
	sb.WriteString("], flags: [")
	for i, f := range c.Flags {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("'")
		sb.WriteString(f)
		sb.WriteString("'")
	}
	sb.WriteString("], error: '")
	sb.WriteString(c.Error)
	sb.WriteString("' }")
	return sb.String()
}
