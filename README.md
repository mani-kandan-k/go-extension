# Go Naming Convention Checker

This VS Code extension checks naming conventions for variables in Go source files.

### Supported Checks:
- Local variables must start with `l`
- Arrays must end with `Arr`
- Maps must end with `Map`
- Structs must end with `Rec`
- Channels must end with `Chan`

### How it Works
On file save, the extension runs a Go program (`go-checker.go`) that parses the file and reports violations.

### Requirements
- Go installed
- `go-checker.go` must be in the root of the extension

---

Created by Mani 🔍
