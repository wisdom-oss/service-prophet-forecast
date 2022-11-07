// Package errors
// This package contains all errors the service may return to HTTP requests
// Those errors include unauthenticated calls and forbidden ones
package errors

import (
	"fmt"
	"net/http"

	"microservice/structs"
	"microservice/vars"
)

const UnauthorizedRequest = "UNAUTHORIZED_REQUEST"
const MissingScope = "SCOPE_MISSING"
const UnsupportedHTTPMethod = "UNSUPPORTED_METHOD"
const DatabaseQueryError = "DATABASE_QUERY_ERROR"
const UnprocessableEntity = "UNPROCESSABLE_ENTITY"
const UniqueConstraintViolation = "UNIQUE_CONSTRAINT_VIOLATION"
const MissingRequiredParameter = "MISSING_REQUEST_PARAMETER"
const DataWriting = "INTERNAL_DATA_WRITE_ERROR"
const RScriptError = "R_SCRIPT_EXECUTION_ERROR"

var errorTitle = map[string]string{
	UnauthorizedRequest:       "Unauthorized Request",
	MissingScope:              "Forbidden",
	UnsupportedHTTPMethod:     "Unsupported HTTP Method",
	DatabaseQueryError:        "Database Query Error",
	UnprocessableEntity:       "Unprocessable Entity",
	UniqueConstraintViolation: "Unique Constraint Violation",
	MissingRequiredParameter:  "Missing required request parameter",
	DataWriting:               "Internal data writing error",
	RScriptError:              "R Script execution error",
}

var errorDescription = map[string]string{
	UnauthorizedRequest: "The resource you tried to access requires authorization. Please check your request",
	MissingScope: "You tried to access a resource which is protected by a scope. " +
		"Your authorization information did not contain the required scope.",
	UnsupportedHTTPMethod: "The used HTTP method is not supported by this microservice. " +
		"Please check the documentation for further information",
	DatabaseQueryError: "The microservice was unable to successfully execute the database query. " +
		"Please check the logs for more information",
	UnprocessableEntity: "The JSON object you sent to the service is not processable. Please check your request",
	UniqueConstraintViolation: "The object you are trying to create already exists in the database. " +
		"Please check your request and the documentation",
	MissingRequiredParameter: "The request you sent was not processable, " +
		"due to a missing parameter in the request. Please check the documentation",
	DataWriting:  "A file needed for making the prognosis could not be written",
	RScriptError: "An error occurred while trying to execute the underlying R script. Please check the server logs",
}

var httpStatus = map[string]int{
	UnauthorizedRequest:       http.StatusUnauthorized,
	MissingScope:              http.StatusForbidden,
	UnsupportedHTTPMethod:     http.StatusMethodNotAllowed,
	DatabaseQueryError:        http.StatusInternalServerError,
	UnprocessableEntity:       http.StatusUnprocessableEntity,
	UniqueConstraintViolation: http.StatusConflict,
	MissingRequiredParameter:  http.StatusBadRequest,
	DataWriting:               http.StatusInternalServerError,
	RScriptError:              http.StatusInternalServerError,
}

func NewRequestError(errorCode string) structs.RequestError {
	return structs.RequestError{
		HttpStatus:       httpStatus[errorCode],
		HttpError:        http.StatusText(httpStatus[errorCode]),
		ErrorCode:        fmt.Sprintf("%s.%s", vars.ServiceName, errorCode),
		ErrorTitle:       errorTitle[errorCode],
		ErrorDescription: errorDescription[errorCode],
	}
}
