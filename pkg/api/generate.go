package api

// Bundle the multi-file OpenAPI spec into a single file
//go:generate go run generate_bundle.go -input spec/openapi.yaml -output generated_openapi.yaml

// Generate the Go types and interfaces using the local tool
//go:generate go tool oapi-codegen -config=config.yaml ./generated_openapi.yaml
