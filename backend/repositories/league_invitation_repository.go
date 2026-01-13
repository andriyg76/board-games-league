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

type LeagueInvitationRepository interface {
	Create(ctx context.Context, invitation *models.LeagueInvitation) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueInvitation, error)
	FindByToken(ctx context.Context, token string) (*models.LeagueInvitation, error)
	FindActiveByCreator(ctx context.Context, leagueID, createdBy primitive.ObjectID) ([]*models.LeagueInvitation, error)
	MarkAsUsed(ctx context.Context, id primitive.ObjectID, usedBy primitive.ObjectID) error
	Cancel(ctx context.Context, id primitive.ObjectID) error
}

type LeagueInvitationRepositoryInstance struct {
	collection *mongo.Collection
}

func NewLeagueInvitationRepository(mongodb *db.MongoDB) (LeagueInvitationRepository, error) {
	repository := &LeagueInvitationRepositoryInstance{
		collection: mongodb.Collection("league_invitations"),
	}
	if err := ensureLeagueInvitationIndexes(repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func ensureLeagueInvitationIndexes(r *LeagueInvitationRepositoryInstance) error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.M{"token": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"league_id": 1},
		},
		{
			Keys: bson.M{"is_used": 1},
		},
		{
			Keys:    bson.M{"expires_at": 1},
			Options: options.Index().SetExpireAfterSeconds(0), // TTL index
		},
	})
	return err
}

func (r *LeagueInvitationRepositoryInstance) Create(ctx context.Context, invitation *models.LeagueInvitation) error {
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()
	invitation.Version = 1
	invitation.IsUsed = false

	// Set expiration to 7 days if not set
	if invitation.ExpiresAt.IsZero() {
		invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	result, err := r.collection.InsertOne(ctx, invitation)
	if err != nil {
		return err
	}

	invitation.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *LeagueInvitationRepositoryInstance) FindByID(ctx context.Context, id primitive.ObjectID) (*models.LeagueInvitation, error) {
	var invitation models.LeagueInvitation
	filter := bson.M{"_id": id}

	if err := r.collection.FindOne(ctx, filter).Decode(&invitation); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &invitation, nil
}

func (r *LeagueInvitationRepositoryInstance) FindByToken(ctx context.Context, token string) (*models.LeagueInvitation, error) {
	var invitation models.LeagueInvitation
	filter := bson.M{"token": token}

	if err := r.collection.FindOne(ctx, filter).Decode(&invitation); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &invitation, nil
}

func (r *LeagueInvitationRepositoryInstance) FindActiveByCreator(ctx context.Context, leagueID, createdBy primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	filter := bson.M{
		"league_id":  leagueID,
		"created_by": createdBy,
		"is_used":    false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []*models.LeagueInvitation
	if err := cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

func (r *LeagueInvitationRepositoryInstance) MarkAsUsed(ctx context.Context, id primitive.ObjectID, usedBy primitive.ObjectID) error {
	filter := bson.M{
		"_id":     id,
		"is_used": false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_used":    true,
			"used_by":    usedBy,
			"used_at":    time.Now(),
			"updated_at": time.Now(),
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("invitation not found or already used")
	}

	return nil
}

func (r *LeagueInvitationRepositoryInstance) Cancel(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{
		"_id":        id,
		"is_used":    false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	// Cancel by setting expires_at to now
	update := bson.M{
		"$set": bson.M{
			"expires_at": time.Now(),
			"updated_at": time.Now(),
		},
		"$inc": bson.M{
			"version": 1,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("invitation not found or already expired/used")
	}

	return nil
}
