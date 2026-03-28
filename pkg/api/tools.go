//go:build tools
// +build tools

package api

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/pb33f/libopenapi"
	_ "github.com/pb33f/libopenapi/bundler"
)
