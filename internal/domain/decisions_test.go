// Package domain_test holds DECISION GUARDS.
//
// These are not unit tests. A unit test says "this function returns the right
// answer". These say "this codebase still obeys the rules we agreed on", and
// they exist because of a real failure: CategoryPolicy was given an hsCode and
// a dutyRate field days after we had written down, twice, that the shipping
// lane in use bundles duty into the per-kg price and never itemises it. Every
// unit test passed. The design was wrong anyway.
//
// A rule that lives only in a document is a rule someone forgets. A rule that
// lives in a test is a rule that fails the build.
//
// Each guard below names the decision, points at where it was written down,
// and says what to do if you are here because it went red — including "the
// decision changed, so change the guard". These are not commandments; they are
// tripwires that force the change to be deliberate.
package domain_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sourceFile is one parsed non-test file under internal/domain.
type sourceFile struct {
	path string // relative, e.g. "catalog/category.go"
	pkg  string // e.g. "catalog"
	ast  *ast.File
}

// domainSources parses every non-test .go file under internal/domain.
//
// Comments are deliberately NOT parsed: a guard that trips on prose would
// fire on the sentence explaining why the rule exists. We check code.
func domainSources(t *testing.T) []sourceFile {
	t.Helper()
	fset := token.NewFileSet()
	var out []sourceFile

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0) // 0 = no comments
		if err != nil {
			return err
		}
		out = append(out, sourceFile{
			path: filepath.ToSlash(path),
			pkg:  f.Name.Name,
			ast:  f,
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking internal/domain: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("no source files found — is this test in the wrong directory?")
	}
	return out
}

// eachIdent visits every identifier and string literal in the file's code.
func eachIdent(f sourceFile, fn func(name string, pos token.Pos)) {
	ast.Inspect(f.ast, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.Ident:
			fn(v.Name, v.Pos())
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				fn(strings.Trim(v.Value, "`\""), v.Pos())
			}
		}
		return true
	})
}

// ── Guard 1 ─────────────────────────────────────────────────────────────────
// DECISION: money is int64 minor units; floating point never touches it.
// WRITTEN DOWN: docs/SETUP.md convention 2, docs/DDD.md §11.
// WHY: 0.1 + 0.2 != 0.3 in binary floating point. A cent per row, times a
// million rows, is real money — and the error is invisible until an audit.
func TestDecision_noFloatingPointInDomain(t *testing.T) {
	for _, f := range domainSources(t) {
		eachIdent(f, func(name string, pos token.Pos) {
			if name == "float32" || name == "float64" {
				t.Errorf("%s uses %s.\n"+
					"  Money is int64 minor units (SETUP.md convention 2).\n"+
					"  If this is genuinely not money, say so here and add the exception.",
					f.path, name)
			}
		})
	}
}

// ── Guard 2 ─────────────────────────────────────────────────────────────────
// DECISION: the domain does not read the clock. Methods take now time.Time.
// WRITTEN DOWN: docs/SETUP.md convention 7.
// WHY: a domain that calls time.Now() cannot be tested at a chosen moment —
// "what does this order look like three days after the deposit" becomes
// untestable, and deadline logic is exactly what needs testing.
func TestDecision_domainDoesNotReadTheClock(t *testing.T) {
	for _, f := range domainSources(t) {
		ast.Inspect(f.ast, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "time" {
				return true
			}
			if sel.Sel.Name == "Now" || sel.Sel.Name == "Since" {
				t.Errorf("%s calls time.%s.\n"+
					"  The domain takes `now time.Time` as a parameter instead\n"+
					"  (SETUP.md convention 7). The app layer owns the clock.",
					f.path, sel.Sel.Name)
			}
			return true
		})
	}
}

// ── Guard 3 ─────────────────────────────────────────────────────────────────
// DECISION: merchants and brands are DATA. No brand name appears in code.
// WRITTEN DOWN: docs/DDD.md §23, package comment in catalog/merchant.go.
// WHY: adding Adidas must be inserting a row, not deploying a release. The
// first `if brand == "Nike"` is the one that makes every later one look fine.
func TestDecision_noBrandNamesInCode(t *testing.T) {
	brands := []string{"nike", "adidas", "sony", "northface", "thenorthface", "zara", "uniqlo"}

	for _, f := range domainSources(t) {
		eachIdent(f, func(name string, pos token.Pos) {
			flat := strings.ToLower(strings.NewReplacer(" ", "", "-", "", "_", "").Replace(name))
			for _, b := range brands {
				if strings.Contains(flat, b) {
					t.Errorf("%s mentions %q in CODE (identifier or string literal).\n"+
						"  A merchant is a row in a table (DDD.md §23).\n"+
						"  Comments may name brands; code may not.",
						f.path, name)
				}
			}
		})
	}
}

