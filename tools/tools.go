//go:build tools
// +build tools

package tools

// Manage tool dependencies via go.mod.
//
// https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// https://github.com/golang/go/issues/25922

//nolint:all
import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/vektra/mockery/v2"
	_ "golang.org/x/tools/cmd/goimports"
	_ "mvdan.cc/gofumpt"
)
