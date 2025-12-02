package messages

import (
	"cent_mes_server/models"
	"context"
	"errors"
)

type Service struct {
	repo Repository
}

var (
	ErrForbidden     = errors.New("Forbidden")
	ErrInvalidSender = errors.New("Invalid token, relogin")
	ErrNoGetter      = errors.New("No recipient")
)

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUserByToken(ctx context.Context, user *models.User) error {
	usr, err := s.repo.GetUserByToken(ctx, user.Token)
	if err != nil {
		return err
	}
	user.Login = usr.Login
	return nil
}

func (s *Service) GetMessages(ctx context.Context, cf *models.ChatFetch, usr *models.User) ([]models.Message, error) {
	if !s.repo.CanUserAccessChat(ctx, usr, cf.ChatID) {
		return nil, ErrForbidden
	}
	return s.repo.MessagesInChat(ctx, cf)
}

func (s *Service) SendMessage(ctx context.Context, msg *models.MessageCreateRequest, token string) error {
	usr, err := s.repo.GetUserByToken(ctx, token)
	if err != nil {
		return ErrInvalidSender
	}
	msg.Sender = usr.Login
	if msg.Getter != "" {
		chat_, err := s.repo.FindOrCreateNewChat(ctx, &models.User{Login: msg.Sender}, &models.User{Login: msg.Getter})
		if err != nil {
			return err
		}
		msg.ChatID = chat_.ID
		s.repo.SendMessage(ctx, msg)
	} else if msg.ChatID != 0 {
		if !s.repo.CanUserAccessChat(ctx, &models.User{Token: token}, msg.ChatID) {
			return ErrForbidden
		}
		s.repo.SendMessage(ctx, msg)
	} else {
		return ErrNoGetter
	}
	return nil
}
