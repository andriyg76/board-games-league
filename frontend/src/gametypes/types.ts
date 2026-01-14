export type GameRoundStatus = 'players_selected' | 'in_progress' | 'scoring' | 'completed';

export interface GameRoundView {
    code: string;
    name: string;
    game_type: string;
    status?: GameRoundStatus;
    start_time: string;
    end_time?: string;
    players: GameRoundPlayerView[];
    team_scores?: TeamScoreView[];
    cooperative_score?: number;
}

export interface GameRoundPlayerView {
    user_id: string;
    position: number;
    score?: number;
    is_moderator: boolean;
    team_name?: string;
}

export interface TeamScoreView {
    name: string;
    score: number;
    position: number;
}