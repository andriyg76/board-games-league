export enum BidRestriction {
  NO_RESTRICTIONS = 'NO_RESTRICTIONS',
  CANNOT_MATCH_CARDS = 'CANNOT_MATCH_CARDS',
  MUST_MATCH_CARDS = 'MUST_MATCH_CARDS'
}

export enum GameVariant {
  STANDARD = 'STANDARD',
  ANNIVERSARY = 'ANNIVERSARY'
}

export enum GameStatus {
  SETUP = 'SETUP',
  IN_PROGRESS = 'IN_PROGRESS',
  COMPLETED = 'COMPLETED'
}

export enum RoundStatus {
  BIDDING = 'BIDDING',
  PLAYING = 'PLAYING',
  COMPLETED = 'COMPLETED'
}

export interface WizardGameConfig {
  bid_restriction: BidRestriction
  game_variant: GameVariant
  first_dealer_index: number
}

export interface WizardPlayer {
  membership_id: string
  player_name: string
  total_score: number
}

export interface WizardPlayerResult {
  bid: number
  actual: number
  score: number
  delta: number
  total_score: number
}

export interface WizardRound {
  round_number: number
  dealer_index: number
  cards_count: number
  player_results: WizardPlayerResult[]
  status: RoundStatus
  completed_at?: string
}

export interface WizardGame {
  code: string
  game_round_id: string
  config: WizardGameConfig
  players: WizardPlayer[]
  rounds: WizardRound[]
  current_round: number
  max_rounds: number
  status: GameStatus
  created_at: string
  updated_at: string
}

// API Request/Response types
export interface CreateGameRequest {
  league_id: string
  game_name: string
  bid_restriction: BidRestriction
  game_variant: GameVariant
  first_dealer_index: number
  player_membership_ids: string[]
}

export interface CreateGameResponse {
  code: string
  game_round_id: string
  current_round: number
  max_rounds: number
  status: string
  players: WizardPlayer[]
}

export interface SubmitBidsRequest {
  bids: number[]
}

export interface SubmitResultsRequest {
  results: number[]
}

export interface EditRoundRequest {
  bids?: number[]
  results?: number[]
}

export interface EditRoundResponse {
  round_number: number
  recalculated_rounds: number[]
  message: string
}

export interface ScoreboardResponse {
  game_code: string
  current_round: number
  max_rounds: number
  players: WizardPlayer[]
  rounds: WizardRound[]
}

export interface FinalStanding {
  player_name: string
  total_score: number
  position: number
}

export interface FinalizeGameResponse {
  wizard_game_code: string
  game_round_id: string
  final_standings: FinalStanding[]
}
