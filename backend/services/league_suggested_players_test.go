package services

import (
	"context"
	"testing"
	"time"

	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUpdatePlayersAfterGame_UpdatesRecentCoPlayersForEachMember(t *testing.T) {
	ctx := context.Background()

	mockLeagueRepo := new(mocks.MockLeagueRepository)
	mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
	mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)

	service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()
	playerIDs := []primitive.ObjectID{id1, id2}

	mockMembershipRepo.
		On("UpdateRecentCoPlayersAfterGame", ctx, id1, mock.Anything, mock.Anything).
		Return(nil).Once()
	mockMembershipRepo.
		On("UpdateRecentCoPlayersAfterGame", ctx, id2, mock.Anything, mock.Anything).
		Return(nil).Once()

	err := service.UpdatePlayersAfterGame(ctx, playerIDs)

	assert.NoError(t, err)
	mockMembershipRepo.AssertExpectations(t)
}

func TestUpdatePlayersAfterGame_NoPlayersDoesNothing(t *testing.T) {
	ctx := context.Background()

	mockLeagueRepo := new(mocks.MockLeagueRepository)
	mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
	mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)

	service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

	err := service.UpdatePlayersAfterGame(ctx, []primitive.ObjectID{})

	assert.NoError(t, err)
	// No expectations to assert – should simply return without calling repo
}

func TestGetSuggestedPlayers_ForMemberWithRecentAndOtherPlayers(t *testing.T) {
	ctx := context.Background()

	mockLeagueRepo := new(mocks.MockLeagueRepository)
	mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
	mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)

	service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

	leagueID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	now := time.Now()

	currentMembership := &models.LeagueMembership{
		ID:     primitive.NewObjectID(),
		Alias:  "Current",
		Status: models.MembershipActive,
		RecentCoPlayers: []models.RecentCoPlayer{
			{MembershipID: primitive.NewObjectID(), LastPlayedAt: now.Add(-time.Hour)},
			{MembershipID: primitive.NewObjectID(), LastPlayedAt: now.Add(-2 * time.Hour)},
		},
	}

	// One of the recent co-players will be banned and should be skipped
	recentActiveMember := &models.LeagueMembership{
		ID:     currentMembership.RecentCoPlayers[0].MembershipID,
		Alias:  "Recent Active",
		Status: models.MembershipActive,
	}
	recentBannedMember := &models.LeagueMembership{
		ID:     currentMembership.RecentCoPlayers[1].MembershipID,
		Alias:  "Recent Banned",
		Status: models.MembershipBanned,
	}

	otherMember := &models.LeagueMembership{
		ID:     primitive.NewObjectID(),
		Alias:  "Other",
		Status: models.MembershipActive,
	}

	mockMembershipRepo.
		On("FindByLeagueAndUser", ctx, leagueID, userID).
		Return(currentMembership, nil)

	// Called for each RecentCoPlayer
	mockMembershipRepo.
		On("FindByID", ctx, recentActiveMember.ID).
		Return(recentActiveMember, nil)
	mockMembershipRepo.
		On("FindByID", ctx, recentBannedMember.ID).
		Return(recentBannedMember, nil)

	// other_players should exclude current and recent
	mockMembershipRepo.
		On("FindByLeagueSortedByActivity", ctx, leagueID, mock.Anything, 10).
		Return([]*models.LeagueMembership{otherMember}, nil)

	resp, err := service.GetSuggestedPlayers(ctx, leagueID, userID, false)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Current player present
	if assert.NotNil(t, resp.CurrentPlayer) {
		assert.Equal(t, currentMembership.Alias, resp.CurrentPlayer.Alias)
	}

	// Only active recent co-player should be included
	assert.Len(t, resp.RecentPlayers, 1)
	assert.Equal(t, recentActiveMember.Alias, resp.RecentPlayers[0].Alias)

	// Other players list populated from repository
	assert.Len(t, resp.OtherPlayers, 1)
	assert.Equal(t, otherMember.Alias, resp.OtherPlayers[0].Alias)

	mockMembershipRepo.AssertExpectations(t)
}

func TestGetSuggestedPlayers_ForSuperAdminWithoutMembership_UsesLargerLimit(t *testing.T) {
	ctx := context.Background()

	mockLeagueRepo := new(mocks.MockLeagueRepository)
	mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
	mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockGameRoundRepo := new(mocks.MockGameRoundRepository)

	service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

	leagueID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	// Superadmin has no membership in this league
	mockMembershipRepo.
		On("FindByLeagueAndUser", ctx, leagueID, userID).
		Return(nil, nil)

	// Expect limit 20 and no excluded IDs (nil slice)
	mockMembershipRepo.
		On("FindByLeagueSortedByActivity", ctx, leagueID, mock.Anything, 20).
		Return([]*models.LeagueMembership{}, nil)

	resp, err := service.GetSuggestedPlayers(ctx, leagueID, userID, true)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.CurrentPlayer)
	assert.Empty(t, resp.RecentPlayers)
	assert.Empty(t, resp.OtherPlayers)

	mockMembershipRepo.AssertExpectations(t)
}
