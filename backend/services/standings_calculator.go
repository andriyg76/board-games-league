package services

import (
	"context"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sort"
)

// PointsConfig defines the points awarded for different achievements
type PointsConfig struct {
	ParticipationPoints int64
	ModerationPoints    int64
	PositionPoints      map[int]int64 // position -> points
}

// DefaultPointsConfig provides the default points configuration
var DefaultPointsConfig = PointsConfig{
	ParticipationPoints: 1,
	ModerationPoints:    2,
	PositionPoints: map[int]int64{
		1: 10,
		2: 7,
		3: 5,
		4: 3,
		5: 1,
	},
}

// LeagueStanding represents a player's standing in a league
type LeagueStanding struct {
	UserID              primitive.ObjectID
	UserName            string
	UserAvatar          string
	TotalPoints         int64
	GamesPlayed         int
	GamesModerated      int
	FirstPlaceCount     int
	SecondPlaceCount    int
	ThirdPlaceCount     int
	ParticipationPoints int64
	PositionPoints      int64
	ModerationPoints    int64
}

// getPositionPoints returns points for a given position
func (c *PointsConfig) getPositionPoints(position int) int64 {
	if points, ok := c.PositionPoints[position]; ok {
		return points
	}
	// For positions > 5: max(0, 11 - position)
	points := int64(11 - position)
	if points < 0 {
		return 0
	}
	return points
}

// CalculateStandings computes the standings for all players in a league
func CalculateStandings(
	ctx context.Context,
	rounds []*models.GameRound,
	members []*models.LeagueMembership,
	users map[primitive.ObjectID]*models.User,
	config PointsConfig,
) []*LeagueStanding {
	// Create a map of user standings
	standingsMap := make(map[primitive.ObjectID]*LeagueStanding)

	// Initialize standings for all active members
	for _, member := range members {
		if member.Status != models.MembershipActive {
			continue
		}

		user, ok := users[member.UserID]
		if !ok {
			continue
		}

		standingsMap[member.UserID] = &LeagueStanding{
			UserID:     member.UserID,
			UserName:   user.Name,
			UserAvatar: user.Avatar,
		}
	}

	// Process all completed rounds
	for _, round := range rounds {
		// Skip rounds that are not finished
		if round.EndTime.IsZero() {
			continue
		}

		// Process each player in the round
		for _, player := range round.Players {
			standing, ok := standingsMap[player.PlayerID]
			if !ok {
				// Player is not an active member, skip
				continue
			}

			// Increment games played
			standing.GamesPlayed++

			// Add participation points
			standing.ParticipationPoints += config.ParticipationPoints

			// Add position points if position is valid
			if player.Position > 0 {
				posPoints := config.getPositionPoints(player.Position)
				standing.PositionPoints += posPoints

				// Count podium positions
				switch player.Position {
				case 1:
					standing.FirstPlaceCount++
				case 2:
					standing.SecondPlaceCount++
				case 3:
					standing.ThirdPlaceCount++
				}
			}

			// Add moderation points
			if player.IsModerator {
				standing.GamesModerated++
				standing.ModerationPoints += config.ModerationPoints
			}

			// Calculate total points
			standing.TotalPoints = standing.ParticipationPoints +
				standing.PositionPoints +
				standing.ModerationPoints
		}
	}

	// Convert map to slice
	standings := make([]*LeagueStanding, 0, len(standingsMap))
	for _, standing := range standingsMap {
		standings = append(standings, standing)
	}

	// Sort by total points (descending), then by games played (ascending)
	sort.Slice(standings, func(i, j int) bool {
		if standings[i].TotalPoints == standings[j].TotalPoints {
			return standings[i].GamesPlayed < standings[j].GamesPlayed
		}
		return standings[i].TotalPoints > standings[j].TotalPoints
	})

	return standings
}
