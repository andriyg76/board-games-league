export type ScoringType =
    | 'classic'
    | 'mafia'
    | 'custom'
    | 'cooperative'
    | 'cooperative_with_moderator'
    | 'team_vs_team';

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
    minPlayers: number;
    maxPlayers: number;
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
    }
};