package services

import (
	"context"
	"errors"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestListMyInvitations(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully list active invitations", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		leagueID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		expectedInvitations := []*models.LeagueInvitation{
			{
				ID:        primitive.NewObjectID(),
				LeagueID:  leagueID,
				CreatedBy: userID,
				Token:     "token1",
				IsUsed:    false,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
			{
				ID:        primitive.NewObjectID(),
				LeagueID:  leagueID,
				CreatedBy: userID,
				Token:     "token2",
				IsUsed:    false,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
				CreatedAt: time.Now(),
			},
		}

		mockInvitationRepo.On("FindActiveByCreator", ctx, leagueID, userID).Return(expectedInvitations, nil)

		invitations, err := service.ListMyInvitations(ctx, leagueID, userID)

		assert.NoError(t, err)
		assert.Len(t, invitations, 2)
		assert.Equal(t, "token1", invitations[0].Token)
		assert.Equal(t, "token2", invitations[1].Token)
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Return empty list when no invitations", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		leagueID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mockInvitationRepo.On("FindActiveByCreator", ctx, leagueID, userID).Return([]*models.LeagueInvitation{}, nil)

		invitations, err := service.ListMyInvitations(ctx, leagueID, userID)

		assert.NoError(t, err)
		assert.Empty(t, invitations)
		mockInvitationRepo.AssertExpectations(t)
	})
}

func TestCancelInvitation(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully cancel own invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		invitationID := primitive.NewObjectID()
		token := "test-token-123"

		invitation := &models.LeagueInvitation{
			ID:        invitationID,
			LeagueID:  primitive.NewObjectID(),
			CreatedBy: userID,
			Token:     token,
			IsUsed:    false,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil)
		mockInvitationRepo.On("Cancel", ctx, invitationID).Return(nil)

		err := service.CancelInvitation(ctx, token, userID)

		assert.NoError(t, err)
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail to cancel someone else's invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		ownerID := primitive.NewObjectID()
		otherUserID := primitive.NewObjectID()
		token := "test-token-123"

		invitation := &models.LeagueInvitation{
			ID:        primitive.NewObjectID(),
			LeagueID:  primitive.NewObjectID(),
			CreatedBy: ownerID, // Different user
			Token:     token,
			IsUsed:    false,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil)

		err := service.CancelInvitation(ctx, token, otherUserID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "you can only cancel your own invitations")
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail when invitation not found", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		token := "nonexistent-token"

		mockInvitationRepo.On("FindByToken", ctx, token).Return(nil, nil)

		err := service.CancelInvitation(ctx, token, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invitation not found")
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail when repository returns error", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		token := "test-token-123"

		mockInvitationRepo.On("FindByToken", ctx, token).Return(nil, errors.New("database error"))

		err := service.CancelInvitation(ctx, token, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find invitation")
		mockInvitationRepo.AssertExpectations(t)
	})
}

func TestCreateInvitation(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully create invitation with alias", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		leagueID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		playerAlias := "Петро"

		league := &models.League{
			ID:     leagueID,
			Name:   "Test League",
			Status: models.LeagueActive,
		}

		mockLeagueRepo.On("FindByID", ctx, leagueID).Return(league, nil)
		mockMembershipRepo.On("Create", ctx, mock.AnythingOfType("*models.LeagueMembership")).Return(nil)
		mockInvitationRepo.On("Create", ctx, mock.AnythingOfType("*models.LeagueInvitation")).Return(nil)
		mockMembershipRepo.On("Update", ctx, mock.AnythingOfType("*models.LeagueMembership")).Return(nil)

		invitation, err := service.CreateInvitation(ctx, leagueID, userID, playerAlias)

		assert.NoError(t, err)
		assert.NotNil(t, invitation)
		assert.Equal(t, leagueID, invitation.LeagueID)
		assert.Equal(t, userID, invitation.CreatedBy)
		assert.Equal(t, playerAlias, invitation.PlayerAlias)
		assert.NotEmpty(t, invitation.Token)
		assert.False(t, invitation.IsUsed)
		mockLeagueRepo.AssertExpectations(t)
		mockMembershipRepo.AssertExpectations(t)
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail when alias is empty", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		leagueID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		league := &models.League{
			ID:     leagueID,
			Name:   "Test League",
			Status: models.LeagueActive,
		}

		mockLeagueRepo.On("FindByID", ctx, leagueID).Return(league, nil)

		invitation, err := service.CreateInvitation(ctx, leagueID, userID, "")

		assert.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "player alias is required")
		mockLeagueRepo.AssertExpectations(t)
	})

	t.Run("Fail when league not found", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		leagueID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mockLeagueRepo.On("FindByID", ctx, leagueID).Return(nil, nil)

		invitation, err := service.CreateInvitation(ctx, leagueID, userID, "Петро")

		assert.Error(t, err)
		assert.Nil(t, invitation)
		assert.Contains(t, err.Error(), "league not found")
		mockLeagueRepo.AssertExpectations(t)
	})
}

