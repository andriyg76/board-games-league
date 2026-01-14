import { apiFetch } from './apiClient';

export type ScoringType =
    | 'classic'
    | 'mafia'
    | 'custom'
    | 'cooperative'
    | 'cooperative_with_moderator'
    | 'team_vs_team';

export const ScoringTypes: Record<ScoringType, string> = {
    classic: "Classic board game scoring",
    mafia: "Team vs Team, separate moderator (Mafia)",
    custom: "No scheme - raw scoring enter",
    cooperative: "All players win or loose",
    cooperative_with_moderator: "All players win or loose, separate moderator",
    team_vs_team: "Team vs Team"
}

export type RoleType =
    | 'optional'       // 0+ гравців
    | 'optional_one'   // 0-1 гравець
    | 'exactly_one'    // рівно 1 гравець
    | 'required'       // 1+ гравців
    | 'multiple'       // 2+ гравців
    | 'moderator';     // модератор гри

export const RoleTypes: Record<RoleType, string> = {
    optional: "Optional (0+)",
    optional_one: "Optional, max one (0-1)",
    exactly_one: "Exactly one (1)",
    required: "Required (1+)",
    multiple: "Multiple required (2+)",
    moderator: "Moderator (1)"
}

export interface LocalizedNames {
    [lang: string]: string;
}

export interface Role {
    key: string;
    names: LocalizedNames;
    color: string;
    icon: string;
    role_type: RoleType;
}

export interface GameType {
    code: string;
    key: string;
    names: LocalizedNames;
    icon: string;
    scoring_type: ScoringType;
    roles: Role[];
    min_players: number;
    max_players: number;
    built_in: boolean;
    version: number;
}

// Helper function to get localized name
export function getLocalizedName(names: LocalizedNames | undefined, lang: string = 'en'): string {
    if (!names) return '';
    return names[lang] || names['en'] || Object.values(names)[0] || '';
}

// Deprecated - for backwards compatibility
export interface Label {
    name: string;
    color: string;
    icon: string;
}

