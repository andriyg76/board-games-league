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

export interface Label {
    name: string;
    color: string;
    icon: string;
}

export interface GameType {
    code: string;
    version: number;
    name: string;
    icon: string;
    labels: Label[];
    teams: Label[];
    min_players: number;
    max_players: number;
    scoring_type: ScoringType;
}

export default {
    async getGameTypes() {
        try {
            const response = await fetch(`/api/game_types`);
            if (!response.ok) {
                throw new Error('Error fetching game types');
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching game types:', error);
            throw error;
        }
    },

    async createGameType(gameType: GameType) {
        try {
            const response = await fetch(`/api/game_types`, {
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

    async updateGameType(code: string, gameType: GameType) {
        try {
            console.debug("Storing gametype", gameType)
            const response = await fetch(`/api/game_types/${code}`, {
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

    async deleteGameType(code: string) {
        try {
            const response = await fetch(`/api/game_types/${code}`, {
                method: 'DELETE',
            });
            if (!response.ok) {
                throw new Error('Error deleting game type');
            }
        } catch (error) {
            console.error('Error deleting game type:', error);
            throw error;
        }
    },
    async createGameRound(round: GameRound): Promise<GameRound> {
        return Promise.resolve(null!!)
    }
};

export interface GameRoundPlayer {
    user_id: string;
    score: number;
    is_moderator: boolean;
    team_name?: string;
}

export interface GameRound {
    code?: string;
    name: string;
    game_type: string;
    start_time: string;
    players: GameRoundPlayer[];
    version: number;
}

export interface Player {
    code: string,
    alias: string,
    avatar: string,
}