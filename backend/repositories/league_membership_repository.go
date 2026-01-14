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
	FindByLeagueSortedByActivity(ctx context.Context, leagueID primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]*models.LeagueMembership, error)
	Update(ctx context.Context, membership *models.LeagueMembership) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	IsActiveMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
	UpdateLastActivity(ctx context.Context, membershipID primitive.ObjectID, timestamp time.Time) error
	AddRecentCoPlayer(ctx context.Context, membershipID primitive.ObjectID, coPlayerMembershipID primitive.ObjectID, timestamp time.Time) error
	UpdateRecentCoPlayersAfterGame(ctx context.Context, membershipID primitive.ObjectID, coPlayerIDs []primitive.ObjectID, timestamp time.Time) error
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
		{
			Keys: bson.D{{"league_id", 1}, {"last_activity_at", -1}},
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

// FindByLeagueSortedByActivity returns league members sorted by last_activity_at DESC, with nulls last
func (r *LeagueMembershipRepositoryInstance) FindByLeagueSortedByActivity(ctx context.Context, leagueID primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]*models.LeagueMembership, error) {
	filter := bson.M{
		"league_id": leagueID,
		"status": bson.M{
			"$in": []models.LeagueMembershipStatus{
				models.MembershipActive,
				models.MembershipPending,
				models.MembershipVirtual,
			},
		},
	}

	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}

	// Sort by last_activity_at DESC, with null values last
	// MongoDB sorts nulls first with -1, so we use aggregation
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$addFields", Value: bson.M{
			"has_activity": bson.M{"$gt": bson.A{"$last_activity_at", nil}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "has_activity", Value: -1},
			{Key: "last_activity_at", Value: -1},
			{Key: "created_at", Value: -1},
		}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
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

// UpdateLastActivity updates the last_activity_at field
func (r *LeagueMembershipRepositoryInstance) UpdateLastActivity(ctx context.Context, membershipID primitive.ObjectID, timestamp time.Time) error {
	filter := bson.M{"_id": membershipID}
	update := bson.M{
		"$set": bson.M{
			"last_activity_at": timestamp,
			"updated_at":       time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// AddRecentCoPlayer adds a co-player to the recent_co_players list (at the end, max 10)
func (r *LeagueMembershipRepositoryInstance) AddRecentCoPlayer(ctx context.Context, membershipID primitive.ObjectID, coPlayerMembershipID primitive.ObjectID, timestamp time.Time) error {
	// First, remove if already exists
	filter := bson.M{"_id": membershipID}
	pullUpdate := bson.M{
		"$pull": bson.M{
			"recent_co_players": bson.M{"membership_id": coPlayerMembershipID},
		},
	}
	_, _ = r.collection.UpdateOne(ctx, filter, pullUpdate)

	// Then add to the end
	pushUpdate := bson.M{
		"$push": bson.M{
			"recent_co_players": bson.M{
				"$each":     []models.RecentCoPlayer{{MembershipID: coPlayerMembershipID, LastPlayedAt: timestamp}},
				"$position": 0, // Add at the beginning (most recent first)
				"$slice":    models.MaxRecentCoPlayers,
			},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, pushUpdate)
	return err
}

// UpdateRecentCoPlayersAfterGame updates recent_co_players after a game finishes
// It adds all co-players with the given timestamp, keeping max 10 sorted by LastPlayedAt DESC
func (r *LeagueMembershipRepositoryInstance) UpdateRecentCoPlayersAfterGame(ctx context.Context, membershipID primitive.ObjectID, coPlayerIDs []primitive.ObjectID, timestamp time.Time) error {
	// Get current membership
	membership, err := r.FindByID(ctx, membershipID)
	if err != nil {
		return err
	}
	if membership == nil {
		return errors.New("membership not found")
	}

	// Build new list of co-players
	existingMap := make(map[primitive.ObjectID]models.RecentCoPlayer)
	for _, cp := range membership.RecentCoPlayers {
		existingMap[cp.MembershipID] = cp
	}

	// Update/add co-players from this game
	for _, cpID := range coPlayerIDs {
		existingMap[cpID] = models.RecentCoPlayer{
			MembershipID: cpID,
			LastPlayedAt: timestamp,
		}
	}

	// Convert to slice and sort by LastPlayedAt DESC
	newList := make([]models.RecentCoPlayer, 0, len(existingMap))
	for _, cp := range existingMap {
		newList = append(newList, cp)
	}

	// Sort by LastPlayedAt DESC
	for i := 0; i < len(newList)-1; i++ {
		for j := i + 1; j < len(newList); j++ {
			if newList[j].LastPlayedAt.After(newList[i].LastPlayedAt) {
				newList[i], newList[j] = newList[j], newList[i]
			}
		}
	}

	// Trim to max size
	if len(newList) > models.MaxRecentCoPlayers {
		newList = newList[:models.MaxRecentCoPlayers]
	}

	// Update the document
	filter := bson.M{"_id": membershipID}
	update := bson.M{
		"$set": bson.M{
			"recent_co_players": newList,
			"last_activity_at":  timestamp,
			"updated_at":        time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
