package messages

import (
	"cent_mes_server/models"
	"context"
)

type Repository interface {
	GetUserByToken(ctx context.Context, token string) (*models.User, error)
	MessagesInChat(ctx context.Context, chatCF *models.ChatFetch) ([]models.Message, error)
	FindOrCreateNewChat(ctx context.Context, user1, user2 *models.User) (*models.Chat, error)
	SendMessage(ctx context.Context, msg *models.MessageCreateRequest) error
	CreateNewChat(ctx context.Context, user1, user2 *models.User) *models.Chat
	CanUserAccessChat(ctx context.Context, user *models.User, chatID int64) bool
}
