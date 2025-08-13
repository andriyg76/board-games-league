package services

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error)
	FindByCode(ctx context.Context, code string) (*models.User, error)
	FindAll(ctx context.Context) ([]*models.User, error)
}

type userService struct {
	userRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error) {
	return s.userRepository.FindByID(ctx, ID)
}

func (s *userService) FindByCode(ctx context.Context, code string) (*models.User, error) {
	id, err := utils.CodeToID(code)
	if err != nil {
		return nil, err
	}
	return s.FindByID(ctx, id)
}

func (s *userService) FindAll(ctx context.Context) ([]*models.User, error) {
	return s.userRepository.ListAll(ctx)
}
