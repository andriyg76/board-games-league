package repositories

import (
	"context"
	"errors"
	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type LeagueRepository interface {
	Create(ctx context.Context, league *models.League) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.League, error)
	FindAll(ctx context.Context) ([]*models.League, error)
	FindByStatus(ctx context.Context, status models.LeagueStatus) ([]*models.League, error)
	Update(ctx context.Context, league *models.League) error
}

type LeagueRepositoryInstance struct {
	collection *mongo.Collection
}

func NewLeagueRepository(mongodb *db.MongoDB) (LeagueRepository, error) {
	repository := &LeagueRepositoryInstance{
		collection: mongodb.Collection("leagues"),
	}
	if err := ensureLeagueIndexes(repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func ensureLeagueIndexes(r *LeagueRepositoryInstance) error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.M{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"status": 1},
		},
	})
	return err
}

func (r *LeagueRepositoryInstance) Create(ctx context.Context, league *models.League) error {
	league.CreatedAt = time.Now()
	league.UpdatedAt = time.Now()
	league.Version = 1
	if league.Status == "" {
		league.Status = models.LeagueActive
	}

	result, err := r.collection.InsertOne(ctx, league)
	if err != nil {
		return err
	}

	league.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeagueRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.League, error) {
	var league models.League
	filter := bson.M{"_id": id}

	if err := r.collection.FindOne(ctx, filter).Decode(&league); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &league, nil
}

func (r *LeagueRepositoryInstance) FindAll(ctx context.Context) ([]*models.League, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var leagues []*models.League
	if err := cursor.All(ctx, &leagues); err != nil {
		return nil, err
	}

	return leagues, nil
}

func (r *LeagueRepositoryInstance) FindByStatus(ctx context.Context, status models.LeagueStatus) ([]*models.League, error) {
	filter := bson.M{"status": status}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var leagues []*models.League
	if err := cursor.All(ctx, &leagues); err != nil {
		return nil, err
	}

	return leagues, nil
}

func (r *LeagueRepositoryInstance) Update(ctx context.Context, league *models.League) error {
	league.UpdatedAt = time.Now()
	league.Version++

	filter := bson.M{
		"_id":     league.ID,
		"version": league.Version - 1,
	}

	update := bson.M{
		"$set": league,
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("league not found or version mismatch (optimistic locking)")
	}

	return nil
}
