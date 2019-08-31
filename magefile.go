// +build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/mattn/go-shellwords"
	"github.com/mattn/go-zglob"
)

func init() {
	os.Setenv("GOFLAGS", "-mod=vendor")
	os.Setenv("GO111MODULE", "on")
}

func ldflags() string {
	version := "unknown"
	if v, err := sh.Output("git", "describe", "--tags"); err == nil {
		version = strings.TrimSpace(v)
	}
	return fmt.Sprintf("-X 'main.version=%s'", version)
}

func runVWithArgs(cmd string, args ...string) error {
	envArgs, err := shellwords.Parse(os.Getenv("ARGS"))
	if err != nil {
		return err
	}
	return sh.RunV(cmd, append(args, envArgs...)...)
}

// Format code
func Fmt() error {
	files, err := zglob.Glob("./**/*.go")
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.Contains(file, "vendor/") {
			continue
		}
		if err := sh.RunV("goimports", "-w", file); err != nil {
			return err
		}
	}
	return nil
}

// Check coding style
func Lint() error {
	return sh.RunV("golangci-lint", "run")
}

// Run test
func Test() error {
	return sh.RunV("go", "test", "./...")
}

// Run program
func Run() error {
	return runVWithArgs("go", "run", "main.go")
}

// Install binary
func Install() error {
	return sh.RunV("go", "install", "-ldflags", ldflags(), ".")
}

// Build binary
func Build() error {
	return sh.RunV("go", "build", "-ldflags", ldflags(), ".")
}
