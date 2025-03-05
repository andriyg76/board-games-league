// store/player.ts
import {defineStore} from 'pinia';
import {Player} from '@/api/GameApi';

interface PlayerState {
    players: Player[];
    loading: boolean;
}

export const usePlayerStore = defineStore('player', {
    state: (): PlayerState => ({
        players: [],
        loading: false
    }),

    actions: {
        async loadPlayers() {
            this.loading = true;
            try {
                const response = await fetch('/api/players');
                if (!response.ok) {
                    throw new Error('Failed to load players');
                }
                this.players = await response.json();
            } catch (error) {
                console.error('Error loading players:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },
    },

    getters: {
        getPlayerByCode: (state) => (code: string) => {
            return state.players.find(p => p.code === code);
        },
        getAllPlayers: (state) => state.players
    }
});