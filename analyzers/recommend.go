package analyzers

import (
	"golang.org/x/tools/go/analysis"

	"github.com/gostaticanalysis/ctxfield"
	"github.com/gostaticanalysis/dupimport"
	"github.com/gostaticanalysis/forcetypeassert"
	"github.com/gostaticanalysis/importgroup"
	"github.com/gostaticanalysis/nilerr"
	"github.com/gostaticanalysis/nofmt"
	"github.com/gostaticanalysis/notest"
	"github.com/gostaticanalysis/readonly"
	"github.com/gostaticanalysis/unitconst"
	"github.com/gostaticanalysis/unused"
)

// Recommend returns recommended analyzers including govet's one.
func Recommend() []*analysis.Analyzer {
	return append(Govet(),
		ctxfield.Analyzer,
		dupimport.Analyzer,
		forcetypeassert.Analyzer,
		importgroup.Analyzer,
		nilerr.Analyzer,
		nofmt.Analyzer,
		notest.Analyzer,
		readonly.Analyzer,
		unitconst.Analyzer,
		unused.Analyzer,
	)
}
