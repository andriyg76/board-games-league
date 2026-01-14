package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/bgl/utils"
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
	GetLeagueMemberships(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueMemberInfo, error)
	GetMemberByID(ctx context.Context, membershipID primitive.ObjectID) (*models.LeagueMembership, error)
	IsUserMember(ctx context.Context, leagueID, userID primitive.ObjectID) (bool, error)
	BanUserFromLeague(ctx context.Context, leagueID, userID primitive.ObjectID) error

	// Запрошення
	CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID, playerAlias string) (*models.LeagueInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.League, error)
	PreviewInvitation(ctx context.Context, token string) (*InvitationPreview, error)
	GetInvitationByToken(ctx context.Context, token string) (*models.LeagueInvitation, error)
	ListMyInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error)
	ListMyExpiredInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error)
	CancelInvitation(ctx context.Context, token string, userID primitive.ObjectID) error
	ExtendInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.LeagueInvitation, error)
	UpdatePendingMemberAlias(ctx context.Context, membershipID primitive.ObjectID, userID primitive.ObjectID, newAlias string) error

	// Рейтинг
	GetLeagueStandings(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueStanding, error)

	// Підтримка вибору гравців для гри
	UpdatePlayersAfterGame(ctx context.Context, playerMembershipIDs []primitive.ObjectID) error
	GetSuggestedPlayers(ctx context.Context, leagueID primitive.ObjectID, userID primitive.ObjectID, isSuperAdmin bool) (*SuggestedPlayersResponse, error)
	GetMembershipByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error)
}

// LeagueMemberInfo represents a member with user and membership info
type LeagueMemberInfo struct {
	MembershipID primitive.ObjectID
	UserID       primitive.ObjectID
	UserName     string
	UserAlias    string
	UserAvatar   string
	Status       models.LeagueMembershipStatus
	JoinedAt     time.Time
}

// SuggestedPlayer represents a player suggestion for game creation
type SuggestedPlayer struct {
	MembershipID string    `json:"membership_id"`
	Alias        string    `json:"alias"`
	Avatar       string    `json:"avatar,omitempty"`
	LastPlayedAt string    `json:"last_played_at,omitempty"`
	IsVirtual    bool      `json:"is_virtual"`
}

// SuggestedPlayersResponse contains suggested players for game creation
type SuggestedPlayersResponse struct {
	CurrentPlayer *SuggestedPlayer   `json:"current_player"`
	RecentPlayers []SuggestedPlayer  `json:"recent_players"`
	OtherPlayers  []SuggestedPlayer  `json:"other_players"`
}

// InvitationPreview represents public invitation preview data
type InvitationPreview struct {
	LeagueName   string
	InviterAlias string
	PlayerAlias  string
	ExpiresAt    time.Time
	Status       string // "valid", "expired", "used"
}

// AlreadyMemberError is returned when user is already a member of the league
type AlreadyMemberError struct {
	LeagueCode string
}

func (e *AlreadyMemberError) Error() string {
	return "user is already a member of this league"
}

