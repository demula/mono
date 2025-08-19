package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	cmd := parse(flag.CommandLine, os.Args[1:])
	if cmd.Error != nil {
		if errors.Is(cmd.Error, flag.ErrHelp) {
			cmd.Flags.Usage()
			os.Exit(2)
		}
		_, err := fmt.Fprintln(cmd.Flags.Output(), cmd.Error)
		if err != nil {
			panic(cmd.Error)
		}
		_, err = fmt.Fprintln(cmd.Flags.Output(), "")
		if err != nil {
			panic(cmd.Error)
		}
		cmd.Flags.Usage()
		if errors.Is(cmd.Error, ErrInput) {
			os.Exit(2)
		}
		os.Exit(1)
	}
	err := cmd.Run()
	if err != nil {
		if errors.Is(cmd.Error, ErrInput) {
			_, pErr := fmt.Fprintln(cmd.Flags.Output(), cmd.Error)
			if pErr != nil {
				panic(err)
			}
			_, pErr = fmt.Fprintln(cmd.Flags.Output(), "")
			if pErr != nil {
				panic(err)
			}
			cmd.Flags.Usage()
			os.Exit(2)
		}
		slog.Error(err.Error())
		os.Exit(1)
	}
}