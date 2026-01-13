package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type LeagueService interface {
	// Створення ліги (тільки суперадмін)
	CreateLeague(ctx context.Context, name string) (*models.League, error)

	// Отримання інформації про лігу
	GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*models.League, error)
	ListLeagues(ctx context.Context) ([]*models.League, error)
	ListActiveLeagues(ctx context.Context) ([]*models.League, error)

	// Управління лігою (тільки суперадмін)
	ArchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error
	UnarchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error

	// Управління членством
	GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*models.User, error)
	IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
	BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error

	// Запрошення
	CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID) (*models.LeagueInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.League, error)
	GetInvitationByToken(ctx context.Context, token string) (*models.LeagueInvitation, error)
	ListMyInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error)
	CancelInvitation(ctx context.Context, token string, userID primitive.ObjectID) error

	// Рейтинг
	GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueStanding, error)
}

type leagueServiceInstance struct {
	leagueRepo     repositories.LeagueRepository
	membershipRepo repositories.LeagueMembershipRepository
	invitationRepo repositories.LeagueInvitationRepository
	userRepo       repositories.UserRepository
	gameRoundRepo  repositories.GameRoundRepository
	pointsConfig   PointsConfig
}

func NewLeagueService(
	leagueRepo repositories.LeagueRepository,
	membershipRepo repositories.LeagueMembershipRepository,
	invitationRepo repositories.LeagueInvitationRepository,
	userRepo repositories.UserRepository,
	gameRoundRepo repositories.GameRoundRepository,
) LeagueService {
	return &leagueServiceInstance{
		leagueRepo:     leagueRepo,
		membershipRepo: membershipRepo,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		gameRoundRepo:  gameRoundRepo,
		pointsConfig:   DefaultPointsConfig,
	}
}

func (s *leagueServiceInstance) CreateLeague(ctx context.Context, name string) (*models.League, error) {
	if name == "" {
		return nil, errors.New("league name is required")
	}

	league := &models.League{
		Name:   name,
		Status: models.LeagueActive,
	}

	if err := s.leagueRepo.Create(ctx, league); err != nil {
		return nil, fmt.Errorf("failed to create league: %w", err)
	}

	return league, nil
}

func (s *leagueServiceInstance) GetLeague(ctx context.Context, leagueID primitive.ObjectID) (*models.League, error) {
	league, err := s.leagueRepo.FindByID(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league: %w", err)
	}
	if league == nil {
		return nil, errors.New("league not found")
	}
	return league, nil
}

func (s *leagueServiceInstance) ListLeagues(ctx context.Context) ([]*models.League, error) {
	leagues, err := s.leagueRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list leagues: %w", err)
	}
	return leagues, nil
}

func (s *leagueServiceInstance) ListActiveLeagues(ctx context.Context) ([]*models.League, error) {
	leagues, err := s.leagueRepo.FindByStatus(ctx, models.LeagueActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list active leagues: %w", err)
	}
	return leagues, nil
}

func (s *leagueServiceInstance) ArchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	league, err := s.GetLeague(ctx, leagueID)
	if err != nil {
		return err
	}

	if league.Status == models.LeagueArchived {
		return errors.New("league is already archived")
	}

	league.Status = models.LeagueArchived
	if err := s.leagueRepo.Update(ctx, league); err != nil {
		return fmt.Errorf("failed to archive league: %w", err)
	}

	return nil
}

func (s *leagueServiceInstance) UnarchiveLeague(ctx context.Context, leagueID primitive.ObjectID) error {
	league, err := s.GetLeague(ctx, leagueID)
	if err != nil {
		return err
	}

	if league.Status == models.LeagueActive {
		return errors.New("league is already active")
	}

	league.Status = models.LeagueActive
	if err := s.leagueRepo.Update(ctx, league); err != nil {
		return fmt.Errorf("failed to unarchive league: %w", err)
	}

	return nil
}

