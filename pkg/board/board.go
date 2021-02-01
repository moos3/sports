package board

import (
	"context"
	"net/http"

	rgb "github.com/robbydyer/sports/pkg/rgbmatrix-rpi"
)

// HTTPHandler is the type returned to the sportsmatrix for HTTP endpoints
type HTTPHandler struct {
	Handler func(http.ResponseWriter, *http.Request)
	Path    string
}

// Board is the interface to implement for displaying on the matrix
type Board interface {
	Name() string
	Render(ctx context.Context, matrix rgb.Matrix) error
	HasPriority() bool
	Enabled() bool
	GetHTTPHandlers() ([]*HTTPHandler, error)
}
