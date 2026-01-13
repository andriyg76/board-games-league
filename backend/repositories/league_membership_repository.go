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

type LeagueMembershipRepository interface {
	Create(ctx context.Context, membership *models.LeagueMembership) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueMembership, error)
	FindByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error)
	FindByLeagueAndAlias(ctx context.Context, leagueID primitive.ObjectID, alias string) (*models.LeagueMembership, error)
	FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.LeagueMembership, error)
	FindByUser(ctx context.Context, userID primitive.ObjectID) ([]*models.LeagueMembership, error)
	Update(ctx context.Context, membership *models.LeagueMembership) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	IsActiveMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
}

type LeagueMembershipRepositoryInstance struct {
	collection *mongo.Collection
}

func NewLeagueMembershipRepository(mongodb *db.MongoDB) (LeagueMembershipRepository, error) {
	repository := &LeagueMembershipRepositoryInstance{
		collection: mongodb.Collection("league_memberships"),
	}
	if err := ensureLeagueMembershipIndexes(repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func ensureLeagueMembershipIndexes(r *LeagueMembershipRepositoryInstance) error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{"league_id", 1}, {"user_id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"league_id", 1}},
		},
		{
			Keys: bson.D{{"user_id", 1}},
		},
		{
			Keys: bson.D{{"league_id", 1}, {"status", 1}},
		},
	})
	return err
}

func (r *LeagueMembershipRepositoryInstance) Create(ctx context.Context, membership *models.LeagueMembership) error {
	membership.CreatedAt = time.Now()
	membership.UpdatedAt = time.Now()
	membership.Version = 1
	if membership.Status == "" {
		membership.Status = models.MembershipActive
	}
	if membership.JoinedAt.IsZero() {
		membership.JoinedAt = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, membership)
	if err != nil {
		return err
	}

	membership.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeagueMembershipRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueMembership, error) {
	var membership models.LeagueMembership
	filter := bson.M{"_id": id}

	if err := r.collection.FindOne(ctx, filter).Decode(&membership); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &membership, nil
}

func (r *LeagueMembershipRepositoryInstance) FindByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error) {
	var membership models.LeagueMembership
	filter := bson.M{
		"league_id": leagueID,
		"user_id":   userID,
	}

	if err := r.collection.FindOne(ctx, filter).Decode(&membership); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &membership, nil
}

func (r *LeagueMembershipRepositoryInstance) FindByLeague(ctx context.Context, leagueID primitive.ObjectID) ([]*models.LeagueMembership, error) {
	filter := bson.M{"league_id": leagueID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var memberships []*models.LeagueMembership
	if err := cursor.All(ctx, &memberships); err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *LeagueMembershipRepositoryInstance) FindByUser(ctx context.Context, userID primitive.ObjectID) ([]*models.LeagueMembership, error) {
	filter := bson.M{"user_id": userID}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var memberships []*models.LeagueMembership
	if err := cursor.All(ctx, &memberships); err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *LeagueMembershipRepositoryInstance) Update(ctx context.Context, membership *models.LeagueMembership) error {
	membership.UpdatedAt = time.Now()
	membership.Version++

	filter := bson.M{
		"_id":     membership.ID,
		"version": membership.Version - 1,
	}

	update := bson.M{
		"$set": membership,
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("membership not found or version mismatch (optimistic locking)")
	}

	return nil
}

func (r *LeagueMembershipRepositoryInstance) IsActiveMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{
		"league_id": leagueID,
		"user_id":   userID,
		"status":    models.MembershipActive,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *LeagueMembershipRepositoryInstance) FindByLeagueAndAlias(ctx context.Context, leagueID primitive.ObjectID, alias string) (*models.LeagueMembership, error) {
	var membership models.LeagueMembership
	filter := bson.M{
		"league_id": leagueID,
		"alias":     alias,
	}

	if err := r.collection.FindOne(ctx, filter).Decode(&membership); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &membership, nil
}

func (r *LeagueMembershipRepositoryInstance) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
