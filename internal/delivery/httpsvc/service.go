package httpsvc

import (
	"github.com/irvankadhafi/go-point-of-sales/auth"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/labstack/echo/v4"
)

// Service http service
type Service struct {
	echo               *echo.Group
	authUsecase        model.AuthUsecase
	userUsecase        model.UserUsecase
	appClientUsecase   model.AppClientUsecase
	productUsecase     model.ProductUsecase
	transactionUsecase model.TransactionUsecase
	httpMiddleware     *auth.AuthenticationMiddleware
}

// RouteService add dependencies and use group for routing
func RouteService(
	echo *echo.Group,
	authUsecase model.AuthUsecase,
	userUsecase model.UserUsecase,
	appClientUsecase model.AppClientUsecase,
	productUsecase model.ProductUsecase,
	transactionUsecase model.TransactionUsecase,
	authMiddleware *auth.AuthenticationMiddleware,
) {
	srv := &Service{
		echo:               echo,
		authUsecase:        authUsecase,
		userUsecase:        userUsecase,
		appClientUsecase:   appClientUsecase,
		productUsecase:     productUsecase,
		transactionUsecase: transactionUsecase,
		httpMiddleware:     authMiddleware,
	}

	srv.initRoutes()
}

func (s *Service) initRoutes() {
	// auth
	authRoute := s.echo.Group("/auth")
	{
		authRoute.POST("/login/", s.handleLoginByEmailPassword())
		authRoute.POST("/refresh/", s.handleRefreshToken())
		authRoute.POST("/logout/", s.handleLogout(), s.httpMiddleware.MustAuthenticateAccessToken())
	}

	userRoute := s.echo.Group("/user")
	{
		userRoute.GET("/me/", s.handleGetCurrentLoginUser(), s.httpMiddleware.MustAuthenticateAccessToken())
		userRoute.GET("/:userID/", s.handleGetUserByID(), s.httpMiddleware.MustAuthenticateAccessToken())
	}

	productRoute := s.echo.Group("/products")
	{
		productRoute.GET("/:id/", s.handleGetDetailProductByID(), s.httpMiddleware.MustAuthenticateAccessToken())
		productRoute.PUT("/:id/", s.handleUpdateProductByID(), s.httpMiddleware.MustAuthenticateAccessToken())
		productRoute.DELETE("/:id/", s.handleDeleteProductByID(), s.httpMiddleware.MustAuthenticateAccessToken())
		productRoute.GET("/", s.handleGetListPaginationProducts(), s.httpMiddleware.MustAuthenticateAccessToken())
		productRoute.POST("/", s.handleCreateProduct(), s.httpMiddleware.MustAuthenticateAccessToken())
	}

	transactionRoute := s.echo.Group("/transactions")
	{
		transactionRoute.POST("/", s.handleCreateTransaction(), s.httpMiddleware.MustAuthenticateAccessToken())
	}
}
