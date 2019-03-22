package analyzers

import (
	"golang.org/x/tools/go/analysis"

	"github.com/gostaticanalysis/fourcetypeassert"
	"github.com/gostaticanalysis/nilerr"
	"github.com/gostaticanalysis/nofmt"
	"github.com/gostaticanalysis/notest"
	"github.com/gostaticanalysis/readonly"
	"github.com/gostaticanalysis/wraperrfmt"
)

// Recommend returns recommended analyzers including govet's one.
func Recommend() []*analysis.Analyzer {
	return append(Govet(),
		nilerr.Analyzer,
		wraperrfmt.Analyzer,
		readonly.Analyzer,
		nofmt.Analyzer,
		notest.Analyzer,
		fourcetypeassert.Analyzer,
	)
}
