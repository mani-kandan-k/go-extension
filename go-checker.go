package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"
)

type VarType string

const (
	G_VAR      VarType = "GVAR"
	L_VAR      VarType = "LVAR"
	ARR        VarType = "ARR"
	MAP        VarType = "MAP"
	STRUCT_DEF VarType = "STRUCT"
	STRUCT_VAR VarType = "STRUCT_VAR"
	CHAN       VarType = "CHAN"
	PARAM      VarType = "PARAM"
)

type Violation struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func (pRec *Violation) SetPos(pPos token.Position) {
	pRec.Column = pPos.Column
	pRec.Line = pPos.Line
}

func (pRec *Violation) SetIndet(pIdent *ast.Ident) {
	pRec.Name = pIdent.Name
}

func (pRec *Violation) SetMsg(pIndicator VarType) {
	switch pIndicator {
	case G_VAR:
		pRec.Message = fmt.Sprintf("Global variable '%s' should start with 'G'", pRec.Name)
	case L_VAR:
		pRec.Message = fmt.Sprintf("Local variable '%s' should start with 'l'", pRec.Name)
	case ARR:
		pRec.Message = fmt.Sprintf("Array variable '%s' should end with 'Arr'", pRec.Name)
	case MAP:
		pRec.Message = fmt.Sprintf("Map variable '%s' should end with 'Map'", pRec.Name)
	case STRUCT_DEF:
		pRec.Message = fmt.Sprintf("Struct '%s' should end with 'Struct'", pRec.Name)
	case STRUCT_VAR:
		pRec.Message = fmt.Sprintf("Struct variable '%s' should end with 'Rec'", pRec.Name)
	case PARAM:
		pRec.Message = fmt.Sprintf("Parameter variable '%s' should start with 'p'", pRec.Name)
	case CHAN:
		pRec.Message = fmt.Sprintf("Channel variable '%s' should end with 'Chan'", pRec.Name)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-checker-3 <filename>")
		os.Exit(1)
	}
	GoChecker(os.Args[1])
}

func GoChecker(pFileName string) {
	var lNamingViolationsArr []Violation
	lFileSet := token.NewFileSet()
	lAstFileNode, lErr := parser.ParseFile(lFileSet, pFileName, nil, parser.AllErrors)
	if lErr != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", lErr)
		os.Exit(1)
	}
	lNamingViolationsArr = ParseNode(lAstFileNode, lFileSet, lAstFileNode)
	// Output all violations as JSON
	_ = json.NewEncoder(os.Stdout).Encode(lNamingViolationsArr)
}

