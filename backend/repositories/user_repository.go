// repositories/user_repository.go
package repositories

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Collection) *UserRepository {
	repository := &UserRepository{
		collection: db,
	}
	repository.ensureIndexes()
	return repository
}

func (r *UserRepository) ensureIndexes() {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"alias": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := r.collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		panic(err)
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *UserRepository) AliasExists(ctx context.Context, alias string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"alias": alias})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