export default {
    async getGameTypes(): Promise<GameType[]> {
        try {
            const response = await apiFetch(`/api/game_types`);
            if (!response.ok) {
                throw new Error('Error fetching game types');
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching game types:', error);
            throw error;
        }
    },

    async getGameType(code: string): Promise<GameType> {
        try {
            const response = await apiFetch(`/api/game_types/${code}`);
            if (!response.ok) {
                throw new Error('Error fetching game type');
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching game type:', error);
            throw error;
        }
    },

    async createGameType(gameType: Partial<GameType>): Promise<GameType> {
        try {
            const response = await apiFetch(`/api/game_types`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(gameType),
            });
            if (!response.ok) {
                throw new Error('Error creating game type');
            }
            return await response.json();
        } catch (error) {
            console.error('Error creating game type:', error);
            throw error;
        }
    },

    async updateGameType(code: string, gameType: Partial<GameType>): Promise<GameType> {
        try {
            console.debug("Storing gametype", gameType)
            const response = await apiFetch(`/api/game_types/${code}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(gameType),
            });
            if (!response.ok) {
                throw new Error('Error updating game type');
            }
            return await response.json();
        } catch (error) {
            console.error('Error updating game type:', error);
            throw error;
        }
    },

    async deleteGameType(code: string): Promise<void> {
        try {
            const response = await apiFetch(`/api/game_types/${code}`, {
                method: 'DELETE',
            });
            if (!response.ok) {
                if (response.status === 403) {
                    throw new Error('Cannot delete built-in game type');
                }
                throw new Error('Error deleting game type');
            }
        } catch (error) {
            console.error('Error deleting game type:', error);
            throw error;
        }
    },

    async listGameRounds(): Promise<GameRound[]> {
        const response = await apiFetch('/api/game_rounds');
        if (!response.ok) {
            throw new Error('Failed to load game rounds');
        }
        return await response.json();
    },

    async createGameRound(round: GameRound): Promise<GameRound> {
        const response = await apiFetch('/api/game_rounds', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(round),
        });
        if (!response.ok) {
            throw new Error('Error creating game round');
        }
        return await response.json();
    },

    async updateGameRound(round: GameRound): Promise<GameRound> {
        const response = await apiFetch(`/api/game_rounds/${round.code}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(round),
        })
        if (!response.ok) {
            throw new Error('Error updating game round');
        }
        return await response.json();
    },

    async finalizeGameRound(code: string, finalizationData: FinalizeGameRoundRequest): Promise<void> {
        const response = await apiFetch(`/api/game_rounds/${code}/finalize`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(finalizationData),
        });
        if (!response.ok) {
            throw new Error('Error finalizing game round');
        }
    },

    async getGameRound(code: string): Promise<GameRound> {
        const response = await apiFetch(`/api/game_rounds/${code}`);
        if (!response.ok) {
            throw new Error('Failed to get game round');
        }
        return await response.json();
    },

    async listPlayers(): Promise<Player[]> {
        const response = await apiFetch('/api/players');
        if (!response.ok) {
            throw new Error('Failed to load players');
        }
        return await response.json();
    },

    async getPlayer(code: string): Promise<Player> {
        const response = await apiFetch(`/api/players/${code}`);
        if (!response.ok) {
            throw new Error('Failed to get player');
        }
        return await response.json();
    },

    async getCurrentPlayer(): Promise<Player> {
        const response = await apiFetch('/api/players/i_am');
        if (!response.ok) {
            throw new Error('Failed to get current player');
        }
        return await response.json();
    },

    // League-specific game round methods
    async listLeagueGameRounds(leagueCode: string, options?: { active?: boolean; status?: GameRoundStatus }): Promise<GameRound[]> {
        let url = `/api/leagues/${leagueCode}/game_rounds`;
        const params = new URLSearchParams();
        if (options?.active) params.set('active', 'true');
        if (options?.status) params.set('status', options.status);
        if (params.toString()) url += `?${params.toString()}`;

        const response = await apiFetch(url);
        if (!response.ok) {
            throw new Error('Failed to load league game rounds');
        }
        return await response.json();
    },

    async createLeagueGameRound(leagueCode: string, round: CreateLeagueGameRoundRequest): Promise<GameRound> {
        const response = await apiFetch(`/api/leagues/${leagueCode}/game_rounds`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(round),
        });
        if (!response.ok) {
            throw new Error('Error creating game round');
        }
        return await response.json();
    },

    // Update player roles (step 3)
    async updateRoles(code: string, players: PlayerRoleUpdate[]): Promise<GameRound> {
        const response = await apiFetch(`/api/game_rounds/${code}/roles`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ players }),
        });
        if (!response.ok) {
            throw new Error('Error updating roles');
        }
        return await response.json();
    },

    // Update scores (step 4)
    async updateScores(code: string, playerScores: Record<string, number>, teamScores?: Record<string, number>): Promise<GameRound> {
        const response = await apiFetch(`/api/game_rounds/${code}/scores`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ player_scores: playerScores, team_scores: teamScores }),
        });
        if (!response.ok) {
            throw new Error('Error updating scores');
        }
        return await response.json();
    },

    // Update round status
    async updateRoundStatus(code: string, status: GameRoundStatus, version: number): Promise<void> {
        const response = await apiFetch(`/api/game_rounds/${code}/status`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status, version }),
        });
        if (!response.ok) {
            throw new Error('Error updating round status');
        }
    },
};

// Game Round Status
export type GameRoundStatus = 'players_selected' | 'in_progress' | 'scoring' | 'completed';

export interface GameRoundPlayer {
    user_id: string;
    membership_id?: string;
    score: number;
    is_moderator: boolean;
    team_name?: string;
    label_name?: string;
    position?: number;
}

export interface GameRound {
    code?: string;
    name: string;
    game_type: string;
    game_type_id?: string;
    league_id?: string;
    status?: GameRoundStatus;
    start_time: string;
    end_time?: string;
    players: GameRoundPlayer[];
    version: number;
}

export interface Player {
    code: string,
    alias: string,
    avatar: string,
}

export interface FinalizeGameRoundRequest {
    player_scores: Record<string, number>;
    team_scores?: Record<string, number>;
    cooperative_score?: number;
}

export interface CreateLeagueGameRoundRequest {
    name?: string;
    type: string;
    players: {
        membership_id: string;
        position: number;
        is_moderator?: boolean;
        team_name?: string;
    }[];
}

export interface PlayerRoleUpdate {
    membership_id: string;
    role_key?: string;
    team_name?: string;
    is_moderator?: boolean;
}
