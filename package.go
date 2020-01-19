package main

import (
	"golang.org/x/tools/go/packages"
	"strings"
)

// pkgErrors is a convenience error type that combines errors returned from parsing packages
type pkgErrors []packages.Error

func (pe pkgErrors) Error() string {
	var msgs []string
	for _, e := range pe {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "\n")
}

// RemoveKind returns a pkgErrors with all errors matching 'kind' removed.
func (pe pkgErrors) RemoveKind(kind packages.ErrorKind) pkgErrors {
	var pkgErrs pkgErrors
	for _, e := range pe {
		if e.Kind == kind {
			continue
		}
		pkgErrs = append(pkgErrs, e)
	}
	return pkgErrs
}
