// Package docs serves the Mithras API documentation.
package docs

import (
	"context"
	"fmt"

	"github.com/rizesql/mithras/pkg/api"
	"github.com/rizesql/mithras/pkg/httpkit"
)

type handler struct {
}

// New creates a new handler for the Mithras API documentation.
func New() *handler {
	return &handler{}
}

func (h *handler) Method() string { return "GET" }
func (h *handler) Path() string   { return "/docs" }

func (h *handler) Handle(_ context.Context, c *httpkit.Context) error {

	html := fmt.Sprintf(`
<!doctype html>
<html>
  <head>
    <title>Mithras API Reference</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
  <script
    id="api-reference"
    type="application/json">
   %s
  </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>
		`, string(api.Spec))

	c.Res().AddHeader("Content-Type", "text/html")
	return c.Res().Send(200, []byte(html))
}