func ParseNode(pNode ast.Node, pFileSet *token.FileSet, pFile *ast.File) (lNamingViolationsArr []Violation) {
	lTypeConfig := types.Config{Importer: importer.Default()}
	lTypeInfo := &types.Info{
		Defs: map[*ast.Ident]types.Object{},
		Uses: map[*ast.Ident]types.Object{},
	}

	_, lErr := lTypeConfig.Check("", pFileSet, []*ast.File{pFile}, lTypeInfo)
	if lErr != nil {
		log.Println("Type checking error:", lErr)
	}

	globalVarSeen := map[string]struct{}{}

	ast.Inspect(pNode, func(n ast.Node) bool {
		switch node := n.(type) {

		// Struct Definition
		case *ast.TypeSpec:
			if _, ok := node.Type.(*ast.StructType); ok {
				if !strings.HasSuffix(node.Name.Name, "Struct") {
					var v Violation
					v.SetIndet(node.Name)
					v.SetPos(pFileSet.Position(node.Name.Pos()))
					v.SetMsg(STRUCT_DEF)
					lNamingViolationsArr = append(lNamingViolationsArr, v)
				}
			}

		// Function Parameters and Local Variables
		case *ast.FuncDecl:
			// Parameter Checks
			if node.Type.Params != nil {
				for _, param := range node.Type.Params.List {
					for _, name := range param.Names {
						if name.Name == "_" {
							continue
						}
						if !strings.HasPrefix(name.Name, "p") {
							var v Violation
							v.SetIndet(name)
							v.SetPos(pFileSet.Position(name.Pos()))
							v.SetMsg(PARAM)
							lNamingViolationsArr = append(lNamingViolationsArr, v)
						}
					}
				}
			}

			// Local Variable Checks
			if node.Body != nil {
				for _, stmt := range node.Body.List {
					switch s := stmt.(type) {
					case *ast.DeclStmt:
						if genDecl, ok := s.Decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
							for _, spec := range genDecl.Specs {
								if valSpec, ok := spec.(*ast.ValueSpec); ok {
									for _, ident := range valSpec.Names {
										if ident.Name == "_" {
											continue
										}
										if !strings.HasPrefix(ident.Name, "l") {
											var v Violation
											v.SetIndet(ident)
											v.SetPos(pFileSet.Position(ident.Pos()))
											v.SetMsg(L_VAR)
											lNamingViolationsArr = append(lNamingViolationsArr, v)
										}
										obj := lTypeInfo.Defs[ident]
										if obj != nil {
											if obj.Type() != nil {
												if v := analyzeVariable(obj.Type(), ident, pFileSet); v != nil {
													lNamingViolationsArr = append(lNamingViolationsArr, *v)
												}
											}
										}
									}
								}
							}
						}
					case *ast.AssignStmt:
						if s.Tok == token.DEFINE {
							for _, lhs := range s.Lhs {
								if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
									if !strings.HasPrefix(ident.Name, "l") {
										var v Violation
										v.SetIndet(ident)
										v.SetPos(pFileSet.Position(ident.Pos()))
										v.SetMsg(L_VAR)
										lNamingViolationsArr = append(lNamingViolationsArr, v)
									}
									obj := lTypeInfo.Defs[ident]
									if obj != nil {
										if obj.Type() != nil {
											if v := analyzeVariable(obj.Type(), ident, pFileSet); v != nil {
												lNamingViolationsArr = append(lNamingViolationsArr, *v)
											}
										}
									}
								}
							}
						}
					}
				}
			}

		// Global Variable Declaration
		case *ast.GenDecl:
			if node.Tok != token.VAR {
				break
			}
			for _, spec := range node.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, ident := range valueSpec.Names {
					if ident.Name == "_" {
						continue
					}
					obj := lTypeInfo.Defs[ident]
					if obj == nil {
						continue
					}
					if _, seen := globalVarSeen[ident.Name]; seen {
						continue
					}
					globalVarSeen[ident.Name] = struct{}{}
					if obj.Parent() == obj.Pkg().Scope() && !strings.HasPrefix(ident.Name, "G") {
						var v Violation
						v.SetIndet(ident)
						v.SetPos(pFileSet.Position(ident.Pos()))
						v.SetMsg(G_VAR)
						lNamingViolationsArr = append(lNamingViolationsArr, v)
					}
					if obj.Type() != nil {
						if v := analyzeVariable(obj.Type(), ident, pFileSet); v != nil {
							lNamingViolationsArr = append(lNamingViolationsArr, *v)
						}
					}
				}
			}
		}
		return true
	})

	return
}

func analyzeVariable(typ types.Type, ident *ast.Ident, pFileSet *token.FileSet) *Violation {

	if typ == nil {
		return nil
	}

	var v Violation
	v.SetIndet(ident)
	v.SetPos(pFileSet.Position(ident.Pos()))

	name := ident.Name
	hasViolation := false

	switch typ.Underlying().(type) {
	case *types.Array, *types.Slice:
		if !strings.HasSuffix(name, "Arr") {
			v.SetMsg(ARR)
			hasViolation = true
		}
	case *types.Map:
		if !strings.HasSuffix(name, "Map") {
			v.SetMsg(MAP)
			hasViolation = true
		}
	case *types.Struct:
		if !strings.HasSuffix(name, "Rec") {
			v.SetMsg(STRUCT_VAR)
			hasViolation = true
		}

	case *types.Chan:
		if !strings.HasSuffix(name, "Chan") {
			v.SetMsg(CHAN)
			hasViolation = true
		}
	}

	if hasViolation {
		return &v
	}
	return nil
}
