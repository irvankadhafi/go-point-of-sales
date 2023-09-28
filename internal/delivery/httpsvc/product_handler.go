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

// Endpoint CreateProduct
//	@Summary	Store a product
//	@Description
//	@Tags		Product
//	@Accept		json
//	@Produce	json
//	@Param		Accept			header		string						false	"Example: application/json"
//	@Param		Authorization	header		string						true	"Use Token: Bearer {token}"
//	@Param		Content-Type	header		string						false	"Example: application/json"
//	@Param		Body			body		model.CreateProductInput	true	"payload"
//	@Success	200				{object}	model.ProductResponse
//	@Router		/products [post]
func (s *Service) handleCreateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		req := model.CreateProductInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}
		if err := validate.Var(req.Price, "numeric"); err != nil {
			logger.Error(err)
			return ErrInvalidArgument
		}

		product, err := s.productUsecase.Create(ctx, requester, req)
		switch err {
		case nil:
			break
		case usecase.ErrAlreadyExist:
			return ErrProductNameAlreadyExist
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.JSON(http.StatusCreated, setSuccessResponse(product.ToProductResponse()))
	}
}

// Endpoint Get List Pagination of Products
//	@Summary	Endpoint for get list pagination of products
//	@Description
//	@Tags		Product
//	@Accept		json
//	@Produce	json
//	@Param		Accept			header		string						false	"Example: application/json"
//	@Param		Authorization	header		string						true	"Use Token: Bearer {token}"
//	@Param		Content-Type	header		string						false	"Example: application/json"
//	@Param		request			query		model.ProductSearchCriteria	false	"Query Params"
//	@Success	200				{object}	paginationResponse[[]model.ProductResponse]{items=[]model.ProductResponse}
//	@Router		/products [get]
func (s *Service) handleGetListPaginationProducts() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		var criteria model.ProductSearchCriteria
		if err := c.Bind(&criteria); err != nil {
			return err
		}

		criteria.SetDefaultValue()

		products, count, err := s.productUsecase.Search(ctx, requester, criteria)
		switch err {
		case nil:
			break
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		response := toResourcePaginationResponse(criteria.Page, criteria.Size, count, products.ToListProductResponse())
		return c.JSON(http.StatusOK, setSuccessResponse(response))
	}
}

// Endpoint Get Detail Product By ID
//	@Summary	Endpoint for get detail product by id
//	@Description
//	@Tags		Product
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string	true	"Use Token from Auth Service : Bearer {token}"
//	@Param		Accept			header		string	false	"Example: application/json"
//	@Param		Content-Type	header		string	false	"Example: application/json"
//	@Param		id				path		int		true	"Example: 1"
//	@Success	200				{object}	model.ProductResponse
//	@Router		/products/{id} [get]
func (s *Service) handleGetDetailProductByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		product, err := s.productUsecase.FindByID(ctx, requester, utils.StringToInt64(c.Param("id")))
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.JSON(http.StatusOK, setSuccessResponse(product.ToProductResponse()))
	}
}

// Endpoint Update Product By ID
//	@Summary	Endpoint for update product by ID
//	@Description
//	@Tags		Product
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string						true	"Use Token from Auth Service : Bearer {token}"
//	@Param		Accept			header		string						false	"Example: application/json"
//	@Param		Content-Type	header		string						false	"Example: application/json"
//	@Param		id				path		int							true	"Example: 1"
//	@Param		Body			body		model.UpdateProductInput	true	"payload"
//	@Success	200				{object}	model.ProductResponse
//	@Router		/products/{id} [put]
func (s *Service) handleUpdateProductByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		req := model.UpdateProductInput{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}

		if err := validate.Var(req.Price, "numeric"); err != nil {
			logger.Error(err)
			return ErrInvalidArgument
		}

		id := utils.StringToInt64(c.Param("id"))
		product, err := s.productUsecase.UpdateByID(ctx, requester, id, req)
		switch err {
		case nil:
			break
		case usecase.ErrNotFound:
			return ErrNotFound
		case usecase.ErrAlreadyExist:
			return ErrProductNameAlreadyExist
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.JSON(http.StatusCreated, setSuccessResponse(product.ToProductResponse()))
	}
}

// Endpoint Delete Product By ID
//	@Summary	Endpoint for delete product by ID
//	@Description
//	@Tags		Product
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string	true	"Use Token from Auth Service : Bearer {token}"
//	@Param		Accept			header		string	false	"Example: application/json"
//	@Param		Content-Type	header		string	false	"Example: application/json"
//	@Param		id				path		int		true	"Example: 1"
//	@Success	200				{object}	successResponse
//	@Router		/products/{id} [delete]
func (s *Service) handleDeleteProductByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		id := utils.StringToInt64(c.Param("id"))
		err := s.productUsecase.DeleteByID(ctx, requester, id)
		switch err {
		case nil:
			return c.JSON(http.StatusOK, setSuccessResponse(nil))
		case usecase.ErrNotFound:
			return ErrNotFound
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

	}
}
