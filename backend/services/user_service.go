package services

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService interface {
	FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error)
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
