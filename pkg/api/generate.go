package api

//go:generate go run generate_bundle.go -input spec/openapi.yaml -output generated_openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config=config.yaml ./generated_openapi.yaml