// ── Guard 4 ─────────────────────────────────────────────────────────────────
// DECISION: catalog says what a thing IS. It knows nothing about duty, tax,
// customs or HS tariff codes.
// WRITTEN DOWN: docs/CATALOG.md, "Đã gỡ khỏi CategoryPolicy".
// WHY: this is the guard that would have caught the real mistake. Duty is a
// function of (goods, SHIPPING LANE) — the lane in use bundles it into the
// per-kg price and never itemises it, so a duty rate stored on a category is a
// number that looks authoritative and is ignored. It belongs to pricing,
// attached to a ShippingLane.
//
// IF THIS GOES RED: do not delete the guard. Either the field belongs in
// pricing, or the business now runs a self-managed commercial import — and in
// that case update docs/CATALOG.md first, then move this guard.
func TestDecision_catalogKnowsNothingAboutTax(t *testing.T) {
	forbidden := []string{"duty", "tariff", "hscode", "customs", "vat", "salestax"}

	for _, f := range domainSources(t) {
		if f.pkg != "catalog" {
			continue
		}
		eachIdent(f, func(name string, pos token.Pos) {
			flat := strings.ToLower(strings.NewReplacer(" ", "", "-", "", "_", "").Replace(name))
			for _, bad := range forbidden {
				if strings.Contains(flat, bad) {
					t.Errorf("%s mentions %q in code.\n"+
						"  Duty depends on the SHIPPING LANE, not on the goods:\n"+
						"    forwarder lane  -> bundled into the per-kg price, no duty line\n"+
						"    commercial lane -> itemised from an HS code\n"+
						"  So it belongs to pricing/ShippingLane, not catalog (CATALOG.md).",
						f.path, name)
				}
			}
		})
	}
}

// ── Guard 5 ─────────────────────────────────────────────────────────────────
// DECISION: aggregates expose behaviour, not setters.
// WRITTEN DOWN: docs/DDD.md §9 (anemic model), §13 (invariants).
// WHY: the whole reason this project is not the old Symfony Booking entity.
// `SetStatus("bananas")` is what an anemic model allows; Confirm() is what a
// rich one allows, and Confirm() can say no.
func TestDecision_noSettersOnDomainTypes(t *testing.T) {
	for _, f := range domainSources(t) {
		for _, decl := range f.ast.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "Set") {
				t.Errorf("%s declares method %s.\n"+
					"  Aggregates change through named behaviour (Rename, Confirm,\n"+
					"  EnableSourcing), never through setters (DDD.md §9).\n"+
					"  A setter cannot refuse; a behaviour can.",
					f.path, fn.Name.Name)
			}
		}
	}
}

// ── Guard 6 ─────────────────────────────────────────────────────────────────
// DECISION: the domain imports the standard library and one utility. Nothing
// else — no database driver, no HTTP framework, no SDK.
// WRITTEN DOWN: docs/DDD.md §21, .github/workflows/ci.yml.
// WHY: this is what keeps the whole test suite running in under a second with
// no Docker, no API key and no network. Lose it once and it never comes back.
//
// CI runs the same check with `go list`. It is repeated here so a local
// `go test ./...` catches it before the push, not after.
func TestDecision_domainImportsOnlyStdlibAndAllowlist(t *testing.T) {
	allowed := map[string]bool{
		// A pure utility, in the sense symfony/uid is one: it knows nothing
		// about databases, transport or frameworks. See catalog/merchant.go.
		"github.com/google/uuid": true,
	}

	for _, f := range domainSources(t) {
		for _, imp := range f.ast.Imports {
			path := strings.Trim(imp.Path.Value, `"`)

			switch {
			case !strings.Contains(strings.SplitN(path, "/", 2)[0], "."):
				continue // no dot in the first segment => standard library
			case strings.HasPrefix(path, "github.com/duongsy/portage/internal/domain/"):
				continue // another domain package: allowed, but see Guard 7
			case allowed[path]:
				continue
			}
			t.Errorf("%s imports %q.\n"+
				"  internal/domain may import the standard library, other domain\n"+
				"  packages, and the allowlist in this test — nothing else\n"+
				"  (DDD.md §21). Put the dependency behind an interface here and\n"+
				"  implement it in internal/adapter.",
				f.path, path)
		}
	}
}

// ── Guard 7 ─────────────────────────────────────────────────────────────────
// DECISION: bounded contexts do not import each other. Only shared is common.
// WRITTEN DOWN: docs/DDD.md §6, §8.
// WHY: the moment catalog imports pricing, the two stop being separate
// contexts and the "Product means four different things" problem comes back.
// Contexts talk through ids and domain events, not through Go imports.
func TestDecision_boundedContextsDoNotImportEachOther(t *testing.T) {
	const prefix = "github.com/duongsy/portage/internal/domain/"

	for _, f := range domainSources(t) {
		for _, imp := range f.ast.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(path, prefix) {
				continue
			}
			target := strings.SplitN(strings.TrimPrefix(path, prefix), "/", 2)[0]
			if target == "shared" || target == f.pkg {
				continue
			}
			t.Errorf("%s (context %q) imports context %q.\n"+
				"  Only the shared kernel is common (DDD.md §6, §8).\n"+
				"  Refer to the other context by id, and react to its domain\n"+
				"  events — do not reach into its model.",
				f.path, f.pkg, target)
		}
	}
}
