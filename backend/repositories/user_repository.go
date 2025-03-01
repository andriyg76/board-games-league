// repositories/user_repository.go
package repositories

import (
	"context"
	"errors"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type UserRepository interface {
	CreateUser(ctx context.Context, ser *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	AliasUnique(ctx context.Context, alias string) (bool, error)
}

type UserRepositoryInstance struct {
	collection *mongo.Collection
}

func NewUserRepository(mongodb *db.MongoDB) (*UserRepositoryInstance, error) {
	repository := &UserRepositoryInstance{
		collection: mongodb.Collection("users"),
	}
	if err := ensureIndexes(repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func ensureIndexes(r *UserRepositoryInstance) error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.M{"alias": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.M{"email": 1},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

func (r *UserRepositoryInstance) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepositoryInstance) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	if err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else {
		return &user, err
	}
}

//func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
//	user.CreatedAt = time.Now()
//	user.UpdatedAt = time.Now()
//
//	result, err := r.collection.InsertOne(ctx, user)
//	if err != nil {
//		return err
//	}
//
//	user.ID = result.InsertedID.(primitive.ObjectID)
//	return nil
//}

func (r *UserRepositoryInstance) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UserRepositoryInstance) AliasUnique(ctx context.Context, alias string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"alias": alias})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
