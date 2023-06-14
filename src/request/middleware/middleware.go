package middleware

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

// AdditionalResponseHeaders puts the request id and other aditional headers into the response
func AdditionalResponseHeaders(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			requestID := middleware.GetReqID(request.Context())
			responseWriter.Header().Set("X-Request-ID", requestID)
			nextHandler.ServeHTTP(responseWriter, request)
		},
	)
}

// ParseQueryParametersToContext reads all query parameters from the request URL and puts it into the context of the
// request
func ParseQueryParametersToContext(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			// get the request context
			ctx := request.Context()
			// iterate through the query parameters and add the values to the request context
			for key, value := range request.URL.Query() {
				ctx = context.WithValue(ctx, key, value)
			}
			// now serve the request to the next handler
			nextHandler.ServeHTTP(responseWriter, request.WithContext(ctx))
		},
	)
}