// IsAlreadyMemberError checks if the error is AlreadyMemberError and returns the league code
func IsAlreadyMemberError(err error) (string, bool) {
	var alreadyMemberErr *AlreadyMemberError
	if errors.As(err, &alreadyMemberErr) {
		return alreadyMemberErr.LeagueCode, true
	}
	return "", false
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

func (s *leagueServiceInstance) GetLeagueMemberships(ctx context.Context, leagueID primitive.ObjectID) ([]*LeagueMemberInfo, error) {
	memberships, err := s.membershipRepo.FindByLeague(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to get league memberships: %w", err)
	}

	members := make([]*LeagueMemberInfo, 0, len(memberships))
	for _, membership := range memberships {
		// Skip banned members
		if membership.Status == models.MembershipBanned {
			continue
		}

		member := &LeagueMemberInfo{
			MembershipID: membership.ID,
			UserID:       membership.UserID,
			UserAlias:    membership.Alias,
			Status:       membership.Status,
			JoinedAt:     membership.JoinedAt,
		}

		// If user exists, get their info
		if !membership.UserID.IsZero() {
			user, err := s.userRepo.FindByID(ctx, membership.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user %s: %w", membership.UserID.Hex(), err)
			}
			if user != nil {
				member.UserName = user.Name
				member.UserAvatar = user.Avatar
			}
		}

		// For pending members, use alias as name if no user
		if member.UserName == "" && member.UserAlias != "" {
			member.UserName = member.UserAlias
		}

		members = append(members, member)
	}

	return members, nil
}

func (s *leagueServiceInstance) GetMemberByID(ctx context.Context, membershipID primitive.ObjectID) (*models.LeagueMembership, error) {
	membership, err := s.membershipRepo.FindByID(ctx, membershipID)
	if err != nil {
		return nil, fmt.Errorf("failed to find membership: %w", err)
	}
	if membership == nil {
		return nil, errors.New("membership not found")
	}
	return membership, nil
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

func (s *leagueServiceInstance) CreateInvitation(ctx context.Context, leagueID, createdBy primitive.ObjectID, playerAlias string) (*models.LeagueInvitation, error) {
	// Verify league exists
	if _, err := s.GetLeague(ctx, leagueID); err != nil {
		return nil, err
	}

	// Validate alias
	if playerAlias == "" {
		return nil, errors.New("player alias is required")
	}

	// Check if alias already exists in this league
	existingMembership, err := s.membershipRepo.FindByLeagueAndAlias(ctx, leagueID, playerAlias)
	if err != nil {
		return nil, fmt.Errorf("failed to check alias uniqueness: %w", err)
	}

	var membership *models.LeagueMembership

	now := time.Now()

	if existingMembership != nil {
		// Alias exists - check if it's a virtual member that can be reused
		if existingMembership.Status == models.MembershipVirtual {
			// Reuse virtual membership - convert back to pending
			existingMembership.Status = models.MembershipPending
			existingMembership.LastActivityAt = now
			if err := s.membershipRepo.Update(ctx, existingMembership); err != nil {
				return nil, fmt.Errorf("failed to update virtual membership: %w", err)
			}
			membership = existingMembership
		} else {
			// Alias is taken by active or pending member
			return nil, errors.New("alias already exists in this league")
		}
	} else {
		// Create new pending membership
		membership = &models.LeagueMembership{
			LeagueID:       leagueID,
			Alias:          playerAlias,
			Status:         models.MembershipPending,
			JoinedAt:       now,
			LastActivityAt: now,
		}

		if err := s.membershipRepo.Create(ctx, membership); err != nil {
			return nil, fmt.Errorf("failed to create pending membership: %w", err)
		}
	}

	// Generate cryptographically secure token
	token, err := generateInvitationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	invitation := &models.LeagueInvitation{
		LeagueID:     leagueID,
		CreatedBy:    createdBy,
		Token:        token,
		PlayerAlias:  playerAlias,
		MembershipID: membership.ID,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Update membership with invitation ID
	membership.InvitationID = invitation.ID
	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to link membership to invitation: %w", err)
	}

	// Add the new member to creator's recent_co_players cache
	creatorMembership, _ := s.membershipRepo.FindByLeagueAndUser(ctx, leagueID, createdBy)
	if creatorMembership != nil {
		// Add to the end of the cache (will push out oldest if at max)
		_ = s.membershipRepo.AddRecentCoPlayer(ctx, creatorMembership.ID, membership.ID, now)
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

	// Check self-use: creator cannot use their own invitation
	if invitation.CreatedBy == userID {
		return nil, errors.New("you cannot accept your own invitation")
	}

	// Check if user is already an active member
	existing, err := s.membershipRepo.FindByLeagueAndUser(ctx, invitation.LeagueID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil && existing.Status == models.MembershipActive {
		// Get league code for the error
		league, _ := s.leagueRepo.FindByID(ctx, invitation.LeagueID)
		leagueCode := ""
		if league != nil {
			leagueCode = utils.IdToCode(league.ID)
		}
		return nil, &AlreadyMemberError{LeagueCode: leagueCode}
	}

	// Get the pending membership created with the invitation
	membership, err := s.membershipRepo.FindByID(ctx, invitation.MembershipID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending membership: %w", err)
	}
	if membership == nil {
		return nil, errors.New("pending membership not found")
	}

	// Update pending membership to active
	membership.UserID = userID
	membership.Status = models.MembershipActive
	membership.JoinedAt = time.Now()

	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to activate membership: %w", err)
	}

	// Mark invitation as used
	if err := s.invitationRepo.MarkAsUsed(ctx, invitation.ID, userID); err != nil {
		return nil, fmt.Errorf("failed to mark invitation as used: %w", err)
	}

	// Get and return the league
	return s.GetLeague(ctx, invitation.LeagueID)
}

func (s *leagueServiceInstance) PreviewInvitation(ctx context.Context, token string) (*InvitationPreview, error) {
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// Get league name
	league, err := s.leagueRepo.FindByID(ctx, invitation.LeagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to find league: %w", err)
	}
	if league == nil {
		return nil, errors.New("league not found")
	}

	// Get inviter info (membership alias or user name)
	inviterAlias := ""
	inviterMembership, _ := s.membershipRepo.FindByLeagueAndUser(ctx, invitation.LeagueID, invitation.CreatedBy)
	if inviterMembership != nil && inviterMembership.Alias != "" {
		inviterAlias = inviterMembership.Alias
	} else {
		inviterUser, _ := s.userRepo.FindByID(ctx, invitation.CreatedBy)
		if inviterUser != nil {
			inviterAlias = inviterUser.Name
		}
	}

	// Determine status
	status := "valid"
	if invitation.IsUsed {
		status = "used"
	} else if time.Now().After(invitation.ExpiresAt) {
		status = "expired"
	}

	return &InvitationPreview{
		LeagueName:   league.Name,
		InviterAlias: inviterAlias,
		PlayerAlias:  invitation.PlayerAlias,
		ExpiresAt:    invitation.ExpiresAt,
		Status:       status,
	}, nil
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

	// Handle the pending membership
	if !invitation.MembershipID.IsZero() {
		membership, err := s.membershipRepo.FindByID(ctx, invitation.MembershipID)
		if err != nil {
			return fmt.Errorf("failed to find membership: %w", err)
		}
		if membership != nil && membership.Status == models.MembershipPending {
			// Check if membership has games
			hasGames, err := s.gameRoundRepo.HasGamesForMembership(ctx, membership.ID)
			if err != nil {
				return fmt.Errorf("failed to check games: %w", err)
			}

			if hasGames {
				// Has games - convert to virtual
				membership.Status = models.MembershipVirtual
				if err := s.membershipRepo.Update(ctx, membership); err != nil {
					return fmt.Errorf("failed to update membership to virtual: %w", err)
				}
			} else {
				// No games - delete the membership
				if err := s.membershipRepo.Delete(ctx, membership.ID); err != nil {
					return fmt.Errorf("failed to delete membership: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *leagueServiceInstance) ListMyExpiredInvitations(ctx context.Context, leagueID, userID primitive.ObjectID) ([]*models.LeagueInvitation, error) {
	invitations, err := s.invitationRepo.FindExpiredByCreator(ctx, leagueID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired invitations: %w", err)
	}
	return invitations, nil
}

func (s *leagueServiceInstance) ExtendInvitation(ctx context.Context, token string, userID primitive.ObjectID) (*models.LeagueInvitation, error) {
	// Get invitation by token to verify ownership
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return nil, errors.New("invitation not found")
	}

	// Verify the user is the creator
	if invitation.CreatedBy != userID {
		return nil, errors.New("you can only extend your own invitations")
	}

	// Can only extend if not used
	if invitation.IsUsed {
		return nil, errors.New("cannot extend used invitation")
	}

	// Extend by 7 days
	if err := s.invitationRepo.Extend(ctx, invitation.ID, 7*24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to extend invitation: %w", err)
	}

	// Return updated invitation
	return s.invitationRepo.FindByToken(ctx, token)
}

func (s *leagueServiceInstance) UpdatePendingMemberAlias(ctx context.Context, membershipID primitive.ObjectID, userID primitive.ObjectID, newAlias string) error {
	membership, err := s.membershipRepo.FindByID(ctx, membershipID)
	if err != nil {
		return fmt.Errorf("failed to find membership: %w", err)
	}
	if membership == nil {
		return errors.New("membership not found")
	}

	// Only pending memberships can have their alias edited
	if membership.Status != models.MembershipPending {
		return errors.New("can only edit alias of pending members")
	}

	// Get the invitation to verify ownership
	invitation, err := s.invitationRepo.FindByID(ctx, membership.InvitationID)
	if err != nil {
		return fmt.Errorf("failed to find invitation: %w", err)
	}
	if invitation == nil {
		return errors.New("associated invitation not found")
	}

	// Verify the user is the creator of the invitation
	if invitation.CreatedBy != userID {
		return errors.New("you can only edit aliases for invitations you created")
	}

	if newAlias == "" {
		return errors.New("alias cannot be empty")
	}

	membership.Alias = newAlias
	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return fmt.Errorf("failed to update membership alias: %w", err)
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

// UpdatePlayersAfterGame updates recent_co_players and last_activity_at for all players after a game
func (s *leagueServiceInstance) UpdatePlayersAfterGame(ctx context.Context, playerMembershipIDs []primitive.ObjectID) error {
	if len(playerMembershipIDs) == 0 {
		return nil
	}

	now := time.Now()

	// For each player, update their recent_co_players with all other players from this game
	for _, membershipID := range playerMembershipIDs {
		// Get other players (all except current)
		coPlayerIDs := make([]primitive.ObjectID, 0, len(playerMembershipIDs)-1)
		for _, otherID := range playerMembershipIDs {
			if otherID != membershipID {
				coPlayerIDs = append(coPlayerIDs, otherID)
			}
		}

		// Update this player's recent co-players cache
		if err := s.membershipRepo.UpdateRecentCoPlayersAfterGame(ctx, membershipID, coPlayerIDs, now); err != nil {
			// Log error but continue with other players
			fmt.Printf("Failed to update recent co-players for membership %s: %v\n", membershipID.Hex(), err)
		}
	}

	return nil
}

// GetMembershipByLeagueAndUser returns a membership by league and user IDs
func (s *leagueServiceInstance) GetMembershipByLeagueAndUser(ctx context.Context, leagueID, userID primitive.ObjectID) (*models.LeagueMembership, error) {
	return s.membershipRepo.FindByLeagueAndUser(ctx, leagueID, userID)
}

// GetSuggestedPlayers returns suggested players for game creation
func (s *leagueServiceInstance) GetSuggestedPlayers(ctx context.Context, leagueID primitive.ObjectID, userID primitive.ObjectID, isSuperAdmin bool) (*SuggestedPlayersResponse, error) {
	response := &SuggestedPlayersResponse{
		RecentPlayers: []SuggestedPlayer{},
		OtherPlayers:  []SuggestedPlayer{},
	}

	// Get current user's membership
	var currentMembership *models.LeagueMembership
	var excludeIDs []primitive.ObjectID

	if !userID.IsZero() {
		membership, err := s.membershipRepo.FindByLeagueAndUser(ctx, leagueID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to find current user membership: %w", err)
		}
		currentMembership = membership
	}

	// If current user is a member, add them to response and exclude list
	if currentMembership != nil && currentMembership.Status == models.MembershipActive {
		response.CurrentPlayer = s.membershipToSuggestedPlayer(currentMembership, nil)
		excludeIDs = append(excludeIDs, currentMembership.ID)

		// Add recent co-players from cache
		for _, coPlayer := range currentMembership.RecentCoPlayers {
			coPlayerMembership, err := s.membershipRepo.FindByID(ctx, coPlayer.MembershipID)
			if err != nil || coPlayerMembership == nil {
				continue
			}
			// Skip banned members
			if coPlayerMembership.Status == models.MembershipBanned {
				continue
			}

			player := s.membershipToSuggestedPlayer(coPlayerMembership, &coPlayer.LastPlayedAt)
			response.RecentPlayers = append(response.RecentPlayers, *player)
			excludeIDs = append(excludeIDs, coPlayer.MembershipID)
		}
	}

	// Determine limit for other players
	otherPlayersLimit := 10
	if isSuperAdmin && currentMembership == nil {
		otherPlayersLimit = 20
	}

	// Get other players sorted by last_activity_at
	otherMemberships, err := s.membershipRepo.FindByLeagueSortedByActivity(ctx, leagueID, excludeIDs, otherPlayersLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to find other members: %w", err)
	}

	for _, membership := range otherMemberships {
		player := s.membershipToSuggestedPlayer(membership, nil)
		response.OtherPlayers = append(response.OtherPlayers, *player)
	}

	return response, nil
}

// membershipToSuggestedPlayer converts a membership to a SuggestedPlayer
func (s *leagueServiceInstance) membershipToSuggestedPlayer(membership *models.LeagueMembership, lastPlayedAt *time.Time) *SuggestedPlayer {
	player := &SuggestedPlayer{
		MembershipID: utils.IdToCode(membership.ID),
		Alias:        membership.Alias,
		IsVirtual:    membership.Status == models.MembershipVirtual || membership.Status == models.MembershipPending,
	}

	// Get user avatar if user exists
	if !membership.UserID.IsZero() {
		// We could fetch user here, but for performance we skip it
		// The caller can enrich this data if needed
	}

	if lastPlayedAt != nil && !lastPlayedAt.IsZero() {
		player.LastPlayedAt = lastPlayedAt.Format(time.RFC3339)
	}

	return player
}
