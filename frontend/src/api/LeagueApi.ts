// League API types and methods
import { apiFetch } from './apiClient';

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

export default {
    /**
     * Create a new league (superadmin only)
     */
    async createLeague(name: string): Promise<League> {
        const response = await apiFetch('/api/leagues', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name }),
        });
        if (!response.ok) {
            throw new Error('Failed to create league');
        }
        return await response.json();
    },

    /**
     * Get all leagues
     */
    async listLeagues(): Promise<League[]> {
        const response = await apiFetch('/api/leagues');
        if (!response.ok) {
            throw new Error('Failed to load leagues');
        }
        return await response.json();
    },

    /**
     * Get league details by code
     */
    async getLeague(code: string): Promise<League> {
        const response = await apiFetch(`/api/leagues/${code}`);
        if (!response.ok) {
            throw new Error('Failed to get league');
        }
        return await response.json();
    },

    /**
     * Get league members
     */
    async getLeagueMembers(code: string): Promise<LeagueMember[]> {
        const response = await apiFetch(`/api/leagues/${code}/members`);
        if (!response.ok) {
            throw new Error('Failed to get league members');
        }
        return await response.json();
    },

    /**
     * Get league standings
     */
    async getLeagueStandings(code: string): Promise<LeagueStanding[]> {
        const response = await apiFetch(`/api/leagues/${code}/standings`);
        if (!response.ok) {
            throw new Error('Failed to get league standings');
        }
        return await response.json();
    },

    /**
     * Create an invitation for a league (members only)
     */
    async createInvitation(leagueCode: string, alias: string): Promise<LeagueInvitation> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ alias }),
        });
        if (!response.ok) {
            throw new Error('Failed to create invitation');
        }
        return await response.json();
    },

    /**
     * List my active invitations for a league
     */
    async listMyInvitations(leagueCode: string): Promise<LeagueInvitation[]> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations`);
        if (!response.ok) {
            throw new Error('Failed to list invitations');
        }
        return await response.json();
    },

    /**
     * List my expired invitations for a league
     */
    async listMyExpiredInvitations(leagueCode: string): Promise<LeagueInvitation[]> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations/expired`);
        if (!response.ok) {
            throw new Error('Failed to list expired invitations');
        }
        return await response.json();
    },

    /**
     * Cancel an invitation by token
     */
    async cancelInvitation(leagueCode: string, token: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations/${encodeURIComponent(token)}/cancel`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to cancel invitation');
        }
    },

    /**
     * Extend an invitation by 7 days
     */
    async extendInvitation(leagueCode: string, token: string): Promise<LeagueInvitation> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/invitations/${encodeURIComponent(token)}/extend`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to extend invitation');
        }
        return await response.json();
    },

    /**
     * Update pending member alias
     */
    async updatePendingMemberAlias(leagueCode: string, memberCode: string, alias: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/members/${memberCode}/alias`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
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
            headers: {
                'Content-Type': 'application/json',
            },
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
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to ban user from league');
        }
    },

    /**
     * Archive a league (superadmin only)
     */
    async archiveLeague(code: string): Promise<void> {
        const response = await apiFetch(`/api/leagues/${code}/archive`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
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
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to unarchive league');
        }
    },
};
