package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

const aliasFileName = "alias.go"

var (
	targetPkgSpec string
	writeToStdout bool
	rootCmd       = &cobra.Command{
		Use:   "aliasgen [packages]",
		Short: "create a go file that aliases outside packages into the current package scope",
		Long: `Create a file that imports the specified packages and renames the objects within as objects in the current package.

For example, if package x defines a function 'SomeFunction' and variable 'SomeVar', then a this will create a go file in the current package with:
var (
	SomeFunction = x.SomeFunction
	SomeVar      = x.SomeVar
)
Types and constants will also be aliased.`,
		Example: `aliasgen ./a/sub/package ./anotherpackage`,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return GenerateAlias(targetPkgSpec, args...)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&targetPkgSpec, "target", "t", ".", fmt.Sprintf("package to add the %s file to", aliasFileName))
	rootCmd.Flags().BoolVarP(&writeToStdout, "stdout", "s", false, "write to standard output instead of a file")
}

func GenerateAlias(targetPkgSpec string, aliasPkgSpecs ...string) error {

	// Get name of package to create alias for, and full package file path
	targetPkg, err := packages.Load(&packages.Config{Mode: packages.NeedName | packages.NeedFiles}, targetPkgSpec)
	if err != nil {
		return fmt.Errorf("could not load target package: %w", err)
	}
	if len(targetPkg) == 0 {
		return fmt.Errorf("cannot find package for path '%s'", targetPkgSpec)
	}
	if len(targetPkg) > 1 {
		return fmt.Errorf("found more than one target package for path '%s", targetPkgSpec)
	}
	errs := pkgErrors(targetPkg[0].Errors)
	errs = errs.RemoveKind(packages.TypeError) // ignore type errors
	if len(errs) > 0 {
		return fmt.Errorf("could not load target package: %w", errs)
	}
	if len(targetPkg[0].GoFiles) == 0 {
		return errors.New("cannot find files for target package")
	}
	targetPkgName := targetPkg[0].Name
	aliasFilePath := filepath.Join(filepath.Dir(targetPkg[0].GoFiles[0]), aliasFileName)

	// Parse packages to alias
	// This returns a list of packages that have been parsed into AST, and types.
	// Note: this won't return test packages unless the right Mode flag is set.
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes}, aliasPkgSpecs...)
	if err != nil {
		return fmt.Errorf("could not load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found for paths %s", aliasPkgSpecs)
	}
	for _, p := range pkgs {
		errs = pkgErrors(p.Errors)
		errs = errs.RemoveKind(packages.TypeError) // ignore type errors
		if len(errs) > 0 {
			return fmt.Errorf("could not load packages: %w", errs)
		}
	}

	// Extract data
	td := TemplateData{TargetPackageName: targetPkgName}
	for _, p := range pkgs {
		td.ExtractAndAppendAliases(p.Types)
	}

	// Generate alias file
	t, err := template.New("alias").Parse(aliasTemplate)
	if err != nil {
		return fmt.Errorf("couldn't parse template: %w", err)
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, td); err != nil {
		return fmt.Errorf("couldn't execute template: %w", err)
	}

	// Gofmt alias file
	// Note: formatting could change with different go versions
	out, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("couldn't format alias file: %w", err)
	}

	// Write out file
	if writeToStdout {
		fmt.Println(string(out))
	} else {
		if err := ioutil.WriteFile(aliasFilePath, out, 0644); err != nil {
			return fmt.Errorf("couldn't write to file: %w", err)
		}
	}

	return nil
}
