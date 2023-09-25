package httpsvc

import (
	"github.com/irvankadhafi/go-point-of-sales/internal/delivery"
	"github.com/irvankadhafi/go-point-of-sales/internal/usecase"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type userResponse struct {
	ID        int64     `json:"id" `
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      rbac.Role `json:"role"`
	CreatedBy string    `json:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Service) handleGetCurrentLoginUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		user, err := s.userUsecase.FindByID(ctx, requester, requester.ID)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		return c.JSON(http.StatusOK, res)
	}
}

func (s *Service) handleGetUserByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		requester := delivery.GetAuthUserFromCtx(ctx)

		logger := logrus.WithFields(logrus.Fields{
			"ctx":         utils.DumpIncomingContext(ctx),
			"requesterID": requester.ID,
		})

		user, err := s.userUsecase.FindByID(ctx, requester, utils.StringToInt64(c.Param("userID")))
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		case usecase.ErrPermissionDenied:
			return ErrPermissionDenied
		default:
			logger.Error(err)
			return ErrInternal
		}

		res := userResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		return c.JSON(http.StatusOK, res)
	}
}
