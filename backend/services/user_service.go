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
	userCache      UserCache
}

func NewUserService(userRepository repositories.UserRepository, userCache UserCache) UserService {
	return &userService{
		userRepository: userRepository,
		userCache:      userCache,
	}
}

func (s *userService) FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error) {
	// Try cache first
	if user, ok := s.userCache.GetByID(ID); ok {
		return user, nil
	}

	// Cache miss - fetch from repository
	user, err := s.userRepository.FindByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if user != nil {
		s.userCache.Set(user)
	}

	return user, nil
}

func (s *userService) FindByCode(ctx context.Context, code string) (*models.User, error) {
	// Try cache first
	if user, ok := s.userCache.GetByCode(code); ok {
		return user, nil
	}

	// Cache miss - convert code to ID and fetch
	id, err := utils.CodeToID(code)
	if err != nil {
		return nil, err
	}

	return s.FindByID(ctx, id)
}

func (s *userService) FindAll(ctx context.Context) ([]*models.User, error) {
	return s.userRepository.ListAll(ctx)
}
