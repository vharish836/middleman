package handler

import (
	"net/http"

	"github.com/vharish836/middleman/context"
)

// Service ....
type Service interface {
	// BuildAPIContext validates and extracts the data
	// from request into returning Context. In case
	// of error, proper status code is set using given writer
	// and nil context is returned
	BuildAPIContext(*http.Request, http.ResponseWriter) (context.Context, error)
	// APIHanlder takes the built Context and processes
	// the same. The response is built into the context
	// In case of error, same is stored in the context
	APIHandler(context.Context) context.Context
	// WriteResponse builds the response from given context
	// and sends using the given writer
	WriteResponse(context.Context, http.ResponseWriter)	
}

// Handler ...
type Handler struct {
	service Service
}

// NewHandler ...
func NewHandler(s Service) *Handler {
	return &Handler{s}
}

func (h Handler) ServeHTTP(s http.ResponseWriter, r *http.Request) {
	ctx, err := h.service.BuildAPIContext(r, s)
	if err != nil {
		return
	}
	ctx = h.service.APIHandler(ctx)
	h.service.WriteResponse(ctx,s)
	return
}
