// ================================================================ 2 ================================================================

/* package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type VarInfo struct {
	Name string `json:"name"`
	Line int    `json:"line"`
	Kind string `json:"kind"` // "var" or ":="
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-checker <filename.go>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Create a new token file set
	fset := token.NewFileSet()

	// Parse the Go source file
	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	var variables []VarInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		// Handle var declarations
		case *ast.GenDecl:
			if stmt.Tok == token.VAR {
				for _, spec := range stmt.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range valSpec.Names {
							pos := fset.Position(name.Pos())
							variables = append(variables, VarInfo{
								Name: name.Name,
								Line: pos.Line,
								Kind: "var",
							})
						}
					}
				}
			}
		// Handle := declarations
		case *ast.AssignStmt:
			if stmt.Tok == token.DEFINE {
				for _, lhs := range stmt.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						pos := fset.Position(ident.Pos())
						variables = append(variables, VarInfo{
							Name: ident.Name,
							Line: pos.Line,
							Kind: ":=",
						})
					}
				}
			}
		}
		return true
	})

	// Output as JSON
	if err := json.NewEncoder(os.Stdout).Encode(variables); err != nil {
		fmt.Fprintf(os.Stderr, "JSON encoding failed: %v\n", err)
		os.Exit(1)
	}
} */

// ================================================================ 3 ================================================================

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
)

type Violation struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func checkNaming(name, typ string) (bool, string) {
	switch typ {
	case "array":
		if !strings.HasSuffix(name, "Arr") {
			return false, fmt.Sprintf("Array variable '%s' should end with 'Arr'", name)
		}
	case "map":
		if !strings.HasSuffix(name, "Map") {
			return false, fmt.Sprintf("Map variable '%s' should end with 'Map'", name)
		}
	default:
		if !strings.HasPrefix(name, "l") {
			return false, fmt.Sprintf("Local variable '%s' should start with 'l'", name)
		}
	}
	return true, ""
}

func analyzeFile(filename string) {
	var violations []Violation
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	// Type info needed to detect map and array
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	_, _ = conf.Check("pkg", fset, []*ast.File{node}, info)

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {

		// var declarations
		case *ast.GenDecl:
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, ident := range vs.Names {
							if ident.Name == "_" {
								continue
							}
							pos := fset.Position(ident.Pos())
							varType := "var"
							if t := info.Defs[ident]; t != nil {
								if strings.HasPrefix(t.Type().String(), "[]") {
									varType = "array"
								} else if strings.HasPrefix(t.Type().String(), "map[") {
									varType = "map"
								}
							}
							if ok, msg := checkNaming(ident.Name, varType); !ok {
								violations = append(violations, Violation{
									Name:    ident.Name,
									Message: msg,
									Line:    pos.Line,
									Column:  pos.Column,
								})
							}
						}
					}
				}
			} // CASE-1
			// Check for global scope (outside of functions) by using parent node scope check
			if x.Tok == token.VAR && x.TokPos.IsValid() && x.Lparen.IsValid() {
				for _, spec := range x.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, ident := range vs.Names {
							if ident.Name == "_" {
								continue
							}
							// Check if it's global
							if info.Defs[ident] != nil && info.Defs[ident].Parent() == types.Universe {
								pos := fset.Position(ident.Pos())
								if !strings.HasPrefix(ident.Name, "G") {
									violations = append(violations, Violation{
										Name:    ident.Name,
										Message: fmt.Sprintf("Global variable '%s' should start with 'G'", ident.Name),
										Line:    pos.Line,
										Column:  pos.Column,
									})
								}
							}
						}
					}
				}
			} //  CASE-4
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if _, ok := ts.Type.(*ast.StructType); ok {
							if !strings.HasSuffix(ts.Name.Name, "Struct") {
								pos := fset.Position(ts.Name.Pos())
								violations = append(violations, Violation{
									Name:    ts.Name.Name,
									Message: fmt.Sprintf("Struct type '%s' should end with 'Struct'", ts.Name.Name),
									Line:    pos.Line,
									Column:  pos.Column,
								})
							}
						}
					}
				}
			} // CASE-5

		// := assignments
		case *ast.AssignStmt:
			if x.Tok == token.DEFINE {
				for _, lhs := range x.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
						pos := fset.Position(ident.Pos())
						varType := "var"
						if t, ok := info.Defs[ident]; ok && t.Type() != nil {
							if strings.HasPrefix(t.Type().String(), "[]") {
								varType = "array"
							} else if strings.HasPrefix(t.Type().String(), "map[") {
								varType = "map"
							}
						}
						if ok, msg := checkNaming(ident.Name, varType); !ok {
							violations = append(violations, Violation{
								Name:    ident.Name,
								Message: msg,
								Line:    pos.Line,
								Column:  pos.Column,
							})
						}
					}
				}
			} // CASE-2

		// function parameters
		case *ast.FuncDecl:
			for _, param := range x.Type.Params.List {
				for _, name := range param.Names {
					if name.Name == "_" {
						continue
					}
					if !strings.HasPrefix(name.Name, "p") {
						pos := fset.Position(name.Pos())
						violations = append(violations, Violation{
							Name:    name.Name,
							Message: fmt.Sprintf("Function parameter '%s' should start with 'p'", name.Name),
							Line:    pos.Line,
							Column:  pos.Column,
						})
					}
				}
			} // CASE-3
			// if x.Type.Params != nil && len(x.Type.Params.List) > 0 {
			// 	firstParam := x.Type.Params.List[0]
			// 	// Check if it's of the required type, e.g., *helpers.HelperStruct
			// 	if exprStr := fmt.Sprintf("%s", firstParam.Type); exprStr != "*helpers.HelperStruct" {
			// 		pos := fset.Position(firstParam.Pos())
			// 		violations = append(violations, Violation{
			// 			Name:    "", // No specific name for this rule
			// 			Message: fmt.Sprintf("First parameter of function '%s' should be of type '*helpers.HelperStruct'", x.Name.Name),
			// 			Line:    pos.Line,
			// 			Column:  pos.Column,
			// 		})
			// 	}
			// } // CASE-6

		}
		return true
	})

	// Output all violations as JSON
	_ = json.NewEncoder(os.Stdout).Encode(violations)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-checker <filename>")
		os.Exit(1)
	}
	analyzeFile(os.Args[1])
}
