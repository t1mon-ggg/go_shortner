package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/go-critic/go-critic/checkers/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"

	"github.com/t1mon-ggg/go_shortner/internal/linter"
)

type checks []*analysis.Analyzer

var c checks

var (
	builtIn = flag.Bool("builtin", false, "To add standart golang static checks")
	static  = flag.Bool("static", false, "To add SA checks from https://staticcheck.io/")
	extra   = flag.Bool("extra", false, "To add SA checks from https://github.com/go-critic/go-critic")
	total   = flag.Bool("total", false, "To all checks")
)

// Requred - os.Exit linter
func (c *checks) Requred() *checks {
	*c = append(*c, linter.OsExitAnalyzer)
	return c
}

// Builtin - add builtin analizers to check list
func (c *checks) Builtin() *checks {
	fmt.Println("Builtin checks added...")
	builtin := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	}
	*c = append(*c, builtin...)
	return c
}

// Builtin - add SA class analizers and S1006, ST1000, QF1004 checks from http://staticcheck.io/ to check list
func (c *checks) Static() *checks {
	fmt.Println("Staticchecks added...")
	saCheck := "SA"

	otherChecks := map[string]bool{
		"S1006":  true,
		"ST1000": true,
		"QF1004": true,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, saCheck) || otherChecks[v.Analyzer.Name] {
			*c = append(*c, v.Analyzer)
		}
	}
	return c
}

// Extra - add SA class analizers and S1006, ST1000, QF1004 checks from http://staticcheck.io/ to check list
func (c *checks) Extra() *checks {
	fmt.Println("Gocritic checks added...")
	*c = append(*c, analyzer.Analyzer)
	return c
}

func init() {
	c = make(checks, 0)
	c.Requred()
	flag.Parse()
	if *total {
		c.Builtin()
		c.Static()
		c.Extra()
	} else {
		if *builtIn {
			c.Builtin()
		}
		if *static {
			c.Static()
		}
		if *extra {
			c.Extra()
		}
	}
	if len(c) == 0 {
		log.Fatalln("Please, define checks")
	}
}

func main() {
	multichecker.Main(
		c...,
	)
}
