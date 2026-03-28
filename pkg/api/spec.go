package api

import _ "embed"

// Spec is the OpenAPI specification for the service
// Loaded from the generated_openapi.yaml file and embedded into the binary
//
//go:embed generated_openapi.yaml
var Spec []byte
