package httpsvc

import (
	"encoding/base64"
	"github.com/irvankadhafi/go-point-of-sales/internal/delivery"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/usecase"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type loginResponse struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	TokenType             string `json:"token_type"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
}

func (s *Service) handleLoginByEmailPassword() echo.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(c echo.Context) error {
		req := request{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		client, err := s.getAppClient(c)
		if err != nil {
			logrus.Error(err)
			return err
		}

		session, err := s.authUsecase.LoginByEmailPassword(c.Request().Context(), model.LoginRequest{
			AppID:         client.ID,
			Email:         req.Email,
			PlainPassword: req.Password,
			UserAgent:     c.Request().UserAgent(),
		})
		switch err {
		case nil:
			break
		case usecase.ErrNotFound, usecase.ErrUnauthorized:
			return ErrEmailPasswordNotMatch
		case usecase.ErrLoginByEmailPasswordLocked:
			return ErrLoginByEmailPasswordLocked
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		default:
			logrus.Error(err)
			return ErrInternal
		}

		res := loginResponse{
			AccessToken:           session.AccessToken,
			AccessTokenExpiresAt:  utils.FormatTimeRFC3339(&session.AccessTokenExpiredAt),
			RefreshToken:          session.RefreshToken,
			RefreshTokenExpiresAt: utils.FormatTimeRFC3339(&session.RefreshTokenExpiredAt),
			TokenType:             "Bearer",
		}

		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleRefreshToken() echo.HandlerFunc {
	type request struct {
		RefreshToken string `json:"refresh_token"`
	}

	return func(c echo.Context) error {
		req := request{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		session, err := s.authUsecase.RefreshToken(c.Request().Context(), model.RefreshTokenRequest{
			RefreshToken: req.RefreshToken,
			UserAgent:    c.Request().UserAgent(),
		})
		switch err {
		case nil:
		case usecase.ErrRefreshTokenExpired, usecase.ErrNotFound:
			return ErrUnauthenticated
		default:
			logrus.Error(err)
			return ErrInternal
		}

		res := loginResponse{
			AccessToken:           session.AccessToken,
			AccessTokenExpiresAt:  utils.FormatTimeRFC3339(&session.AccessTokenExpiredAt),
			RefreshToken:          session.RefreshToken,
			RefreshTokenExpiresAt: utils.FormatTimeRFC3339(&session.RefreshTokenExpiredAt),
			TokenType:             "Bearer",
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleLogout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		err := s.authUsecase.DeleteSessionByID(c.Request().Context(), requester.SessionID)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logrus.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func (s *Service) getAppClient(c echo.Context) (*model.AppClient, error) {
	clientID, clientSecret, err := s.parseBasicAuth(c.Request())
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	client, err := s.appClientUsecase.FindClient(c.Request().Context(), clientID, clientSecret)
	if err != nil {
		logrus.Error(err)
		return nil, echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	return client, nil
}

// get credentials from headers, Authorization: Basic <credentials>
// where credentials is the Base64 encoding of ClientID and ClientSecret joined by a single colon `:`
func (s *Service) parseBasicAuth(req *http.Request) (clientID, clientSecret string, err error) {
	authHeaders := strings.Split(req.Header.Get("Authorization"), " ")
	if (len(authHeaders) != 2) || (authHeaders[0] != "Basic") {
		err = ErrPermissionDenied
		return
	}

	decodedCredentialsTokenByte, err := base64.StdEncoding.DecodeString(strings.TrimSpace(authHeaders[1]))
	if err != nil {
		logrus.WithField("headers", authHeaders).Error(err)
		return
	}

	credentials := strings.Split(string(decodedCredentialsTokenByte), ":")
	if len(credentials) != 2 {
		logrus.WithField("basic token", string(decodedCredentialsTokenByte)).Error(err)
		err = ErrInvalidArgument
		return
	}
	clientID = credentials[0]
	clientSecret = credentials[1]

	return
}
