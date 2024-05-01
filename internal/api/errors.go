package api

import (
	"net/http"

	"github.com/go-chi/render"
)

// ErrResponse is the base error type which encapsulates all returned errors.

// @Description Error object encapsulating
// @Description all returned API errors.
type ErrResponse struct {
	Err            error `json:"-"`
	HTTPStatusCode int   `json:"-"`

	StatusText string `json:"status"`          // A terse error description.
	ErrorText  string `json:"error,omitempty"` // A more detailed error description.
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrNotFound() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     http.StatusText(http.StatusNotFound),
	}
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func (h *Handlers) ErrInternalServer(err error) render.Renderer {
	h.logger.Error("internal server error", "err", err)
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 500,
		StatusText:     "Internal Server Error.",
		ErrorText:      err.Error(),
	}
}
