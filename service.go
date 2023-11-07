package engi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/KlyuchnikovV/engi/internal"
	"github.com/KlyuchnikovV/engi/internal/context"
	"github.com/KlyuchnikovV/engi/internal/request"
	"github.com/KlyuchnikovV/engi/internal/types"
	"github.com/KlyuchnikovV/engi/response"
	"github.com/KlyuchnikovV/logist"
)

type (
	ServiceAPI interface {
		// Prefix - prefix of all paths for this service.
		Prefix() string

		// Routers returns the handlers and their relative paths (relative to the service) for registration.
		Routers() Routes
	}

	MiddlewaresAPI interface {
		Middlewares() []Middleware
	}
)

type (
	// Service - provides basic service methods.
	Service struct {
		middlewares []request.Middleware

		handlers map[string]internal.HanlderNode

		marshaler types.Marshaler
		responser types.Responser

		logger *logist.Logist

		api  ServiceAPI
		path string
	}

	RouteByPath func(*Service, string) error
	Middleware  func(*Service)
	Routes      map[string]RouteByPath
)

func NewService(engine *Engine, api ServiceAPI, path string) *Service {
	return &Service{
		handlers: make(map[string]internal.HanlderNode),

		logger:    engine.logger,
		marshaler: engine.responseMarshaler,
		responser: engine.responseObject,

		api:  api,
		path: path,
	}
}

// Serve should write reply headers and data to the ResponseWriter
// and then return. Returning signals that the request is finished; it
// is not valid to use the ResponseWriter or read from the
// Request.Body after or concurrently with the completion of the
// Serve call.
func (srv *Service) Serve(uri string, r *http.Request, w http.ResponseWriter) {
	var ctx = context.NewContext(w, r, srv.marshaler, srv.responser)

	for _, middleware := range srv.middlewares {
		err := middleware.Handle(ctx.Request, ctx.Response.ResponseWriter())
		if err == nil {
			continue
		}

		var response response.AsObject

		if ok := errors.As(err, &response); !ok {
			//  TODO:
		}

		switch response.Code {
		case http.StatusOK:
			err = ctx.Response.OK(response.Code)
		default:
			err = ctx.Response.Error(response.Code, response.ErrorString)
		}

		if err != nil {
			srv.logger.Error(err.Error())
		}

		return
	}

	uri, _ = strings.CutPrefix(strings.Trim(uri, "/"), srv.api.Prefix())

	if _, ok := srv.handlers[r.Method]; !ok {
		if err := ctx.Response.NotFound("method '%s' not appliable for '%s'", r.Method, r.URL.Path); err != nil {
			srv.logger.Error(err.Error())
		}

		return
	}

	if !srv.handlers[r.Method].Handle(uri, ctx) {
		if err := ctx.Response.NotFound("path '%s' not found for method '%s'", r.URL.Path, r.Method); err != nil {
			srv.logger.Error(err.Error())
		}
	}
}

// add - creates route with custom method and path.
func (srv *Service) add(
	method, path string,
	route Route,
	middlewares ...request.Middleware,
) error {
	if _, ok := srv.handlers[method]; !ok {
		srv.handlers[method] = internal.NewStringHandler("", nil)
	}

	for _, middleware := range middlewares {
		validator, ok := middleware.(request.ParamsValidator)
		if !ok {
			continue
		}

		if err := validator.Validate(path); err != nil {
			return fmt.Errorf("%w, service: %s", err, srv.api.Prefix())
		}
	}

	srv.handlers[method].Add(
		srv.handle(route, middlewares...),
		strings.Split(path, "/")...,
	)

	return nil
}

func (srv *Service) handle(route Route, middlewares ...request.Middleware) context.Handler {
	return func(ctx *context.Context) error {
		var response response.AsObject

		for _, middleware := range middlewares {
			if err := middleware.Handle(ctx.Request, ctx.Response.ResponseWriter()); err != nil {
				if ok := errors.As(err, &response); !ok {
					//  TODO:
				}

				if err := ctx.Response.Error(response.Code, response.ErrorString); err != nil {
					srv.logger.Error(err.Error())
					return err
				}

				return err
			}
		}

		if err := route(ctx); err != nil {
			if err := ctx.Response.InternalServerError(err.Error()); err != nil {
				srv.logger.Error(err.Error())
				return err
			}
		}

		return nil
	}
}
