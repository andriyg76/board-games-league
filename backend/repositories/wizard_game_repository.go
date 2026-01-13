package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WizardGameRepository interface {
	Create(ctx context.Context, game *models.WizardGame) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.WizardGame, error)
	FindByCode(ctx context.Context, code string) (*models.WizardGame, error)
	FindByGameRoundID(ctx context.Context, gameRoundID primitive.ObjectID) (*models.WizardGame, error)
	Update(ctx context.Context, game *models.WizardGame) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByCode(ctx context.Context, code string) error
}

type wizardGameRepositoryInstance struct {
	collection *mongo.Collection
}

func NewWizardGameRepository(mongodb *db.MongoDB) (WizardGameRepository, error) {
	repo := &wizardGameRepositoryInstance{
		collection: mongodb.Collection("wizard_games"),
	}

	if err := repo.ensureIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return repo, nil
}

func (r *wizardGameRepositoryInstance) ensureIndexes() error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{"code", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"game_round_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	})
	return err
}

func (r *wizardGameRepositoryInstance) Create(ctx context.Context, game *models.WizardGame) error {
	now := time.Now()
	game.CreatedAt = now
	game.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, game)
	if err != nil {
		return fmt.Errorf("failed to create wizard game: %w", err)
	}

	game.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *wizardGameRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.WizardGame, error) {
	var game models.WizardGame
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wizard game not found")
		}
		return nil, fmt.Errorf("failed to find wizard game: %w", err)
	}
	return &game, nil
}

func (r *wizardGameRepositoryInstance) FindByCode(ctx context.Context, code string) (*models.WizardGame, error) {
	var game models.WizardGame
	err := r.collection.FindOne(ctx, bson.M{"code": code}).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wizard game not found")
		}
		return nil, fmt.Errorf("failed to find wizard game: %w", err)
	}
	return &game, nil
}

func (r *wizardGameRepositoryInstance) FindByGameRoundID(ctx context.Context, gameRoundID primitive.ObjectID) (*models.WizardGame, error) {
	var game models.WizardGame
	err := r.collection.FindOne(ctx, bson.M{"game_round_id": gameRoundID}).Decode(&game)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("wizard game not found")
		}
		return nil, fmt.Errorf("failed to find wizard game: %w", err)
	}
	return &game, nil
}

func (r *wizardGameRepositoryInstance) Update(ctx context.Context, game *models.WizardGame) error {
	game.UpdatedAt = time.Now()

	result, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"_id": game.ID},
		game,
	)
	if err != nil {
		return fmt.Errorf("failed to update wizard game: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("wizard game not found")
	}

	return nil
}

func (r *wizardGameRepositoryInstance) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete wizard game: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("wizard game not found")
	}

	return nil
}

func (r *wizardGameRepositoryInstance) DeleteByCode(ctx context.Context, code string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"code": code})
	if err != nil {
		return fmt.Errorf("failed to delete wizard game: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("wizard game not found")
	}

	return nil
}
