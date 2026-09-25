package service

import (
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/costguard/costguard/internal/config"
	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/model"
	"github.com/costguard/costguard/internal/repository"
	"github.com/costguard/costguard/internal/util"
)

// AuthService authenticates users and issues JWT tokens.
type AuthService struct {
	cfg    *config.Config
	users  *repository.UserRepository
	logger *slog.Logger
}

// NewAuthService builds an AuthService.
func NewAuthService(cfg *config.Config, users *repository.UserRepository, logger *slog.Logger) *AuthService {
	return &AuthService{cfg: cfg, users: users, logger: logger}
}

// Login authenticates a user and returns a token.
func (s *AuthService) Login(username, password string) (*model.User, string, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", constants.NewAppError(constants.CodeUnauthorized, "用户名或密码错误")
		}
		return nil, "", fmt.Errorf("login: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", constants.NewAppError(constants.CodeUnauthorized, "用户名或密码错误")
	}
	token, err := util.GenerateToken(s.cfg.JWTSecret, user.ID, user.Role, s.cfg.JWTExpireDuration())
	if err != nil {
		return nil, "", fmt.Errorf("login: generate token: %w", err)
	}
	s.logger.Info("user logged in", "user_id", user.ID, "username", username)
	return user, token, nil
}
