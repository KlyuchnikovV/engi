package webapi

import (
	"net/http"
	"strconv"
	"time"
)

// getParam - extracting parameter from context, calls middleware and saves to 'context.queryParameters[key]'.
// After this parameter can be retrieved from context using 'context.Query' methods.
func (api *ServiceBase) getParam(
	ctx *Context,
	key string,
	convert func(string) (interface{}, error),
) error {
	var param = ctx.Request.getParam(key)
	if len(param) == 0 {
		return NewErrorResponse(http.StatusBadRequest, "parameter '%s' not found", key)
	}

	result, err := convert(param)
	if err != nil {
		return err
	}

	if result != nil {
		ctx.Request.updateParam(key, result)
	}

	return nil
}

// WithBody - takes pointer to structure and saves casted request body into context.
// Result can be retrieved from context using 'context.QueryParams.Body()'.
func (api *ServiceBase) WithBody(pointer interface{}) HandlerFunc {
	return func(ctx *Context) error {
		return ctx.Bind(pointer)
	}
}

// WithBool - queries mandatory boolean parameter from request by 'key'.
// Result can be retrieved from context using 'context.QueryParams.Bool(key)'.
func (api *ServiceBase) WithBool(key string) HandlerFunc {
	return func(ctx *Context) error {
		return api.getParam(ctx, key, func(param string) (interface{}, error) {
			return strconv.ParseBool(param)
		})
	}
}

// WithInt - queries mandatory integer parameter from request by 'key'.
// Result can be retrieved from context using 'context.QueryParams.Integer(key)'.
func (api *ServiceBase) WithInt(key string) HandlerFunc {
	return func(ctx *Context) error {
		return api.getParam(ctx, key, func(param string) (interface{}, error) {
			var (
				intBase = 10
				bitSize = 64
			)

			result, err := strconv.ParseInt(param, intBase, bitSize)
			if err != nil {
				return nil, NewErrorResponse(http.StatusBadRequest, "parameter '%s' not of type int", key)
			}

			return result, err
		})
	}
}

// WithFloat - queries mandatory floating point number parameter from request by 'key'.
// Result can be retrieved from context using 'context.QueryParams.Float(key)'.
func (api *ServiceBase) WithFloat(key string) HandlerFunc {
	return func(ctx *Context) error {
		return api.getParam(ctx, key, func(param string) (interface{}, error) {
			var bitSize = 64

			result, err := strconv.ParseFloat(param, bitSize)
			if err != nil {
				return nil, NewErrorResponse(http.StatusBadRequest, "parameter '%s' not of type float", key)
			}

			return result, err
		})
	}
}

// WithString - queries mandatory string parameter from request by 'key'.
// Result can be retrieved from context using 'context.QueryParams.String(key)'.
func (api *ServiceBase) WithString(key string) HandlerFunc {
	return func(ctx *Context) error {
		return api.getParam(ctx, key, func(param string) (interface{}, error) {
			return param, nil
		})
	}
}

// WithTime - queries mandatory time parameter from request by 'key' using 'layout'.
// Result can be retrieved from context using 'context.QueryParams.Time(key, layout)'.
func (api *ServiceBase) WithTime(key, layout string) HandlerFunc {
	return func(ctx *Context) error {
		return api.getParam(ctx, key, func(param string) (interface{}, error) {
			result, err := time.Parse(layout, param)
			if err != nil {
				return nil, NewErrorResponse(http.StatusBadRequest,
					"could not parse '%s' param to datetime using '%s' layout", key, layout,
				)
			}

			return result, err
		})
	}
}
