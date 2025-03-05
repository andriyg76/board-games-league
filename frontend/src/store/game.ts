// store/game.ts
import { defineStore } from 'pinia';
import { GameRound, GameType } from '@/api/GameApi';
import GameApi from '@/api/GameApi';

interface GameState {
    gameTypes: GameType[];
    activeRounds: GameRound[];
    loading: boolean;
}

export const useGameStore = defineStore('game', {
    state: (): GameState => ({
        gameTypes: [],
        activeRounds: [],
        loading: false
    }),

    actions: {
        async loadGameTypes() {
            this.loading = true;
            try {
                const types = await GameApi.getGameTypes();
                this.gameTypes = types;
            } catch (error) {
                console.error('Error loading game types:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        async addActiveRound(round: GameRound): Promise<GameRound> {
            const savedRound = await GameApi.createGameRound(round);
            this.activeRounds.push(savedRound);
            return savedRound;
        },

        // ... other existing actions
    },

    getters: {
        getGameTypeByCode: (state) => (code: string) => {
            return state.gameTypes.find(gt => gt.code === code);
        },
        isLoading: (state) => state.loading
    }
});