func (s *leagueServiceInstance) GetLeagueMembers(ctx context.Context, leagueID primitive.ObjectID) ([]*models.User, error) {
	memberships, err := s.membershipRepo.FindByLeague(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league members: %w", err)
	}

	users := make([]*models.User, 0, len(memberships))
	for _, membership := range memberships {
		if membership.Status != models.MembershipActive {
			continue
		}

		user, err := s.userRepo.FindByID(ctx, membership.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %s: %w", membership.UserID.Hex(), err)
		}
		if user != nil {
			users = append(users, user)
		}
	}

	return users, nil
}

func (s *leagueServiceInstance) IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error) {
	return s.membershipRepo.IsActiveMember(ctx, leagueID, userID)
}

func (s *leagueServiceInstance) BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error {
	membership, err := s.membershipRepo.FindByLeagueAndUser(ctx, leagueID, userID)
	if err != nil {
		return fmt.Errorf("failed to find membership: %w", err)
	}
	if membership == nil {
		return errors.New("user is not a member of this league")
	}

	if membership.Status == models.MembershipBanned {
		return errors.New("user is already banned")
	}

	membership.Status = models.MembershipBanned
	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

func (s *leagueServiceInstance) CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID) (*models.LeagueInvitation, error) {
	// Verify league exists
	if _, err := s.GetLeague(ctx, leagueID); err != nil {
		return nil, err
	}

	// Generate cryptographically secure token
	token, err := generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	invitation := &models.LeagueInvitation{
		LeagueID:  leagueID,
		CreatedBy: createdBy,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return invitation, nil
}

func (s *leagueServiceInstance) AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.League, error) {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// Validate invitation
	if invitation.IsUsed {
		return nil, errors.New("invitation has already been used")
	}
	if time.Now().After(invitation.ExpiresAt) {
		return nil, errors.New("invitation has expired")
	}

	// Check if user is already a member
	existing, err := s.membershipRepo.FindByLeagueAndUser(ctx, invitation.LeagueID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil {
		return nil, errors.New("user is already a member of this league")
	}

	// Create membership
	membership := &models.LeagueMembership{
		LeagueID: invitation.LeagueID,
		UserID:   userID,
		Status:   models.MembershipActive,
		JoinedAt: time.Now(),
	}

	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Mark invitation as used
	if err := s.invitationRepo.MarkAsUsed(ctx, invitation.ID, userID); err != nil {
		return nil, fmt.Errorf("failed to mark invitation as used: %w", err)
	}

	// Get and return the league
	return s.GetLeague(ctx, invitation.LeagueID)
}

func (s *leagueServiceInstance) GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueStanding, error) {
	// Get all game rounds for this league
	rounds, err := s.gameRoundRepo.FindByLeague(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game rounds: %w", err)
	}

	// Get all memberships
	memberships, err := s.membershipRepo.FindByLeague(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memberships: %w", err)
	}

	// Get all users
	usersMap := make(map[primitive.ObjectID]*models.User)
	for _, membership := range memberships {
		user, err := s.userRepo.FindByID(ctx, membership.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %s: %w", membership.UserID.Hex(), err)
		}
		if user != nil {
			usersMap[user.ID] = user
		}
	}

	// Calculate standings
	standings := CalculateStandings(ctx, rounds, memberships, usersMap, s.pointsConfig)

	return standings, nil
}

func (s *leagueServiceInstance) GetInvitationByToken(ctx context.Context, token string) (*models.LeagueInvitation, error) {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}
	return invitation, nil
}

func (s *leagueServiceInstance) ListMyInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	invitations, err := s.invitationRepo.FindActiveByCreator(ctx, leagueID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	return invitations, nil
}

func (s *leagueServiceInstance) CancelInvitation(ctx context.Context, token string, userID primitive.ObjectID) error {
	// Get invitation by token to verify ownership
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return errors.New("invitation not found")
	}

	// Verify the user is the creator
	if invitation.CreatedBy != userID {
		return errors.New("you can only cancel your own invitations")
	}

	// Cancel the invitation
	if err := s.invitationRepo.Cancel(ctx, invitation.ID); err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	return nil
}

// generateInvitationToken generates a cryptographically secure random token
func generateInvitationToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
