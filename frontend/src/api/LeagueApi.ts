// League API types and methods

export type LeagueStatus = 'active' | 'archived';
export type LeagueMembershipStatus = 'active' | 'banned';

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
    status: LeagueMembershipStatus;
    joined_at: string;
}

export interface LeagueInvitation {
    token: string;
    league_id: string;
    expires_at: string;
    created_at: string;
}

export interface LeagueStanding {
    user_id: string;
    user_name: string;
    user_avatar: string;
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

export default {
    /**
     * Create a new league (superadmin only)
     */
    async createLeague(name: string): Promise<League> {
        const response = await fetch('/api/leagues', {
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
        const response = await fetch('/api/leagues');
        if (!response.ok) {
            throw new Error('Failed to load leagues');
        }
        return await response.json();
    },

    /**
     * Get league details by code
     */
    async getLeague(code: string): Promise<League> {
        const response = await fetch(`/api/leagues/${code}`);
        if (!response.ok) {
            throw new Error('Failed to get league');
        }
        return await response.json();
    },

    /**
     * Get league members
     */
    async getLeagueMembers(code: string): Promise<LeagueMember[]> {
        const response = await fetch(`/api/leagues/${code}/members`);
        if (!response.ok) {
            throw new Error('Failed to get league members');
        }
        return await response.json();
    },

    /**
     * Get league standings
     */
    async getLeagueStandings(code: string): Promise<LeagueStanding[]> {
        const response = await fetch(`/api/leagues/${code}/standings`);
        if (!response.ok) {
            throw new Error('Failed to get league standings');
        }
        return await response.json();
    },

    /**
     * Create an invitation for a league (members only)
     */
    async createInvitation(leagueCode: string): Promise<CreateInvitationResponse> {
        const response = await fetch(`/api/leagues/${leagueCode}/invitations`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
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
        const response = await fetch(`/api/leagues/${leagueCode}/invitations`);
        if (!response.ok) {
            throw new Error('Failed to list invitations');
        }
        return await response.json();
    },

    /**
     * Cancel an invitation by token
     */
    async cancelInvitation(leagueCode: string, token: string): Promise<void> {
        const response = await fetch(`/api/leagues/${leagueCode}/invitations/${encodeURIComponent(token)}/cancel`, {
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
     * Accept an invitation and join a league
     */
    async acceptInvitation(token: string): Promise<AcceptInvitationResponse> {
        const response = await fetch(`/api/leagues/join/${token}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });
        if (!response.ok) {
            throw new Error('Failed to accept invitation');
        }
        return await response.json();
    },

    /**
     * Ban a user from a league (superadmin only)
     */
    async banUserFromLeague(leagueCode: string, userCode: string): Promise<void> {
        const response = await fetch(`/api/leagues/${leagueCode}/ban/${userCode}`, {
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
        const response = await fetch(`/api/leagues/${code}/archive`, {
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
        const response = await fetch(`/api/leagues/${code}/unarchive`, {
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
