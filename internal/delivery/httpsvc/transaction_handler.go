package httpsvc

import (
	"github.com/irvankadhafi/go-point-of-sales/internal/delivery"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/usecase"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (s *Service) handleCreateTransaction() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		req := model.CreateTransactionInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		transaction, err := s.transactionUsecase.Create(ctx, requester, req)
		switch err {
		case nil:
			break
		case usecase.ErrAlreadyExist:
			return ErrProductNameAlreadyExist
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.JSON(http.StatusCreated, setSuccessResponse(transaction))
	}
}
