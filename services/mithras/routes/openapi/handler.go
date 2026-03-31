// Package openapi provides the OpenAPI specification for the Mithras API.
package openapi

import (
	"context"
	"net/http"

	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
)

type handler struct {
}

// New creates a new OpenAPI handler.
func New() *handler {
	return &handler{}
}

func (h *handler) Method() string { return http.MethodGet }
func (h *handler) Path() string   { return "/openapi.yaml" }

func (h *handler) Handle(_ context.Context, c *httpkit.Context) error {
	c.Res().AddHeader("Content-Type", "application/yaml")
	return c.Res().Send(http.StatusOK, api.Spec)
}
