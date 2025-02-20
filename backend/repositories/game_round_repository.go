package repositories

import (
	"context"
	"fmt"
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

type gameRoundRepositoryInstance struct {
	collection *mongo.Collection
}

func NewGameRoundRepository(mongodb *db.MongoDB) (*gameRoundRepositoryInstance, error) {
	return &gameRoundRepositoryInstance{
		collection: mongodb.Collection("game_rounds"),
	}, nil
}

func (r *gameRoundRepositoryInstance) Create(ctx context.Context, round *models.GameRound) error {
	now := time.Now()
	round.CreatedAt = now
	round.UpdatedAt = now
	round.Version = 1

	result, err := r.collection.InsertOne(ctx, round)
	if err != nil {
		return err
	}
	round.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *gameRoundRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameRound, error) {
	var round models.GameRound
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&round)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &round, err
}

func (r *gameRoundRepositoryInstance) Update(ctx context.Context, round *models.GameRound) error {
	round.UpdatedAt = time.Now()
	currentVersion := round.Version
	round.Version++

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     round.ID,
			"version": currentVersion,
		},
		bson.M{"$set": round},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("concurrent modification detected")
	}
	return nil
}

func (r *gameRoundRepositoryInstance) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