func TestAcceptInvitation(t *testing.T) {
	ctx := context.Background()

	t.Run("Fail when trying to accept own invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		token := "test-token"

		invitation := &models.LeagueInvitation{
			ID:           primitive.NewObjectID(),
			LeagueID:     primitive.NewObjectID(),
			CreatedBy:    userID, // Same user
			Token:        token,
			MembershipID: primitive.NewObjectID(),
			IsUsed:       false,
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil)

		league, err := service.AcceptInvitation(ctx, token, userID)

		assert.Error(t, err)
		assert.Nil(t, league)
		assert.Contains(t, err.Error(), "you cannot accept your own invitation")
		mockInvitationRepo.AssertExpectations(t)
	})
}

func TestExtendInvitation(t *testing.T) {
	ctx := context.Background()

	t.Run("Successfully extend own invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		invitationID := primitive.NewObjectID()
		token := "test-token"

		invitation := &models.LeagueInvitation{
			ID:        invitationID,
			LeagueID:  primitive.NewObjectID(),
			CreatedBy: userID,
			Token:     token,
			IsUsed:    false,
			ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
		}

		extendedInvitation := &models.LeagueInvitation{
			ID:        invitationID,
			LeagueID:  invitation.LeagueID,
			CreatedBy: userID,
			Token:     token,
			IsUsed:    false,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Extended
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil).Once()
		mockInvitationRepo.On("Extend", ctx, invitationID, 7*24*time.Hour).Return(nil)
		mockInvitationRepo.On("FindByToken", ctx, token).Return(extendedInvitation, nil).Once()

		result, err := service.ExtendInvitation(ctx, token, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail to extend someone else's invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		ownerID := primitive.NewObjectID()
		otherUserID := primitive.NewObjectID()
		token := "test-token"

		invitation := &models.LeagueInvitation{
			ID:        primitive.NewObjectID(),
			LeagueID:  primitive.NewObjectID(),
			CreatedBy: ownerID,
			Token:     token,
			IsUsed:    false,
			ExpiresAt: time.Now().Add(-24 * time.Hour),
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil)

		result, err := service.ExtendInvitation(ctx, token, otherUserID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "you can only extend your own invitations")
		mockInvitationRepo.AssertExpectations(t)
	})

	t.Run("Fail to extend used invitation", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		userID := primitive.NewObjectID()
		token := "test-token"

		invitation := &models.LeagueInvitation{
			ID:        primitive.NewObjectID(),
			LeagueID:  primitive.NewObjectID(),
			CreatedBy: userID,
			Token:     token,
			IsUsed:    true, // Already used
			ExpiresAt: time.Now().Add(-24 * time.Hour),
		}

		mockInvitationRepo.On("FindByToken", ctx, token).Return(invitation, nil)

		result, err := service.ExtendInvitation(ctx, token, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "cannot extend used invitation")
		mockInvitationRepo.AssertExpectations(t)
	})
}

func TestUpdatePendingMemberAlias(t *testing.T) {
	ctx := context.Background()

	t.Run("Fail to update alias of non-pending member", func(t *testing.T) {
		mockInvitationRepo := new(mocks.MockLeagueInvitationRepository)
		mockLeagueRepo := new(mocks.MockLeagueRepository)
		mockMembershipRepo := new(mocks.MockLeagueMembershipRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockGameRoundRepo := new(mocks.MockGameRoundRepository)

		service := NewLeagueService(mockLeagueRepo, mockMembershipRepo, mockInvitationRepo, mockUserRepo, mockGameRoundRepo)

		membershipID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		creatorID := primitive.NewObjectID()

		membership := &models.LeagueMembership{
			ID:       membershipID,
			LeagueID: primitive.NewObjectID(),
			UserID:   userID,
			Status:   models.MembershipActive, // Not pending
		}

		invitation := &models.LeagueInvitation{
			ID:           primitive.NewObjectID(),
			MembershipID: membershipID,
			CreatedBy:    creatorID,
		}

		mockMembershipRepo.On("FindByID", ctx, membershipID).Return(membership, nil)
		mockInvitationRepo.On("FindByID", ctx, membership.InvitationID).Return(invitation, nil)

		err := service.UpdatePendingMemberAlias(ctx, membershipID, creatorID, "New Alias")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only edit alias of pending members")
		mockMembershipRepo.AssertExpectations(t)
	})
}


