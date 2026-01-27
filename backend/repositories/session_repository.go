package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/andriyg76/bgl/db"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/hexerr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	FindByRotateToken(ctx context.Context, rotateToken string) (*models.Session, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Session, error)
	Update(ctx context.Context, session *models.Session) error
	Delete(ctx context.Context, rotateToken string) error
	DeleteExpired(ctx context.Context) error
}

type SessionRepositoryInstance struct {
	collection *mongo.Collection
}

func NewSessionRepository(mongodb *db.MongoDB) (SessionRepository, error) {
	repository := &SessionRepositoryInstance{
		collection: mongodb.Collection("sessions"),
	}
	if err := ensureSessionIndexes(repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func ensureSessionIndexes(r *SessionRepositoryInstance) error {
	_, err := r.collection.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.M{"rotate_token": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.M{"user_id": 1},
			Options: options.Index(),
		},
		{
			Keys:    bson.M{"expires_at": 1},
			Options: options.Index(),
		},
	})
	return err
}

func (r *SessionRepositoryInstance) Create(ctx context.Context, session *models.Session) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	session.LastRotationAt = time.Now()
	session.Version = 1

	result, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return err
	}

	session.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *SessionRepositoryInstance) FindByRotateToken(ctx context.Context, rotateToken string) (*models.Session, error) {
	var session models.Session
	filter := bson.M{"rotate_token": rotateToken}
	if err := r.collection.FindOne(ctx, filter).Decode(&session); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepositoryInstance) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Session, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	var sessions []*models.Session
	if err = cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepositoryInstance) Update(ctx context.Context, session *models.Session) error {
	session.UpdatedAt = time.Now()
	currentVersion := session.Version
	session.Version++

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":     session.ID,
			"version": currentVersion,
		},
		bson.M{"$set": session},
	)
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return hexerr.New("concurrent modification detected")
	}
	return nil
}

func (r *SessionRepositoryInstance) Delete(ctx context.Context, rotateToken string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"rotate_token": rotateToken})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return hexerr.New("session not found")
	}
	return nil
}

func (r *SessionRepositoryInstance) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	result, err := r.collection.DeleteMany(ctx, bson.M{"expires_at": bson.M{"$lt": now}})
	if err != nil {
		return err
	}
	if result.DeletedCount > 0 {
		// Log if needed
	}
	return nil
}
