package repositories

import (
	"context"
	"fmt"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type GameRoundRepository interface {
	Create(ctx context.Context, round *models.GameRound) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.GameRound, error)
	FindAll(ctx context.Context) ([]*models.GameRound, error)
	FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.GameRound, error)
	Update(ctx context.Context, round *models.GameRound) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type gameRoundRepositoryInstance struct {
	collection *mongo.Collection
}

func NewGameRoundRepository(mongodb *db.MongoDB) (GameRoundRepository, error) {
	repo := &gameRoundRepositoryInstance{
		collection: mongodb.Collection("game_rounds"),
	}

	if err := repo.ensureIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

func (r *gameRoundRepositoryInstance) ensureIndexes() error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{"_id", 1},
				{"players.user_id", 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{"_id", 1},
				{"players.order", 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{"_id", 1},
				{"team_scores.name", 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{"league_id", 1},
				{"start_time", -1},
			},
		},
	})
	return err
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

func (r *gameRoundRepositoryInstance) FindAll(ctx context.Context) ([]*models.GameRound, error) {
	opts := options.Find().SetSort(bson.D{{"start_time", -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rounds []*models.GameRound
	if err = cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}

	return rounds, nil
}

func (r *gameRoundRepositoryInstance) FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.GameRound, error) {
	filter := bson.M{"league_id": leagueID}
	opts := options.Find().SetSort(bson.D{{"start_time", -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rounds []*models.GameRound
	if err = cursor.All(ctx, &rounds); err != nil {
		return nil, err
	}

	return rounds, nil
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
