package static

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"

	"github.com/rizesql/mithras/internal/ui"
	"github.com/rizesql/mithras/pkg/httpkit"
)

type handler struct {
	dist http.FileSystem
	fs   http.Handler
}

func New() *handler {
	dist, err := fs.Sub(ui.Embed, "dist")
	if err != nil {
		panic(err)
	}

	distFS := http.FS(dist)
	return &handler{
		dist: distFS,
		fs:   http.FileServer(distFS),
	}
}

func (h *handler) Method() string { return http.MethodGet }
func (h *handler) Path() string   { return "/" }

func (h *handler) Handle(_ context.Context, c *httpkit.Context) error {
	reqPath := c.Req().Raw().URL.Path

	f, err := h.dist.Open(reqPath)
	if os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist) {
		c.Req().Raw().URL.Path = "/"
	} else if err == nil {
		if err := f.Close(); err != nil {
			return err
		}
	}

	h.fs.ServeHTTP(c.Res().Writer(), c.Req().Raw())
	return nil
}
