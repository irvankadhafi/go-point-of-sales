package usecase

import (
	"context"
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/helper"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"time"
)

type authUsecase struct {
	userRepo    model.UserRepository
	sessionRepo model.SessionRepository
	rbacRepo    model.RBACRepository
}

func NewAuthUsecase(
	userRepo model.UserRepository,
	sessionRepo model.SessionRepository,
	rbacRepo model.RBACRepository,
) model.AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		rbacRepo:    rbacRepo,
	}
}

func (a *authUsecase) LoginByEmailPassword(ctx context.Context, req model.LoginRequest) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"email":     req.Email,
		"userAgent": req.UserAgent,
	})

	isLocked, err := a.userRepo.IsLoginByEmailPasswordLocked(ctx, req.Email)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if isLocked {
		return nil, ErrLoginByEmailPasswordLocked
	}

	user, err := a.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	cipherPass, err := a.userRepo.FindPasswordByID(ctx, user.ID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if cipherPass == nil {
		logger.Error(err)
		return nil, errors.New("unexpected: no password found")
	}
	if !helper.IsHashedStringMatch([]byte(req.PlainPassword), cipherPass) {
		// obscure the error if the password does not match
		if err := a.userRepo.IncrementLoginByEmailPasswordRetryAttempts(ctx, req.Email); err != nil {
			logger.Error(err)
			return nil, err
		}

		return nil, ErrUnauthorized
	}

	logger = logger.WithField("userID", user.ID)

	now := time.Now()
	session := &model.Session{
		ID:                    utils.GenerateID(),
		UserID:                user.ID,
		AccessTokenExpiredAt:  now.Add(config.AccessTokenDuration()),
		RefreshTokenExpiredAt: now.Add(config.RefreshTokenDuration()),
		AppID:                 req.AppID,
		UserAgent:             req.UserAgent,
		IPAddress:             req.IPAddress,
		Role:                  user.Role,
	}

	accessToken, err := generateToken(user.ID, user.Role, session.AccessTokenExpiredAt)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	refreshToken, err := generateToken(user.ID, user.Role, session.RefreshTokenExpiredAt)
	if err != nil {
		return nil, err
	}

	session.AccessToken = accessToken
	session.RefreshToken = refreshToken
	if err := a.sessionRepo.Create(ctx, session); err != nil {
		logger.Error(err)
		return nil, err
	}

	if err := a.sessionRepo.DeleteByUserIDAndMaxRemainderSession(ctx, user.ID, config.MaxActiveSession()); err != nil {
		logger.Error(err)
		return nil, err
	}

	return session, nil
}

func (a *authUsecase) AuthenticateToken(ctx context.Context, accessToken string) (*model.User, error) {
	session, err := a.sessionRepo.FindByToken(ctx, model.AccessToken, accessToken)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	if session == nil {
		return nil, ErrNotFound
	}

	if session.IsAccessTokenExpired() {
		return nil, ErrAccessTokenExpired
	}

	perm, err := a.rbacRepo.LoadPermission(ctx)
	if err != nil {
		logrus.WithField("userID", session.UserID).Error(err)
		return nil, err
	}

	user, err := a.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		logrus.WithField("userID", session.UserID).Error(err)
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	user.SessionID = session.ID
	user.SetPermission(perm)

	return user, nil
}

func (a *authUsecase) FindRolePermission(ctx context.Context, role rbac.Role) (*rbac.RolePermission, error) {
	perm, err := a.rbacRepo.LoadPermission(ctx)
	if err != nil {
		logrus.WithField("role", role).Error(err)
		return nil, err
	}

	return rbac.NewRolePermission(role, perm), nil
}

func (a *authUsecase) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (*model.Session, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"userAgent": req.UserAgent,
	})

	session, err := a.sessionRepo.FindByToken(ctx, model.RefreshToken, req.RefreshToken)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if session == nil {
		logger.Error(ErrNotFound)
		return nil, ErrNotFound
	}

	user, err := a.userRepo.FindByID(ctx, session.UserID)
	switch {
	case err != nil:
		logger.WithField("userID", session.UserID).Error(err)
		return nil, err
	case user == nil:
		logger.WithField("userID", session.UserID).Error(ErrNotFound)
		return nil, ErrNotFound
	}

	// old session is used to delete the old session cache
	oldSess := *session

	if session.RefreshTokenExpiredAt.Before(time.Now()) {
		logger.Error(ErrRefreshTokenExpired)
		return nil, ErrRefreshTokenExpired
	}

	now := time.Now()
	session.AccessTokenExpiredAt = now.Add(config.AccessTokenDuration())
	session.RefreshTokenExpiredAt = now.Add(config.RefreshTokenDuration())

	newAccessToken, err := generateToken(session.UserID, session.Role, session.AccessTokenExpiredAt)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	newRefreshToken, err := generateToken(session.UserID, session.Role, session.RefreshTokenExpiredAt)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	session.AccessToken = newAccessToken
	session.RefreshToken = newRefreshToken
	session.UserAgent = req.UserAgent

	session, err = a.sessionRepo.RefreshToken(ctx, &oldSess, session)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return session, nil
}

func (a *authUsecase) DeleteSessionByID(ctx context.Context, sessionID int64) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"sessionID": sessionID,
	})

	session, err := a.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		logger.Error(err)
		return err
	}

	if session == nil {
		return ErrNotFound
	}

	err = a.sessionRepo.Delete(ctx, session)
	if err != nil {
		logger.Error(err)
	}

	return err
}

// generateToken and check uniqueness
func generateToken(userID int64, role rbac.Role, expTime time.Time) (string, error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": utils.Int64ToString(userID),
		"role":   role,
		"exp":    expTime,
	})

	token, err := rawToken.SignedString([]byte(config.SecretKey()))
	if err != nil {
		return "", err
	}

	return token, err
}
