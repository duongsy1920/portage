// learn/ is a Remotion (TypeScript) project, not Go code. This file exists for
// one reason: `go test ./...` at the repository root walks into node_modules
// and picks up a stray Go package vendored inside a JavaScript dependency
// (flatted/golang). Declaring a nested module makes the root module stop at
// this directory, so the Go test count stays what CLAUDE.md says it is.
//
// Nothing here is built or imported. If learn/ ever moves out of this
// repository, delete this file with it.
module github.com/duongsy/portage/learn

go 1.23
