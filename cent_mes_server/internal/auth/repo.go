package auth

import (
	"cent_mes_server/models"
	"context"
)

type Repository interface {
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUserByToken(ctx context.Context, token string) (*models.User, error)
	RegisterUser(ctx context.Context, user *models.User) error
}
