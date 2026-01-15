import { defineStore } from 'pinia'
import WizardApi from '@/api/WizardApi'
import type {
  WizardGame,
  CreateGameRequest,
  ScoreboardResponse
} from '@/wizard/types'

interface WizardState {
  currentGame: WizardGame | null
  scoreboard: ScoreboardResponse | null
  loading: boolean
  error: string | null
  leagueCode: string | null
}

export const useWizardStore = defineStore('wizard', {
  state: (): WizardState => ({
    currentGame: null,
    scoreboard: null,
    loading: false,
    error: null,
    leagueCode: null
  }),

  getters: {
    /**
     * Get current round number
     */
    currentRound(state): number {
      return state.currentGame?.current_round || 1
    },

    /**
     * Get max rounds
     */
    maxRounds(state): number {
      return state.currentGame?.max_rounds || 12
    },

    /**
     * Check if game is in progress
     */
    isGameInProgress(state): boolean {
      return state.currentGame?.status === 'IN_PROGRESS'
    },

    /**
     * Check if game is completed
     */
    isGameCompleted(state): boolean {
      return state.currentGame?.status === 'COMPLETED'
    },

    /**
     * Get current dealer index
     */
    currentDealerIndex(state): number | undefined {
      if (!state.currentGame || !state.currentGame.rounds) return undefined
      const round = state.currentGame.rounds[state.currentGame.current_round - 1]
      return round?.dealer_index
    },

    /**
     * Get current round data
     */
    currentRoundData(state) {
      if (!state.currentGame || !state.currentGame.rounds) return null
      return state.currentGame.rounds[state.currentGame.current_round - 1]
    },

    /**
     * Check if all bids are submitted for current round
     */
    areAllBidsSubmitted(state): boolean {
      const round = state.currentGame?.rounds?.[state.currentGame.current_round - 1]
      if (!round) return false
      return round.player_results.every(pr => pr.bid >= 0)
    },

    /**
     * Check if all results are submitted for current round
     */
    areAllResultsSubmitted(state): boolean {
      const round = state.currentGame?.rounds?.[state.currentGame.current_round - 1]
      if (!round) return false
      return round.player_results.every(pr => pr.actual >= 0)
    }
  },

  actions: {
    /**
     * Create new Wizard game
     */
    async createGame(leagueCode: string, request: CreateGameRequest): Promise<void> {
      this.loading = true
      this.error = null
      this.leagueCode = leagueCode
      try {
        const response = await WizardApi.createGame(leagueCode, request)
        // Load full game data
        await this.loadGame(leagueCode, response.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to create game'
        console.error('Error creating game:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Load game by code
     */
    async loadGame(leagueCode: string, code: string): Promise<void> {
      this.loading = true
      this.error = null
      this.leagueCode = leagueCode
      try {
        this.currentGame = await WizardApi.getGame(leagueCode, code)
      } catch (error: any) {
        this.error = error.message || 'Failed to load game'
        console.error('Error loading game:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Load game by GameRound ID
     */
    async loadGameByRoundID(leagueCode: string, gameRoundId: string): Promise<void> {
      this.loading = true
      this.error = null
      this.leagueCode = leagueCode
      try {
        this.currentGame = await WizardApi.getGameByRoundID(leagueCode, gameRoundId)
      } catch (error: any) {
        this.error = error.message || 'Failed to load game'
        console.error('Error loading game:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Submit bids for current round
     */
    async submitBids(bids: number[]): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.submitBids(
          this.leagueCode,
          this.currentGame.code,
          this.currentGame.current_round,
          bids
        )
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to submit bids'
        console.error('Error submitting bids:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Submit results for current round
     */
    async submitResults(results: number[]): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.submitResults(
          this.leagueCode,
          this.currentGame.code,
          this.currentGame.current_round,
          results
        )
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to submit results'
        console.error('Error submitting results:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Complete current round
     */
    async completeRound(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        this.currentGame = await WizardApi.completeRound(
          this.leagueCode,
          this.currentGame.code,
          this.currentGame.current_round
        )
      } catch (error: any) {
        this.error = error.message || 'Failed to complete round'
        console.error('Error completing round:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Restart current round
     */
    async restartRound(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.restartRound(this.leagueCode, this.currentGame.code, this.currentGame.current_round)
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to restart round'
        console.error('Error restarting round:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Edit round (fix mistakes)
     */
    async editRound(roundNumber: number, bids?: number[], results?: number[]): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.editRound(this.leagueCode, this.currentGame.code, roundNumber, { bids, results })
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to edit round'
        console.error('Error editing round:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Load scoreboard
     */
    async loadScoreboard(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        this.scoreboard = await WizardApi.getScoreboard(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to load scoreboard'
        console.error('Error loading scoreboard:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Finalize game
     */
    async finalizeGame(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.finalizeGame(this.leagueCode, this.currentGame.code)
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to finalize game'
        console.error('Error finalizing game:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Move to next round
     */
    async nextRound(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.nextRound(this.leagueCode, this.currentGame.code)
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to move to next round'
        console.error('Error moving to next round:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Move to previous round
     */
    async prevRound(): Promise<void> {
      if (!this.currentGame || !this.leagueCode) {
        throw new Error('No active game or league code')
      }

      this.loading = true
      this.error = null
      try {
        await WizardApi.prevRound(this.leagueCode, this.currentGame.code)
        // Reload game to get updated state
        await this.loadGame(this.leagueCode, this.currentGame.code)
      } catch (error: any) {
        this.error = error.message || 'Failed to move to previous round'
        console.error('Error moving to previous round:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * Clear current game
     */
    clearGame(): void {
      this.currentGame = null
      this.scoreboard = null
      this.error = null
      this.leagueCode = null
    }
  }
})
