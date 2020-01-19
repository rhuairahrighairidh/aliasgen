# Aliasgen

https://github.com/rigelrozanski/multitool/blob/master/cmd/mt/commands/golang.go

example of a go code generators:
- https://github.com/campoy/jsonenums
- https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go
- seems like using templates would be simplest


resources on go parsing
- https://github.com/golang/example/tree/master/gotypes
- https://godoc.org/go/types#Scope
- https://arslan.io/2017/09/14/the-ultimate-guide-to-writing-a-go-tool/
- https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/


this could be run by go generate - leave a comment at the top of a file in a module root directory

Packages contain a name and a path.
Path is the import path - relative to gopath/src or the containing module.
However, if the pkgSpec is outside of the current module (or presumably GOPATH) then the name is unknown and the path is "command-line-arguments". Though if the package isn't even a go package an error is returned.
Also in outside-module case the pkg.Errors has an error about not being inside a package