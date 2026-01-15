// League API types and methods
import { apiFetch, apiJson, apiJsonPost } from './apiClient';

export type LeagueStatus = 'active' | 'archived';
export type LeagueMembershipStatus = 'active' | 'banned' | 'pending' | 'virtual';

export interface League {
    code: string;
    version: number;
    name: string;
    status: LeagueStatus;
    created_at: string;
    updated_at: string;
}

export interface LeagueMember {
    code: string;
    user_id: string;
    user_name: string;
    user_avatar: string;
    alias: string;
    status: LeagueMembershipStatus;
    joined_at: string;
}

export interface LeagueInvitation {
    token: string;
    league_id: string;
    player_alias: string;
    membership_id?: string;
    expires_at: string;
    created_at: string;
}

export interface LeagueStanding {
    membership_id: string;
    user_id: string;
    user_name: string;
    user_alias: string;
    user_avatar: string;
    is_pending: boolean;
    total_points: number;
    games_played: number;
    games_moderated: number;
    first_place_count: number;
    second_place_count: number;
    third_place_count: number;
    participation_points: number;
    position_points: number;
    moderation_points: number;
}

export interface CreateLeagueRequest {
    name: string;
}

export interface CreateInvitationResponse {
    invitation: LeagueInvitation;
    invitation_link: string;
}

export interface AcceptInvitationResponse {
    league: League;
    membership: LeagueMember;
}

export interface InvitationPreview {
    league_name: string;
    inviter_alias: string;
    player_alias: string;
    expires_at: string;
    status: 'valid' | 'expired' | 'used';
}

export interface AcceptInvitationError {
    error: string;
    league_code?: string;
}

// Suggested players for game creation
export interface SuggestedPlayer {
    membership_id: string;
    alias: string;
    avatar?: string;
    last_played_at?: string;
    is_virtual: boolean;
}

export interface SuggestedPlayersResponse {
    current_player: SuggestedPlayer | null;
    recent_players: SuggestedPlayer[];
    other_players: SuggestedPlayer[];
    can_create_membership?: boolean;
    requires_membership?: boolean;
}

export default {
    /**
     * Create a new league (superadmin only)
     */
    createLeague: (name: string): Promise<League> =>
        apiJsonPost('/api/leagues', { name }),

    /**
     * Get all leagues
     */
    listLeagues: (): Promise<League[]> =>
        apiJson('/api/leagues'),

    /**
     * Get league details by code
     */
    getLeague: (code: string): Promise<League> =>
        apiJson(`/api/leagues/${code}`),

    /**
     * Get league members
     */
    getLeagueMembers: (code: string): Promise<LeagueMember[]> =>
        apiJson(`/api/leagues/${code}/members`),

    /**
     * Get league standings
     */
    getLeagueStandings: (code: string): Promise<LeagueStanding[]> =>
        apiJson(`/api/leagues/${code}/standings`),

    /**
     * Create an invitation for a league (members only)
     */
    createInvitation: (leagueCode: string, alias: string): Promise<LeagueInvitation> =>
        apiJsonPost(`/api/leagues/${leagueCode}/invitations`, { alias }),

    /**
     * List my active invitations for a league
     */
    listMyInvitations: (leagueCode: string): Promise<LeagueInvitation[]> =>
        apiJson(`/api/leagues/${leagueCode}/invitations`),

    /**
     * List my expired invitations for a league
     */
    listMyExpiredInvitations: (leagueCode: string): Promise<LeagueInvitation[]> =>
        apiJson(`/api/leagues/${leagueCode}/invitations/expired`),

    /**
     * Cancel an invitation by token
     */
    async cancelInvitation(leagueCode: string, token: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations/${encodeURIComponent(token)}/cancel`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Failed to cancel invitation');
        }
    },

    /**
     * Extend an invitation by 7 days
     */
    extendInvitation: (leagueCode: string, token: string): Promise<LeagueInvitation> =>
        apiJsonPost(`/api/leagues/${leagueCode}/invitations/${encodeURIComponent(token)}/extend`, {}),

    /**
     * Update pending member alias
     */
    async updatePendingMemberAlias(leagueCode: string, memberCode: string, alias: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/members/${memberCode}/alias`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ alias }),
        });
        if (!response.ok) {
            throw new Error('Failed to update member alias');
        }
    },

    /**
     * Preview an invitation (public, no auth required)
     */
    async previewInvitation(token: string): Promise<InvitationPreview> {
        const response = await fetch(`/api/leagues/join/${encodeURIComponent(token)}/preview`);
        if (!response.ok) {
            throw new Error('Invitation not found');
        }
        return await response.json();
    },

    /**
     * Accept an invitation and join a league
     */
    async acceptInvitation(token: string): Promise<AcceptInvitationResponse> {
        const response = await apiFetch(`/api/leagues/join/${encodeURIComponent(token)}`, {
            method: 'POST',
        });
        if (!response.ok) {
            // Check for already member error (409 Conflict)
            if (response.status === 409) {
                const errorData: AcceptInvitationError = await response.json();
                const error = new Error(errorData.error) as Error & { leagueCode?: string };
                error.leagueCode = errorData.league_code;
                throw error;
            }
            const text = await response.text();
            throw new Error(text || 'Failed to accept invitation');
        }
        return await response.json();
    },

    /**
     * Ban a user from a league (superadmin only)
     */
    async banUserFromLeague(leagueCode: string, userCode: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/ban/${userCode}`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Failed to ban user from league');
        }
    },

    /**
     * Unban a user from a league (superadmin only)
     */
    async unbanUserFromLeague(leagueCode: string, userCode: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/unban/${userCode}`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Failed to unban user from league');
        }
    },

    /**
     * Archive a league (superadmin only)
     */
    async archiveLeague(code: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${code}/archive`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Failed to archive league');
        }
    },

    /**
     * Unarchive a league (superadmin only)
     */
    async unarchiveLeague(code: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${code}/unarchive`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error('Failed to unarchive league');
        }
    },

    /**
     * Get suggested players for game creation
     */
    getSuggestedPlayers: (leagueCode: string): Promise<SuggestedPlayersResponse> =>
        apiJson(`/api/leagues/${leagueCode}/suggested-players`),

    /**
     * Create membership for superadmin (superadmin only)
     */
    createMembership: (leagueCode: string, alias?: string): Promise<{ membership_id: string; alias: string; status: string; joined_at: string }> =>
        apiJsonPost(`/api/leagues/${leagueCode}/memberships`, { alias }),
};
