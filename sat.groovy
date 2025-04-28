 *ast.File {
       Package: test.go:1:1
       Name: *ast.Ident {
          NamePos: test.go:1:9
          Name: "main"
       }
       Decls: []ast.Decl (len = 1) {
          0: *ast.GenDecl {
             TokPos: test.go:3:1
             Tok: var
             Lparen: -
             Specs: []ast.Spec (len = 1) {
                0: *ast.ValueSpec {
                   Names: []*ast.Ident (len = 1) {
                      0: *ast.Ident {
                         NamePos: test.go:3:5
                         Name: "list"
                         Obj: *ast.Object {
                            Kind: var
                            Name: "list"
                            Decl: *(obj @ 12)
                            Data: 0
                         }
                      }
                   }
                   Type: *ast.ArrayType {
                      Lbrack: test.go:3:10
                      Elt: *ast.Ident {
                         NamePos: test.go:3:12
                         Name: "int"
                      }
                   }
                }
             }
             Rparen: -
          }
       }
       FileStart: test.go:1:1
       FileEnd: test.go:3:16
       Scope: *ast.Scope {
          Objects: map[string]*ast.Object (len = 1) {
             "list": *(obj @ 17)
          }
       }
       Unresolved: []*ast.Ident (len = 1) {
          0: *(obj @ 27)
       }
       GoVersion: ""
    }