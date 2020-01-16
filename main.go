package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

//https://github.com/rigelrozanski/multitool/blob/master/cmd/mt/commands/golang.go

/* example of a go code generators:
- https://github.com/campoy/jsonenums
- https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go
- seems like using templates would be simplest
*/

/* resources on go parsing
- https://github.com/golang/example/tree/master/gotypes
- https://godoc.org/go/types#Scope
- https://arslan.io/2017/09/14/the-ultimate-guide-to-writing-a-go-tool/
- https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/
*/

// this could be run by go generate - leave a comment at the top of a file in a module root directory

/* TODOs
- split into get types, and print output, decide on data type to pass between the two (does it need one?)
	- also combine multiple pkgs into one struct
- get enclosing package name
- could load enclosing package and modify AST and write back to the directory
*/

var rootCmd = &cobra.Command{
	Use:   "aliasgen [packages]..",
	Short: "short help text",
	Long:  `long help text`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		Main(args)
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func Main(ps []string) {
	fmt.Println(ps)
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes}, ps...) // won't return test pkgs unless Mode flag is set
	if err != nil {
		panic(err)
	}
	for _, p := range pkgs {
		fmt.Println(p.Types.Path()) // same things as pass.Pkg
	}
	fmt.Println("==========")
	run(pkgs[0].Types)

}

func run(pkg *types.Package) (interface{}, error) {
	// how to ignore all test packages? it seems to give packages more than once sometimes
	fmt.Println("---------------------------")
	fmt.Printf("package name: %v\n", pkg.Name())
	fmt.Printf("package path: %v\n", pkg.Path()) // use this for the import path

	// pass.Pkg gives info across the whole set of files in a package
	// notably Scope which contains all the objects defined in tha package

	// pass.TypesInfo gives the thing that would be returned by running the type checker

	// pass.Fset and pass.Files are the source location information, and the AST for each file

	// Looks like it's easy to get all object names from scope, then look them up to get the objects,
	// then type switch them to get the 4 object types we're interested in
	td := TemplateData{
		PackageName: "todo", // TODO need to get enclosing package name, or target name
		Imports:     []string{pkg.Path()},
	}
	objectNames := pkg.Scope().Names()
	for _, n := range objectNames {
		obj := pkg.Scope().Lookup(n)
		if obj.Exported() {
			switch obj.(type) {
			case *types.Const:
				td.Consts = append(td.Consts, Alias{Pkg: pkg.Name(), Name: obj.Name()})
			case *types.Func:
				td.Funcs = append(td.Funcs, Alias{Pkg: pkg.Name(), Name: obj.Name()})
			case *types.Var:
				td.Vars = append(td.Vars, Alias{Pkg: pkg.Name(), Name: obj.Name()})
			case *types.TypeName:
				td.Types = append(td.Types, Alias{Pkg: pkg.Name(), Name: obj.Name()})
			default:
				fmt.Println("-----")
				fmt.Printf(" obj: %v\n", obj)
				fmt.Printf("name: %v\n", obj.Name())
				if obj.Exported() {
					fmt.Printf("%v = %v.%v\n", obj.Id(), pkg.Name(), obj.Id())
				}
			}
		}
	}
	// TODO need to track if there was no objects added, and if so remove the import so it's not printed.

	// then construct a struct thingy with name, package name, class (const,var,type) for each object
	// do this for all packages, filtering out test ones
	// combine structs together - might need to run this outside analysis framework
	// pass into a template and save the file out
	// also need to add imports
	t, err := template.New("t").Parse(aliasTemplate)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, td)
	if err != nil {
		panic(err)
	}

	out, err := format.Source(buf.Bytes()) // formatting could change with different go versions
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	return nil, nil
}
