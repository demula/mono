package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/demula/mono/modules"
)

const releaseUsage = "" +
	`Usage of 'mono release':
Running on the root of your monorepo and set the new version:
	mono release --only-go-mod-sum "v0.1.0-alpha.1"

Specify the root of your monorepo when not in current directory :
	mono release --only-go-mod-sum --context="./testdata" "v0.1.0-alpha.1"

To see debug information:
	mono release --debug --only-go-mod-sum "v0.1.0-alpha.1"

You can skip writing any files by using --dry-run:
	mono release --dry-run --only-go-mod-sum "v0.1.0-alpha.1"

See https://github.com/demula/mono for
examples on how to use it.
`

var ErrNoModulesFound = errors.New("no modules found")

func ReleaseCmd(
	contextDir string,
	version string,
	isDryRun bool,
	isDebug bool,
	flags *flag.FlagSet,
	args []string,
) *Command {
	return &Command{
		Name:  "release",
		Flags: flags,
		Args:  args,
		Run: func() error {
			debug(isDebug, flags, args)
			err := release(contextDir, version, isDryRun)
			if err != nil {
				if errors.Is(err, ErrNoModulesFound) {
					return fmt.Errorf("%w: no modules found at %q", ErrInput, contextDir)
				}
				return err
			}
			return nil
		},
	}
}

func release(ctxDir, version string, isDryRun bool) error {
	ms, err := modules.All(ctxDir)
	if err != nil {
		return fmt.Errorf("failed to fetch monorepo modules: %w", err)
	}
	if len(ms) == 0 {
		return ErrNoModulesFound
	}
	modules.FetchDirectDeps(ms)
	ms, err = modules.SortByDirectDeps(ms, len(ms))
	if err != nil {
		return fmt.Errorf("failed to calculate monorepo interdependencies: %w", err)
	}
	err = modules.UpdateVersion(ms, version)
	if err != nil {
		return fmt.Errorf("failed to update modules to new version: %w", err)
	}
	for _, m := range ms {
		err = modules.UpdateGoMod(m, isDryRun)
		if err != nil {
			return fmt.Errorf("failed to update \"%s/%s\" go.mod: %w", m.Prefix, m.FileName, err)
		}
		err = modules.UpdateGoSum(m, isDryRun)
		if err != nil {
			return fmt.Errorf("failed to update \"%s/%s\" go.sum: %w", m.Prefix, m.FileName, err)
		}
		slog.Info("module updated",
			slog.String("module", m.Path()),
			slog.String("gomod-hash", m.GoModHash),
			slog.String("dir-hash", m.DirHash),
		)
	}
	slog.Info("all modules updated")
	return nil
}
