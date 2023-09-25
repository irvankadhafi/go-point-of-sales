package httpsvc

import (
	"github.com/irvankadhafi/go-point-of-sales/internal/delivery"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/usecase"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
)

type productResponse struct {
	*model.Product
	Price     string `json:"price"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Quantity    int64  `json:"quantity"`
}

func (s *Service) handleCreateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		req := createProductRequest{}
		if err := c.Bind(&req); err != nil {
			logrus.Error(err)
			return ErrInvalidArgument
		}
		if err := validate.Var(req.Price, "numeric"); err != nil {
			logger.Error(err)
			return ErrInvalidArgument
		}

		product, err := s.productUsecase.Create(ctx, requester, model.CreateProductInput{
			Name:        req.Name,
			Description: req.Description,
			Price:       utils.StringToInt64(req.Price),
			Quantity:    req.Quantity,
		})
		switch err {
		case nil:
			break
		case usecase.ErrAlreadyExist:
			return ErrProductNameAlreadyExist
		default:
			logger.Error(err)
			return httpValidationOrInternalErr(err)
		}

		return c.JSON(http.StatusCreated, setSuccessResponse(productResponse{
			Product:   product,
			Price:     utils.Int64ToRupiah(product.Price),
			CreatedAt: utils.FormatTimeRFC3339(&product.CreatedAt),
			UpdatedAt: utils.FormatTimeRFC3339(&product.UpdatedAt),
		}))
	}
}

func (s *Service) handleGetAllProducts() echo.HandlerFunc {
	type metaInfo struct {
		Size      int `json:"size"`
		Count     int `json:"count"`
		CountPage int `json:"count_page"`
		Page      int `json:"page"`
		NextPage  int `json:"next_page"`
	}

	type userCursor struct {
		Items    []productResponse `json:"items"`
		MetaInfo *metaInfo         `json:"meta_info"`
	}
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		logger := logrus.WithFields(logrus.Fields{
			"ctx": utils.DumpIncomingContext(ctx),
		})

		requester := delivery.GetAuthUserFromCtx(ctx)

		page := utils.StringToInt(c.QueryParam("page"))
		size := utils.StringToInt(c.QueryParam("size"))
		sortType := c.QueryParam("sortBy")
		query := c.QueryParam("query")
		criteria := model.ProductSearchCriteria{
			Page:     page,
			Size:     size,
			SortType: model.ProductSortType(sortType),
			Query:    query,
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

		var productResponses []productResponse
		for _, product := range products {
			productResponses = append(productResponses, productResponse{
				Product:   product,
				Price:     utils.Int64ToRupiah(product.Price),
				CreatedAt: utils.FormatTimeRFC3339(&product.CreatedAt),
				UpdatedAt: utils.FormatTimeRFC3339(&product.UpdatedAt),
			})
		}

		hasMore := int(count)-(criteria.Page*criteria.Size) > 0
		res := userCursor{
			Items: productResponses,
			MetaInfo: &metaInfo{
				Size:      size,
				Count:     int(count),
				CountPage: utils.CalculatePages(int(count), criteria.Size),
				Page:      page,
			},
		}
		if hasMore {
			res.MetaInfo.NextPage = page + 1
		}

		return c.JSON(http.StatusOK, setSuccessResponse(res))
	}
}
