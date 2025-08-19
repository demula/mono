package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

var ErrInput = errors.New("input error")

const baseUsage = "" +
	`Usage of 'mono':
Running on the root of your monorepo 'mono {{subcommand}} {{arguments}}:
	mono release --only-go-mod-sum "v0.1.0-alpha.1"

Global flags are allowed before subcommand:
	mono --debug release "v0.1.0-alpha.1"

See https://github.com/demula/mono for
examples on how to use it.
`

func usage(fs *flag.FlagSet, msg string) func() {
	return func() {
		_, err := fmt.Fprint(fs.Output(), msg)
		if err != nil {
			slog.Error("could not use given output for printing usage",
				slog.String("error", err.Error()),
			)
		}
		_, err = fmt.Fprintln(fs.Output(), "\nFlags:")
		if err != nil {
			slog.Error("could not use given output for printing usage",
				slog.String("error", err.Error()),
			)
		}
		fs.PrintDefaults()
	}
}

func parse(baseFS *flag.FlagSet, arguments []string) *Command {
	cmd := &Command{
		Name:  "base",
		Flags: baseFS,
		Run:   func() error { return nil },
	}
	baseFS.Usage = usage(baseFS, baseUsage)

	isDebug := baseFS.Bool("debug", false, "print debug messages")
	getHelp := baseFS.Bool("help", false, "print help information")

	defaultDir, err := os.Getwd()
	if err != nil {
		cmd.Error = errors.New("failed to get current directory")
		return cmd
	}
	contextDir := DirValue(baseFS, "context", defaultDir, "specify starting monorepo folder to look for modules")

	err = baseFS.Parse(arguments)
	if err != nil {
		cmd.Error = fmt.Errorf("%w. %w", ErrInput, err)
		return cmd
	}

	args := baseFS.Args()
	cmd.Run = func() error {
		debug(*isDebug, baseFS, args)
		baseFS.Usage()
		return nil
	}
	if baseFS.NFlag() == 0 && baseFS.NArg() == 0 {
		return cmd
	}
	if len(args) == 0 {
		if *getHelp {
			return cmd
		}
		cmd.Error = fmt.Errorf("%w. missing subcommand", ErrInput)
		return cmd
	}
	cmdName := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = nil
	}
	switch cmdName {
	case "release":
		cmd.Name = "release"
		relFS := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
		relFS.SetOutput(baseFS.Output()) // inherit
		relFS.Usage = usage(relFS, releaseUsage)
		cmd.Flags = relFS
		cmd.Run = func() error {
			debug(*isDebug, relFS, args)
			relFS.Usage()
			return nil
		}

		// Register global flags
		baseFS.VisitAll(func(f *flag.Flag) {
			relFS.Var(f.Value, f.Name, f.Usage)
		})
		// Reset global flags (easier to test setup using cmd.String())
		var resetErr error
		baseFS.Visit(func(f *flag.Flag) {
			if resetErr != nil {
				return
			}
			err := relFS.Set(f.Name, f.Value.String())
			if err != nil {
				resetErr = fmt.Errorf("could not reset flag %q to %q: %w",
					f.Name, f.Value.String(), err)
			}
		})
		if resetErr != nil {
			cmd.Error = resetErr
			return cmd
		}
		// Register local flags
		var (
			isDryRun   = relFS.Bool("dry-run", false, "skip writing to files")
			isOnlyMode = relFS.Bool("only-go-mod-sum", false, "only change go.mod and go.sum files")
		)
		err := relFS.Parse(args)
		if err != nil {
			cmd.Error = fmt.Errorf("%w. %w", ErrInput, err)
			return cmd
		}
		args = relFS.Args()
		if len(args) == 0 {
			if *getHelp {
				return cmd
			}
			cmd.Error = fmt.Errorf("%w. missing version argument", ErrInput)
			return cmd
		}
		if len(args) > 1 {
			cmd.Error = fmt.Errorf("%w. too many arguments", ErrInput)
			return cmd
		}

		if !*isOnlyMode {
			cmd.Error = fmt.Errorf("%w. only \"--only-go-mod-sum\" mode is supported", ErrInput)
			return cmd
		}

		version := args[0]
		if version == "" || !semver.IsValid(version) {
			cmd.Error = fmt.Errorf("%w. invalid version provided", ErrInput)
			return cmd
		}
		cmd = ReleaseCmd(string(*contextDir), version, *isDryRun, *isDebug, relFS, args)
	default:
		cmd.Error = fmt.Errorf("%w. unknown subcommand %q", ErrInput, cmdName)
		return cmd
	}
	return cmd
}

func debug(isDebug bool, fs *flag.FlagSet, args []string) {
	if isDebug {
		slog.SetLogLoggerLevel(slog.LevelDebug)

		var attrs []any
		fs.VisitAll(func(f *flag.Flag) {
			attrs = append(attrs, slog.String(
				f.Name, f.Value.String(),
			))
		})
		for i, arg := range args {
			attrs = append(attrs, slog.String(
				fmt.Sprintf("arg-%d", i), arg,
			))
		}
		slog.Debug("command config", attrs...)
	}
}

type Command struct {
	Name  string
	Args  []string
	Flags *flag.FlagSet
	Error error
	Run   func() error
}

func (c *Command) String() string {
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
	if c.Flags != nil {
		i := 0
		c.Flags.Visit(func(f *flag.Flag) {
			if i != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("'--")
			sb.WriteString(f.Name)
			sb.WriteString("=")
			sb.WriteString(f.Value.String())
			sb.WriteString("'")
			i++
		})
	}
	sb.WriteString("], error: '")
	if c.Error != nil {
		sb.WriteString(c.Error.Error())
	}
	sb.WriteString("' }")
	return sb.String()
}

type dirValue string

func DirValue(fs *flag.FlagSet, name string, value string, usage string) *dirValue {
	dv := dirValue(value)
	fs.Var(&dv, name, usage)
	return &dv
}

func (s *dirValue) Set(val string) error {
	val = filepath.Clean(val)
	f, err := os.Stat(val)
	if err != nil {
		return fmt.Errorf("invalid provided directory path %q. %w", val, err)
	}
	if !f.IsDir() {
		return fmt.Errorf("invalid provided directory path %q. not a directory", val)
	}
	*s = dirValue(val)
	return nil
}

func (s *dirValue) Get() any { return string(*s) }

func (s *dirValue) String() string { return string(*s) }