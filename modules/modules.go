package modules

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/demula/mono/gosum"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"
)

type Module struct {
	Prefix      string
	FileName    string
	GoModHash   string
	DirHash     string
	File        *modfile.File
	Deps        []*Module
	DepsVersion []string
	Sums        map[module.Version][]string
}

func (m *Module) Path() string {
	return m.File.Module.Mod.Path
}

func (m *Module) Version() string {
	return m.File.Module.Mod.Version
}

func All(prefix string) ([]*Module, error) {
	prefix = filepath.Clean(prefix)
	dfs, err := os.ReadDir(prefix)
	if err != nil {
		return nil, err
	}
	var ms []*Module
	for _, f := range dfs {
		if !f.IsDir() || strings.HasPrefix(f.Name(), ".") {
			continue
		}
		gomod := filepath.Join(prefix, f.Name(), "go.mod")
		_, err := os.Stat(gomod)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		m := &Module{
			Prefix:   prefix,
			FileName: f.Name(),
			Sums:     make(map[module.Version][]string),
		}
		contents, err := os.ReadFile(gomod)
		if err != nil {
			return nil, err
		}
		m.File, err = modfile.Parse(gomod, contents, nil)
		if err != nil {
			return nil, err
		}
		sum := filepath.Join(prefix, f.Name(), "go.sum")
		err = gosum.Read(m.Sums, sum)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
		debug(m, "found monorepo module at %s",
			filepath.Join(prefix, f.Name()),
		)
	}
	return ms, nil
}

func FetchDirectDeps(mods []*Module) {
	for _, m := range mods {
		for _, require := range m.File.Require {
			for _, d := range mods {
				if require.Mod.Path == d.Path() {
					m.Deps = append(m.Deps, d)
					m.DepsVersion = append(m.DepsVersion, require.Mod.Version)
					debug(m, "found interdependency %s@%s", d.Path(), require.Mod.Version)
				}
			}
		}
	}
}

func UpdateVersion(mods []*Module, version string) error {
	if !semver.IsValid(version) {
		return fmt.Errorf("invalid version %q", version)
	}
	for _, m := range mods {
		m.File.Module.Mod.Version = version
		for _, d := range m.Deps {
			err := m.File.AddRequire(d.Path(), version)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UpdateGoMod(m *Module, dry bool) error {
	path := filepath.Join(m.Prefix, m.FileName, "go.mod")
	m.File.Cleanup()
	data, _ := m.File.Format()
	var err error
	m.GoModHash, err = GoModHash(data)
	if err != nil {
		return err
	}
	if dry {
		debug(m, "[skipped] writing file %s", path)
		return nil
	}
	debug(m, "writing file %s", path)
	return os.WriteFile(path, data, 0644)
}

func UpdateGoSum(m *Module, dry bool) error {
	path := filepath.Join(m.Prefix, m.FileName, "go.sum")
	for i, d := range m.Deps {
		err := updateSum(m, d, m.DepsVersion[i], "", d.DirHash)
		if err != nil {
			return fmt.Errorf("inconsistent dependencies. failed to update dir hash: %w", err)
		}
		err = updateSum(m, d, m.DepsVersion[i], "/go.mod", d.GoModHash)
		if err != nil {
			return fmt.Errorf("inconsistent dependencies. failed to update go.mod hash: %w", err)
		}
	}
	data := gosum.Format(m.Sums)
	if !dry {
		debug(m, "writing file %s", path)
		err := os.WriteFile(path, data, 0644)
		if err != nil {
			return err
		}
	} else {
		debug(m, "[skipped] writing file %s", path)
	}
	var err error
	m.DirHash, err = DirHash(m)
	return err
}

func updateSum(m *Module, d *Module, version, suffix, hash string) error {
	md := module.Version{
		Path:    d.Path(),
		Version: d.Version() + suffix,
	}
	if len(hash) == 0 {
		return errors.New("empty hash for module" + d.Path())
	}
	m.Sums[md] = []string{hash}

	if len(m.Sums) > 0 {
		mdOld := module.Version{
			Path:    d.Path(),
			Version: version + suffix,
		}
		hashOld, ok := m.Sums[mdOld]
		if !ok {
			return errors.New("missing go sum entry for " + mdOld.String())
		}
		delete(m.Sums, mdOld)
		debug(m, "changed dep %s%s %s --> %s", d.Path(), suffix, hashOld, hash)
	} else {
		debug(m, "added dep to empty sums %s%s %s", d.Path(), suffix, hash)
	}
	return nil
}

func SortByDirectDeps(nodes []*Module, maxIter int) ([]*Module, error) {
	if len(nodes) < 2 {
		return nodes, nil
	}
	slices.SortFunc(nodes, func(a, b *Module) int {
		return cmp.Compare(len(a.Deps), len(b.Deps))
	})
	var resolved []*Module
	unresolved := nodes
	for range maxIter {
		iterUnresolved := []*Module{}
		for _, n := range unresolved {
			if len(n.Deps) == 0 {
				resolved = append(resolved, n)
				continue
			}
			isUnresolved := false
			for _, d := range n.Deps {
				dFound := false
				for _, r := range resolved {
					if d.Path() == r.Path() {
						dFound = true
						break
					}
				}
				if !dFound {
					isUnresolved = true
					break
				}
			}
			if isUnresolved {
				iterUnresolved = append(iterUnresolved, n)
			} else {
				resolved = append(resolved, n)
			}
		}
		if len(iterUnresolved) == 0 {
			return resolved, nil
		}
		unresolved = iterUnresolved
	}
	return nil, errors.New("max iteration for sorting by direct dependencies reached")
}

func GoModHash(data []byte) (string, error) {
	return dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	})
}

// DirHash reads directory and produces its H1 hash.
// Note: remember to modify the go.mod file first before running this function.
func DirHash(m *Module) (string, error) {
	return dirhash.HashDir(
		filepath.Join(m.Prefix, m.FileName),
		m.Path()+"@"+m.Version(),
		dirhash.DefaultHash,
	)
}

func debug(m *Module, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	slog.Debug(msg, slog.String("module", m.Path()))
}
