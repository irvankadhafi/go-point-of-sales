package httpsvc

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	ErrInvalidArgument            = echo.NewHTTPError(http.StatusBadRequest, setErrorMessage("invalid argument"))
	ErrEmailPasswordNotMatch      = echo.NewHTTPError(http.StatusUnauthorized, setErrorMessage("email or password not match"))
	ErrLoginByEmailPasswordLocked = echo.NewHTTPError(http.StatusLocked, setErrorMessage("user is locked from logging in using email and password"))
	ErrPermissionDenied           = echo.NewHTTPError(http.StatusForbidden, setErrorMessage("permission denied"))
	ErrInternal                   = echo.NewHTTPError(http.StatusInternalServerError, setErrorMessage("internal system error"))
	ErrUnauthenticated            = echo.NewHTTPError(http.StatusUnauthorized, setErrorMessage("unauthenticated"))
	ErrNotFound                   = echo.NewHTTPError(http.StatusNotFound, setErrorMessage("record not found"))
	ErrProductNameAlreadyExist    = echo.NewHTTPError(http.StatusBadRequest, setErrorMessage("product name already exist"))
)

// httpValidationOrInternalErr return valdiation or internal error
func httpValidationOrInternalErr(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Jika tidak ada kesalahan validasi, mengembalikan kesalahan internal
		return ErrInternal
	}

	fields := make(map[string]string)
	for _, validationError := range validationErrors {
		tag := validationError.Tag()
		fields[validationError.Field()] = fmt.Sprintf("Failed on the '%s' tag", tag)
	}

	return echo.NewHTTPError(http.StatusBadRequest, setErrorMessage(fields))
}
