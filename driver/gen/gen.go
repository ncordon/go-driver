package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/go/loader"
)

func implements(v types.Type, in *types.Interface) bool {
	if types.Implements(v, in) {
		return true
	}
	if types.Implements(types.NewPointer(v), in) {
		return true
	}
	return false
}

func loadPackage(name string) (*loader.PackageInfo, error) {
	lc := loader.Config{
		Fset: token.NewFileSet(),
	}
	lc.Import(name)
	prog, err := lc.Load()
	if err != nil {
		return nil, err
	}
	return prog.Imported[name], nil
}

func genFile(name string, gen func(w io.Writer)) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()
	gen(file)
	return nil
}

type byName []*types.Named

func (a byName) Len() int {
	return len(a)
}

func (a byName) Less(i, j int) bool {
	return a[i].Obj().Name() < a[j].Obj().Name()
}

func (a byName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

var (
	fAnnAuto = flag.String("ann", "./normalizer/ann_gen.go", "file for auto annotations")
)

func main() {
	flag.Parse()
	pkg, err := loadPackage("go/ast")
	if err != nil {
		log.Fatal(err)
	}

	lookupType := func(name string) *types.Named {
		return pkg.Pkg.Scope().Lookup(name).(*types.TypeName).
			Type().(*types.Named)
	}
	lookupInterface := func(name string) *types.Interface {
		return lookupType(name).Underlying().(*types.Interface)
	}

	astNode := lookupInterface("Node")
	astExpr := lookupInterface("Expr")
	astStmt := lookupInterface("Stmt")

	if !implements(lookupType("Ident"), astNode) {
		var _ ast.Node = &ast.Ident{}
		panic("implements function is broken")
	}

	var (
		nodes, expr, stmt []*types.Named
		seen              = make(map[*types.Named]struct{})
	)
	for _, v := range pkg.Types {
		id, ok := v.Type.(*types.Named)
		if !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		} else if id.Obj().Pkg() != pkg.Pkg {
			continue
		}
		ut := id.Underlying()
		if _, ok := ut.(*types.Interface); ok {
			continue
		}
		seen[id] = struct{}{}
		if implements(id, astNode) {
			nodes = append(nodes, id)
		}
		if implements(id, astExpr) {
			expr = append(expr, id)
		}
		if implements(id, astStmt) {
			stmt = append(stmt, id)
		}
	}

	sort.Sort(byName(nodes))
	sort.Sort(byName(expr))
	sort.Sort(byName(stmt))

	const genHeader = "// Code generated by ./gen/gen.go DO NOT EDIT.\n\n"

	genFile(*fAnnAuto, func(w io.Writer) {
		fmt.Fprintln(w, genHeader+`package normalizer

import (
	"github.com/bblfsh/sdk/v3/uast/role"
)

var typeRoles = map[string][]role.Role{`)
		defer fmt.Fprintln(w, "}")

		genRole := func(name, role string, list []*types.Named) {
			fmt.Fprintf(w, "// all AST nodes that implements ast.%s\n", name)

			for _, n := range list {
				tp := n.Obj().Name()
				fmt.Fprintf(w, "\t%q: {role.%s},\n", tp, role)
			}
			fmt.Fprintln(w)
		}

		genRole("Expr", "Expression", expr)
		genRole("Stmt", "Statement", stmt)
	})
}
