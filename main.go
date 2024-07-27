package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"bobik.squidwock.com/root/gal/genalpha/interpreter"
	"bobik.squidwock.com/root/gal/genalpha/lexer"
	"bobik.squidwock.com/root/gal/genalpha/parser"
	"bobik.squidwock.com/root/gal/genalpha/pkg"
	"bobik.squidwock.com/root/gal/genalpha/utils"
	"github.com/google/subcommands"
)

type runCmd struct{}
type installCmd struct{}
type uninstallCmd struct{}
type buildCmd struct{}

func (*runCmd) Name() string       { return "run" }
func (*installCmd) Name() string   { return "install" }
func (*uninstallCmd) Name() string { return "uninstall" }
func (*buildCmd) Name() string     { return "build" }

func (*runCmd) Synopsis() string       { return "Run the specified file" }
func (*installCmd) Synopsis() string   { return "Install the specified package" }
func (*uninstallCmd) Synopsis() string { return "Uninstall the specified package" }
func (*buildCmd) Synopsis() string     { return "Build a package" }

func (*runCmd) Usage() string       { return "run <path>" }
func (*installCmd) Usage() string   { return "install <package>" }
func (*uninstallCmd) Usage() string { return "uninstall <package>" }
func (*buildCmd) Usage() string     { return "build <path>" }

func (p *runCmd) SetFlags(f *flag.FlagSet)       {}
func (p *installCmd) SetFlags(f *flag.FlagSet)   {}
func (p *uninstallCmd) SetFlags(f *flag.FlagSet) {}
func (p *buildCmd) SetFlags(f *flag.FlagSet)     {}

func (p *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		panic("missing path to file")
	}
	filename := f.Arg(0)
	contents := utils.ReadContents(filename)
	tokens := lexer.Lex(contents)
	ast := parser.Parse(tokens)
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	interpreter.Interpret(&ast, f.Args()[1:], currentDir+"/")
	return subcommands.ExitSuccess
}

func (p *installCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		// install from package.yml
		return subcommands.ExitSuccess
	}
	name := f.Arg(0)
	err := pkg.InstallPackage(name)
	if err != nil {
		panic(err)
	}
	return subcommands.ExitSuccess
}

func (p *uninstallCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if f.NArg() == 0 {
		panic("missing command")
	}
	name := f.Arg(0)
	err := pkg.UninstallPackage(name)
	if err != nil {
		panic(err)
	}
	return subcommands.ExitSuccess
}

func (p *buildCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	filename := f.Arg(0)

	if f.NArg() == 0 {
		filename = "package.yml"
	}

	pack, err := pkg.ParsePackage(filename)
	if err != nil {
		panic(err)
	}
	path, err := pkg.BuildPackage(pack)
	if err != nil {
		panic(err)
	}
	fmt.Println("built package at", path)
	return subcommands.ExitSuccess
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			if r == "" {
				return
			}
			fmt.Println(r)
		}
	}()

	subcommands.Register(&runCmd{}, "")
	subcommands.Register(&installCmd{}, "")
	subcommands.Register(&uninstallCmd{}, "")
	subcommands.Register(&buildCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
