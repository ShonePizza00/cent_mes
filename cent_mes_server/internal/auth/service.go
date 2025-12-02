package auth

import (
	"cent_mes_server/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var (
	ErrUserExists = errors.New("user already exists")
	ErrNoUser     = errors.New("no user with given credentials")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterUser(ctx context.Context, user *models.User) error {
	_, err := s.repo.GetUserByLogin(ctx, user.Login)
	if err == nil {
		return ErrUserExists
	}
	token_encoder := sha256.New()
	token_encoder.Write([]byte(user.Login))
	token_encoder.Write([]byte(user.Password))
	user.Token = hex.EncodeToString(token_encoder.Sum(nil))
	passwd_hash := sha256.Sum256([]byte(user.Password))
	user.Password = hex.EncodeToString(passwd_hash[:])
	return s.repo.RegisterUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, user *models.User) error {
	db_user, err := s.repo.GetUserByLogin(ctx, user.Login)
	if err != nil {
		return err
	}
	token_encoder := sha256.New()
	token_encoder.Write([]byte(user.Login))
	token_encoder.Write([]byte(user.Password))
	user.Token = hex.EncodeToString(token_encoder.Sum(nil))
	passwd_hash := sha256.Sum256([]byte(user.Password))
	user.Password = hex.EncodeToString(passwd_hash[:])
	if db_user.Password != user.Password {
		return ErrNoUser
	}
	return nil
}

func (s *Service) GetUserByToken(ctx context.Context, user *models.User) error {
	usr, err := s.repo.GetUserByToken(ctx, user.Token)
	if err != nil {
		return err
	}
	user.Login = usr.Login
	return nil
}
