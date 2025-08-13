// store/player.ts
import {defineStore} from 'pinia';
import {Player} from '@/api/GameApi';
import GameApi from '@/api/GameApi';

interface PlayerState {
    players: Player[];
    currentPlayer: Player | null;
    loading: boolean;
}

export const usePlayerStore = defineStore('player', {
    state: (): PlayerState => ({
        players: [],
        currentPlayer: null,
        loading: false
    }),

    getters: {
        async allPlayers(state): Promise<Player[]> {
            if (state.players.length === 0) {
                state.loading = true;
                try {
                    state.players = await GameApi.listPlayers();
                } finally {
                    state.loading = false;
                }
            }
            return state.players;
        },

        async currentPlayer(state): Promise<Player | null> {
            if (!state.currentPlayer) {
                state.loading = true;
                try {
                    state.currentPlayer = await GameApi.getCurrentPlayer();
                } finally {
                    state.loading = false;
                }
            }
            return state.currentPlayer;
        },

        async getPlayerByCode(state) {
            return async (code: string): Promise<Player> => {
                const player = state.players.find(p => p.code === code);
                if (player) return player;

                state.loading = true;
                try {
                    const player = await GameApi.getPlayer(code);
                    if (player) {
                        state.players.push(player);
                    }
                    return player;
                } finally {
                    state.loading = false;
                }
            };
        },

        isLoading: (state) => state.loading
    }
});

