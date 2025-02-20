package repositories

import (
	"context"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/glog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type GameTypeRepository interface {
	Create(ctx context.Context, gameType *models.GameType) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameType, error)
	FindAll(ctx context.Context) ([]*models.GameType, error)
	Update(ctx context.Context, gameType *models.GameType) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindByName(ctx context.Context, name string) (*models.GameType, error)
}

type mongoGameTypeRepository struct {
	collection *mongo.Collection
}

func NewGameTypeRepository(db *db.MongoDB) (GameTypeRepository, error) {
	return &mongoGameTypeRepository{
		collection: db.Collection("game_types"),
	}, nil
}

func (r *mongoGameTypeRepository) FindByName(ctx context.Context, name string) (*models.GameType, error) {
	return nil, glog.Error("not implemented")
}

func (r *mongoGameTypeRepository) Create(ctx context.Context, gameType *models.GameType) error {
	gameType.CreatedAt = time.Now()
	gameType.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, gameType)
	if err != nil {
		return err
	}

	gameType.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *mongoGameTypeRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameType, error) {
	var gameType models.GameType
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&gameType)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &gameType, nil
}

func (r *mongoGameTypeRepository) FindAll(ctx context.Context) ([]*models.GameType, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var gameTypes []*models.GameType
	if err = cursor.All(ctx, &gameTypes); err != nil {
		return nil, err
	}
	return gameTypes, nil
}

func (r *mongoGameTypeRepository) Update(ctx context.Context, gameType *models.GameType) error {
	gameType.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": gameType.ID},
		gameType,
	)
	return err
}

func (r *mongoGameTypeRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
