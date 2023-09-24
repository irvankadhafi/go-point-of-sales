package httpsvc

import (
	"github.com/irvankadhafi/go-point-of-sales/auth"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/labstack/echo"
)

// Service http service
type Service struct {
	echo             *echo.Group
	authUsecase      model.AuthUsecase
	userUsecase      model.UserUsecase
	httpMiddleware   *auth.AuthenticationMiddleware
	appClientUsecase model.AppClientUsecase
}

// RouteService add dependencies and use group for routing
func RouteService(
	echo *echo.Group,
	authUsecase model.AuthUsecase,
	userUsecase model.UserUsecase,
	authMiddleware *auth.AuthenticationMiddleware,
	appClientUsecase model.AppClientUsecase,
) {
	srv := &Service{
		echo:             echo,
		authUsecase:      authUsecase,
		userUsecase:      userUsecase,
		httpMiddleware:   authMiddleware,
		appClientUsecase: appClientUsecase,
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
}
