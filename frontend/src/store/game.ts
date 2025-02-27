import {defineStore} from 'pinia';
import {GameRound, GameType} from '@/api/GameApi';

interface GameState {
    gameTypes: GameType[];
    activeRounds: GameRound[];
}

export const useGameStore = defineStore('game', {
    state: (): GameState => ({
        gameTypes: [],
        activeRounds: []
    }),

    actions: {
        setGameTypes(types: GameType[]) {
            this.gameTypes = types;
        },

        async addActiveRound(round: GameRound) {
            this.activeRounds.push(round);
        },

        async updateRound(round: GameRound) {
            if (!round.code) {
                throw new Error("round code is not set")
            }

            for (let i = 0; i < this.activeRounds.length; i++) {
                if (this.activeRounds[i].code === round.code) {
                    this.activeRounds[i] = round
                    return
                }
            }

            throw new Error("round not found in active rounds")
        },

        removeActiveRound(code: string) {
            const index = this.activeRounds.findIndex(r => r.code === code);
            if (index > -1) {
                this.activeRounds.splice(index, 1);
            }
        }
    },

    getters: {
        getGameTypeByCode: (state) => (code: string) => {
            return state.gameTypes.find(gt => gt.code === code);
        },
        getActiveRoundByCode: (state) => (code: string) => {
            return state.activeRounds.find(r => r.code === code);
        }
    }
});