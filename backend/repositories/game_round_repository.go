package repositories

import (
	"context"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type GameRoundRepository interface {
	Create(ctx context.Context, round *models.GameRound) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameRound, error)
	Update(ctx context.Context, round *models.GameRound) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type GameRoundRepositoryInstance struct {
	collection *mongo.Collection
}

func NewGameRoundRepository(mongodb *db.MongoDB) (*GameRoundRepositoryInstance, error) {
	return &GameRoundRepositoryInstance{
		collection: mongodb.Collection("game_rounds"),
	}, nil
}

func (r *GameRoundRepositoryInstance) Create(ctx context.Context, round *models.GameRound) error {
	now := time.Now()
	round.CreatedAt = now
	round.UpdatedAt = now
	result, err := r.collection.InsertOne(ctx, round)
	if err != nil {
		return err
	}
	round.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *GameRoundRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameRound, error) {
	var round models.GameRound
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&round)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &round, err
}

func (r *GameRoundRepositoryInstance) Update(ctx context.Context, round *models.GameRound) error {
	round.UpdatedAt = time.Now()
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": round.ID}, round)
	return err
}

func (r *GameRoundRepositoryInstance) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
