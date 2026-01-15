// store/league.ts
import { defineStore } from 'pinia';
import LeagueApi, {
    League,
    LeagueInvitation,
    LeagueMember,
    LeagueStanding,
} from '@/api/LeagueApi';

// LocalStorage key for storing the current league code
const CURRENT_LEAGUE_CODE_KEY = 'currentLeagueCode';

interface LeagueState {
    leagues: League[];
    currentLeague: League | null;
    currentLeagueMembers: LeagueMember[];
    currentLeagueStandings: LeagueStanding[];
    loading: boolean;
    error: string | null;
}

export const useLeagueStore = defineStore('league', {
    state: (): LeagueState => ({
        leagues: [],
        currentLeague: null,
        currentLeagueMembers: [],
        currentLeagueStandings: [],
        loading: false,
        error: null,
    }),

    actions: {
        /**
         * Load all leagues
         */
        async loadLeagues() {
            this.loading = true;
            this.error = null;
            try {
                this.leagues = await LeagueApi.listLeagues();
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to load leagues';
                console.error('Error loading leagues:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Create a new league (superadmin only)
         */
        async createLeague(name: string): Promise<League> {
            this.loading = true;
            this.error = null;
            try {
                const league = await LeagueApi.createLeague(name);
                this.leagues.push(league);
                return league;
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to create league';
                console.error('Error creating league:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Set the current active league and load its data
         */
        async setCurrentLeague(code: string) {
            this.loading = true;
            this.error = null;
            try {
                // Load league details
                this.currentLeague = await LeagueApi.getLeague(code);

                // Save to localStorage
                localStorage.setItem(CURRENT_LEAGUE_CODE_KEY, code);

                // Load members and standings in parallel
                await Promise.all([
                    this.loadCurrentLeagueMembers(),
                    this.loadCurrentLeagueStandings(),
                ]);
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to load league';
                console.error('Error loading league:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Load members of the current league
         */
        async loadCurrentLeagueMembers() {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                this.currentLeagueMembers = await LeagueApi.getLeagueMembers(this.currentLeague.code);
            } catch (error) {
                console.error('Error loading league members:', error);
                throw error;
            }
        },

        /**
         * Load standings of the current league
         */
        async loadCurrentLeagueStandings() {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                this.currentLeagueStandings = await LeagueApi.getLeagueStandings(this.currentLeague.code);
            } catch (error) {
                console.error('Error loading league standings:', error);
                throw error;
            }
        },

        /**
         * Create an invitation for the current league with a player alias
         */
        async createInvitation(alias: string): Promise<LeagueInvitation> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            this.loading = true;
            this.error = null;
            try {
                return await LeagueApi.createInvitation(this.currentLeague.code, alias);
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to create invitation';
                console.error('Error creating invitation:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * List my active invitations for the current league
         */
        async listMyInvitations(): Promise<LeagueInvitation[]> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                return await LeagueApi.listMyInvitations(this.currentLeague.code);
            } catch (error) {
                console.error('Error listing invitations:', error);
                throw error;
            }
        },

        /**
         * List my expired invitations for the current league
         */
        async listMyExpiredInvitations(): Promise<LeagueInvitation[]> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                return await LeagueApi.listMyExpiredInvitations(this.currentLeague.code);
            } catch (error) {
                console.error('Error listing expired invitations:', error);
                throw error;
            }
        },

        /**
         * Cancel an invitation by token
         */
        async cancelInvitation(token: string): Promise<void> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                await LeagueApi.cancelInvitation(this.currentLeague.code, token);
            } catch (error) {
                console.error('Error cancelling invitation:', error);
                throw error;
            }
        },

        /**
         * Extend an invitation by 7 days
         */
        async extendInvitation(token: string): Promise<LeagueInvitation> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                return await LeagueApi.extendInvitation(this.currentLeague.code, token);
            } catch (error) {
                console.error('Error extending invitation:', error);
                throw error;
            }
        },

        /**
         * Update pending member alias
         */
        async updatePendingMemberAlias(memberCode: string, alias: string): Promise<void> {
            if (!this.currentLeague) {
                throw new Error('No current league set');
            }
            try {
                await LeagueApi.updatePendingMemberAlias(this.currentLeague.code, memberCode, alias);
                // Reload members to get updated data
                await this.loadCurrentLeagueMembers();
            } catch (error) {
                console.error('Error updating member alias:', error);
                throw error;
            }
        },

        /**
         * Accept an invitation and join a league
         */
        async acceptInvitation(token: string) {
            this.loading = true;
            this.error = null;
            try {
                const result = await LeagueApi.acceptInvitation(token);

                // Add league to list if not already there
                const existingLeague = this.leagues.find(l => l.code === result.league.code);
                if (!existingLeague) {
                    this.leagues.push(result.league);
                }

                // Set as current league
                await this.setCurrentLeague(result.league.code);

                return result;
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to accept invitation';
                console.error('Error accepting invitation:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Ban a user from a league (superadmin only)
         */
        async banUser(leagueCode: string, userCode: string) {
            this.loading = true;
            this.error = null;
            try {
                await LeagueApi.banUserFromLeague(leagueCode, userCode);

                // Reload members if this is the current league
                if (this.currentLeague && this.currentLeague.code === leagueCode) {
                    await this.loadCurrentLeagueMembers();
                }
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to ban user';
                console.error('Error banning user:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Unban a user from a league (superadmin only)
         */
        async unbanUser(leagueCode: string, userCode: string) {
            this.loading = true;
            this.error = null;
            try {
                await LeagueApi.unbanUserFromLeague(leagueCode, userCode);

                // Reload members if this is the current league
                if (this.currentLeague && this.currentLeague.code === leagueCode) {
                    await this.loadCurrentLeagueMembers();
                }
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to unban user';
                console.error('Error unbanning user:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Archive a league (superadmin only)
         */
        async archiveLeague(code: string) {
            this.loading = true;
            this.error = null;
            try {
                await LeagueApi.archiveLeague(code);

                // Update league in local state
                const league = this.leagues.find(l => l.code === code);
                if (league) {
                    league.status = 'archived';
                }

                // Update current league if it's the one being archived
                if (this.currentLeague && this.currentLeague.code === code) {
                    this.currentLeague.status = 'archived';
                }
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to archive league';
                console.error('Error archiving league:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Unarchive a league (superadmin only)
         */
        async unarchiveLeague(code: string) {
            this.loading = true;
            this.error = null;
            try {
                await LeagueApi.unarchiveLeague(code);

                // Update league in local state
                const league = this.leagues.find(l => l.code === code);
                if (league) {
                    league.status = 'active';
                }

                // Update current league if it's the one being unarchived
                if (this.currentLeague && this.currentLeague.code === code) {
                    this.currentLeague.status = 'active';
                }
            } catch (error) {
                this.error = error instanceof Error ? error.message : 'Failed to unarchive league';
                console.error('Error unarchiving league:', error);
                throw error;
            } finally {
                this.loading = false;
            }
        },

        /**
         * Clear current league selection
         */
        clearCurrentLeague() {
            this.currentLeague = null;
            this.currentLeagueMembers = [];
            this.currentLeagueStandings = [];
            // Clear from localStorage
            localStorage.removeItem(CURRENT_LEAGUE_CODE_KEY);
        },

        /**
         * Get the saved league code from localStorage
         */
        getSavedLeagueCode(): string | null {
            return localStorage.getItem(CURRENT_LEAGUE_CODE_KEY);
        },
    },

    getters: {
        /**
         * Get the current league code from state or localStorage
         */
        currentLeagueCode: (state) => {
            // First check if we have currentLeague in state
            if (state.currentLeague?.code) {
                return state.currentLeague.code;
            }
            // Fallback to localStorage
            return localStorage.getItem(CURRENT_LEAGUE_CODE_KEY) || '';
        },

        /**
         * Get active leagues only
         */
        activeLeagues: (state) => {
            return state.leagues.filter(league => league.status === 'active');
        },

        /**
         * Get archived leagues only
         */
        archivedLeagues: (state) => {
            return state.leagues.filter(league => league.status === 'archived');
        },

        /**
         * Check if currently loading
         */
        isLoading: (state) => state.loading,

        /**
         * Get current error message
         */
        errorMessage: (state) => state.error,

        /**
         * Get league by code
         */
        getLeagueByCode: (state) => (code: string) => {
            return state.leagues.find(league => league.code === code);
        },

        /**
         * Check if user is a member of a league
         */
        isUserMemberOfLeague: (state) => (leagueCode: string, userCode: string) => {
            if (state.currentLeague && state.currentLeague.code === leagueCode) {
                return state.currentLeagueMembers.some(
                    member => member.user_id === userCode && member.status === 'active'
                );
            }
            return false;
        },

        /**
         * Get top N players in current league standings
         */
        getTopPlayers: (state) => (n: number = 10) => {
            return state.currentLeagueStandings
                .slice()
                .sort((a, b) => b.total_points - a.total_points)
                .slice(0, n);
        },
    },
});
