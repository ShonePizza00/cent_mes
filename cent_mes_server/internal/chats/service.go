package chats

import (
	"cent_mes_server/models"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetChats(ctx context.Context, user *models.User) ([]models.Chat, error) {
	usr, err := s.repo.GetUserByToken(ctx, user.Token)
	user.Login = usr.Login
	if err != nil {
		return nil, err
	}
	return s.repo.UserChats(ctx, user)
}
