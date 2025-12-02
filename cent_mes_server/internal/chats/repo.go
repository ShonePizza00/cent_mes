package chats

import (
	"cent_mes_server/models"
	"context"
)

type Repository interface {
	GetUserByToken(ctx context.Context, token string) (*models.User, error)
	UserChats(ctx context.Context, user *models.User) ([]models.Chat, error)
}
