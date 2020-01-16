package main

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

const aliasFileName = "alias.go"

var (
	targetPkgSpec string
	rootCmd       = &cobra.Command{
		Use:   "aliasgen [packages]",
		Short: "short help text",
		Long:  `long help text`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(targetPkgSpec, args...)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&targetPkgSpec, "target", "t", ".", fmt.Sprintf("package to add a %s file to", aliasFileName))
}

func Run(targetPkgSpec string, aliasPkgSpecs ...string) error {

	// Get name of package to create alias for
	targetPkg, err := packages.Load(&packages.Config{Mode: packages.NeedName}, targetPkgSpec)
	if err != nil {
		return fmt.Errorf("could not load target package: %w", err)
	}
	if len(targetPkg) == 0 {
		return fmt.Errorf("cannot find package for path '%s'", targetPkgSpec)
	}
	if len(targetPkg) > 1 {
		return fmt.Errorf("found more than one target package for path '%s", targetPkgSpec)
	}
	targetPkgName := targetPkg[0].Name

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
	fmt.Println("AFASDF", targetPkg[0].PkgPath)
	// TODO need to get file name for writing alias to
	// ideally specify as a system path, then load module from there, and write to the file path
	// but don't know how to load module from there
	// this should ideally also check whether alias pkgs can be imported from target
	// if err := ioutil.WriteFile("/tmp/dat1", out, 0644); err != nil {
	// 	return fmt.Errorf("couldn't write to file: %w", err)
	// }
	fmt.Println(string(out))

	return nil
}
