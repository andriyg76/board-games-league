import { apiFetch } from './apiClient'
import type {
  CreateGameRequest,
  CreateGameResponse,
  WizardGame,
  SubmitBidsRequest,
  SubmitResultsRequest,
  EditRoundRequest,
  EditRoundResponse,
  ScoreboardResponse,
  FinalizeGameResponse
} from '@/wizard/types'

export default {
  /**
   * Create a new Wizard game
   */
  async createGame(leagueCode: string, request: CreateGameRequest): Promise<CreateGameResponse> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(request)
      })
      if (!response.ok) {
        throw new Error('Error creating Wizard game')
      }
      return await response.json()
    } catch (error) {
      console.error('Error creating Wizard game:', error)
      throw error
    }
  },

  /**
   * Get game by code
   */
  async getGame(leagueCode: string, code: string): Promise<WizardGame> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}`)
      if (!response.ok) {
        throw new Error('Error fetching Wizard game')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching Wizard game:', error)
      throw error
    }
  },

  /**
   * Get game by GameRound ID
   */
  async getGameByRoundID(leagueCode: string, gameRoundId: string): Promise<WizardGame> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/by-round/${gameRoundId}`)
      if (!response.ok) {
        throw new Error('Error fetching Wizard game')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching Wizard game:', error)
      throw error
    }
  },

  /**
   * Delete game
   */
  async deleteGame(leagueCode: string, code: string): Promise<void> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}`, {
        method: 'DELETE'
      })
      if (!response.ok) {
        throw new Error('Error deleting Wizard game')
      }
    } catch (error) {
      console.error('Error deleting Wizard game:', error)
      throw error
    }
  },

  /**
   * Submit bids for a round
   */
  async submitBids(leagueCode: string, code: string, roundNumber: number, bids: number[]): Promise<void> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/rounds/${roundNumber}/bids`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ bids })
      })
      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Error submitting bids')
      }
    } catch (error) {
      console.error('Error submitting bids:', error)
      throw error
    }
  },

  /**
   * Submit results for a round
   */
  async submitResults(leagueCode: string, code: string, roundNumber: number, results: number[]): Promise<void> {
    try {
      const response = await apiFetch(
        `/api/leagues/${leagueCode}/wizard/games/${code}/rounds/${roundNumber}/results`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ results })
        }
      )
      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Error submitting results')
      }
    } catch (error) {
      console.error('Error submitting results:', error)
      throw error
    }
  },

  /**
   * Complete a round (calculate scores)
   */
  async completeRound(leagueCode: string, code: string, roundNumber: number): Promise<WizardGame> {
    try {
      const response = await apiFetch(
        `/api/leagues/${leagueCode}/wizard/games/${code}/rounds/${roundNumber}/complete`,
        {
          method: 'POST'
        }
      )
      if (!response.ok) {
        throw new Error('Error completing round')
      }
      return await response.json()
    } catch (error) {
      console.error('Error completing round:', error)
      throw error
    }
  },

  /**
   * Restart a round (clear bids/results)
   */
  async restartRound(leagueCode: string, code: string, roundNumber: number): Promise<void> {
    try {
      const response = await apiFetch(
        `/api/leagues/${leagueCode}/wizard/games/${code}/rounds/${roundNumber}/restart`,
        {
          method: 'POST'
        }
      )
      if (!response.ok) {
        throw new Error('Error restarting round')
      }
    } catch (error) {
      console.error('Error restarting round:', error)
      throw error
    }
  },

  /**
   * Edit round (fix mistakes)
   */
  async editRound(
    leagueCode: string,
    code: string,
    roundNumber: number,
    data: EditRoundRequest
  ): Promise<EditRoundResponse> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/rounds/${roundNumber}/edit`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      })
      if (!response.ok) {
        throw new Error('Error editing round')
      }
      return await response.json()
    } catch (error) {
      console.error('Error editing round:', error)
      throw error
    }
  },

  /**
   * Get full scoreboard
   */
  async getScoreboard(leagueCode: string, code: string): Promise<ScoreboardResponse> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/scoreboard`)
      if (!response.ok) {
        throw new Error('Error fetching scoreboard')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching scoreboard:', error)
      throw error
    }
  },

  /**
   * Finalize game (update GameRound scores)
   */
  async finalizeGame(leagueCode: string, code: string): Promise<FinalizeGameResponse> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/finalize`, {
        method: 'POST'
      })
      if (!response.ok) {
        throw new Error('Error finalizing game')
      }
      return await response.json()
    } catch (error) {
      console.error('Error finalizing game:', error)
      throw error
    }
  },

  /**
   * Move to next round
   */
  async nextRound(leagueCode: string, code: string): Promise<void> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/next-round`, {
        method: 'POST'
      })
      if (!response.ok) {
        throw new Error('Error moving to next round')
      }
    } catch (error) {
      console.error('Error moving to next round:', error)
      throw error
    }
  },

  /**
   * Move to previous round (view only)
   */
  async prevRound(leagueCode: string, code: string): Promise<void> {
    try {
      const response = await apiFetch(`/api/leagues/${leagueCode}/wizard/games/${code}/prev-round`, {
        method: 'POST'
      })
      if (!response.ok) {
        throw new Error('Error moving to previous round')
      }
    } catch (error) {
      console.error('Error moving to previous round:', error)
      throw error
    }
  }
}
