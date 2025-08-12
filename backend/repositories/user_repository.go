// repositories/user_repository.go
package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, ser *models.User) error
	FindByExternalId(ctx context.Context, externalIDs []string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	AliasUnique(ctx context.Context, alias string) (bool, error)
	FindByID(ctx context.Context, ID primitive.ObjectID) (*models.User, error)
	ListAll(ctx context.Context) ([]*models.User, error)
}

type UserRepositoryInstance struct {
	collection *mongo.Collection
}

func NewUserRepository(mongodb *db.MongoDB) (UserRepository, error) {
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
			Keys:    bson.M{"external_ids": 1},
			Options: options.Index(),
		},
	})
	return err
}

func (r *UserRepositoryInstance) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Version = 1

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepositoryInstance) FindByExternalId(ctx context.Context, externalIDs []string) (*models.User, error) {
	var user models.User

	for _, iD := range externalIDs {
		filter := bson.M{"external_ids": iD}
		if err := r.collection.FindOne(ctx, filter).Decode(&user); errors.Is(err, mongo.ErrNoDocuments) {
			continue
		} else if err != nil {
			return nil, err
		} else {
			return &user, nil
		}
	}
	return nil, nil
}

func (r *UserRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User

	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else {
		return &user, err
	}
}

func (r *UserRepositoryInstance) ListAll(ctx context.Context) ([]*models.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var users []*models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepositoryInstance) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	currentVersion := user.Version
	user.Version++

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     user.ID,
			"version": currentVersion,
		},
		bson.M{"$set": user},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("concurrent modification detected")
	}
	return nil
}

func (r *UserRepositoryInstance) AliasUnique(ctx context.Context, alias string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"alias": alias})
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
