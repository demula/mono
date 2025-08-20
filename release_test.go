package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRelease(t *testing.T) {
	t.Parallel()
	const golden = "./testdata/golden/"
	const goldenVersion = "v1.0.0-rc.1"

	tests := []struct {
		name     string
		context  string
		version  string
		dryRun   bool
		expected string
		errMsg   string
	}{
		{
			name:     "from a previous release",
			context:  "./testdata/prev-release/",
			version:  goldenVersion,
			dryRun:   false,
			expected: golden,
		},
		{
			name:     "from development",
			context:  "./testdata/dev-commits/",
			version:  goldenVersion,
			dryRun:   false,
			expected: golden,
		},
		{
			name:     "from development with dry-run",
			context:  "./testdata/dev-commits/",
			version:  goldenVersion,
			dryRun:   true,
			expected: "./testdata/dev-commits/",
		},
		{
			name:    "invalid version",
			context: "./testdata/dev-commits/",
			version: "not#valid",
			dryRun:  false,
			errMsg:  "failed to update modules to new version: invalid version \"not#valid\"",
		},
		{
			name:    "no modules found",
			context: "./testdata/empty/",
			version: goldenVersion,
			dryRun:  false,
			errMsg:  "no modules found",
		},
		{
			name:    "wrong interdependencies",
			context: "./testdata/wrong-interdeps/",
			version: goldenVersion,
			dryRun:  false,
			errMsg:  "failed to calculate monorepo interdependencies: max iteration for sorting by direct dependencies reached",
		},

		{
			name:    "corrupt cli go.mod",
			context: "./testdata/corrupt-gomod/",
			version: goldenVersion,
			dryRun:  false,
			errMsg:  "failed to fetch monorepo modules",
		},
		{
			name:    "inconsistent cli go.sum",
			context: "./testdata/inconsistent-gosum/",
			version: goldenVersion,
			dryRun:  false,
			errMsg:  "go.sum: inconsistent dependencies",
		},
		{
			name:    "missing cli go.sum",
			context: "./testdata/missing-gosum/",
			version: goldenVersion,
			dryRun:  false,
			errMsg:  "go.sum: inconsistent dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualPath := t.TempDir()
			tmpfs := os.DirFS(tt.context)
			err := os.CopyFS(actualPath, tmpfs)
			if err != nil {
				t.Fatal(err)
			}

			err = release(actualPath, tt.version, tt.dryRun)
			if tt.errMsg != "" {
				if err == nil {
					t.Fatalf("expected error %q", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("error %q does not match expected error %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error %q", err)
			}

			expectedPath := tt.expected
			if tt.context == tt.expected {
				expectedPath = actualPath
			}

			if tt.expected == "" {
				expectedPath = golden
			}

			assertAgainstGoldenTemplate(t, actualPath, expectedPath)
		})
	}
}

func assertAgainstGoldenTemplate(t *testing.T, actualPath, expectedPath string) {
	err := filepath.WalkDir(expectedPath, func(path string, expected fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == expectedPath {
			// skip root dir
			return nil
		}

		if expected.Name() == ".git" {
			// skip git directories
			return nil
		}

		rel, err := filepath.Rel(expectedPath, path)
		if err != nil {
			t.Errorf("failed to make base for %q with root %q", path, expectedPath)
			return err
		}
		aPath := filepath.Join(actualPath, rel)
		ePath := filepath.Join(expectedPath, rel)
		actual, err := os.Stat(aPath)
		if err != nil {
			t.Errorf("failed to access %q file", aPath)
			return err
		}

		if expected.IsDir() {
			return nil
		}
		if actual.IsDir() {
			t.Errorf("%q file should not be a directory", aPath)
			return err
		}

		af, err := os.Open(aPath)
		if err != nil {
			t.Errorf("could not compare %q. error opening %q: %s",
				expected.Name(), aPath, err.Error(),
			)
			return err
		}
		defer func() {
			err := af.Close()
			if err != nil {
				t.Errorf("could not close file %q: %s", aPath, err.Error())
			}
		}()

		ef, err := os.Open(ePath)
		if err != nil {
			t.Errorf("could not compare %q. error opening %q: %s",
				expected.Name(), ePath, err.Error(),
			)
			return err
		}
		defer func() {
			err := ef.Close()
			if err != nil {
				t.Errorf("could not close file %q: %s", aPath, err.Error())
			}
		}()
		assertEqualFile(t, ef, af)
		return nil
	})
	if err != nil {
		t.Errorf("failed when walking directory %q: %s",
			expectedPath, err.Error(),
		)
	}
}

func assertEqualFile(t *testing.T, expected, actual *os.File) {
	expectedLines, err := readLines(expected)
	if err != nil {
		t.Errorf("could not read lines from  %q. error : %s",
			expected.Name(), err.Error(),
		)
		return
	}
	actualLines, err := readLines(actual)
	if err != nil {
		t.Errorf("could not read lines from  %q. error : %s",
			actual.Name(), err.Error(),
		)
		return
	}
	if len(actualLines) != len(expectedLines) {
		t.Errorf("expected %q content size %d does not match %d.\nexpected:\n%s\ngot:\n%s\n",
			expected.Name(),
			len(expectedLines),
			len(actualLines),
			printLines(expectedLines, -1),
			printLines(actualLines, -1),
		)
		return
	}
	for i := range expectedLines {
		if actualLines[i] != expectedLines[i] {
			t.Errorf("expected %q content differs on line %d.\nexpected:\n%s\ngot:\n%s\n",
				expected.Name(),
				i,
				printLines(expectedLines, i),
				printLines(actualLines, i),
			)
			return
		}
	}
}

func readLines(f *os.File) ([]string, error) {
	actualScanner := bufio.NewScanner(f)
	var out []string
	for actualScanner.Scan() {
		out = append(out, actualScanner.Text())
	}
	return out, actualScanner.Err()
}

func printLines(lines []string, markLine int) string {
	sb := strings.Builder{}
	for i, line := range lines {
		sb.WriteString("    ")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(": ")
		sb.WriteString(line)
		sb.WriteString("\n")
		if markLine == i {
			sb.WriteString("       ^^^^^^^^^\n\n")
		}
	}
	return sb.String()
}
