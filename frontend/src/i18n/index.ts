import { createI18n } from 'vue-i18n'

const messages = {
    en: {
        gameTypes: {
            title: 'Game Types',
            edit: 'Edit',
            create: 'Create',
            delete: 'Delete',
            name: 'Game Type Name',
            scoringType: 'Scoring Type',
            labels: 'Labels',
            teams: 'Teams',
            addLabel: 'Add Label',
            addTeam: 'Add Team',
            labelName: 'Label Name',
            teamName: 'Team Name',
            icon: 'Icon',
            update: 'Update',
            cancel: 'Cancel'
        },
        scoring: {
            classic: 'Classic board game scoring',
            mafia: 'Team vs Team, separate moderator (Mafia)',
            custom: 'No scheme - raw scoring enter',
            cooperative: 'All players win or loose',
            cooperative_with_moderator: 'All players win or loose, separate moderator',
            team_vs_team: 'Team vs Team'
        },
        auth: {
            login: 'Login',
            logout: 'Logout',
            loggingOut: 'Logging out...'
        }
    },
    uk: {
        gameTypes: {
            title: 'Типи Ігор',
            edit: 'Редагувати',
            create: 'Створити',
            delete: 'Видалити',
            name: 'Назва типу гри',
            scoringType: 'Тип підрахунку очок',
            labels: 'Мітки',
            teams: 'Команди',
            addLabel: 'Додати мітку',
            addTeam: 'Додати команду',
            labelName: 'Назва мітки',
            teamName: 'Назва команди',
            icon: 'Іконка',
            update: 'Оновити',
            cancel: 'Скасувати'
        },
        scoring: {
            classic: 'Класичний підрахунок очок настільної гри',
            mafia: 'Команда проти команди, окремий модератор (Мафія)',
            custom: 'Без схеми - довільне введення очок',
            cooperative: 'Всі гравці виграють або програють',
            cooperative_with_moderator: 'Всі гравці виграють або програють, окремий модератор',
            team_vs_team: 'Команда проти команди'
        },
        auth: {
            login: 'Увійти',
            logout: 'Вийти',
            loggingOut: 'Вихід...'
        }
    },
    et: {
        gameTypes: {
            title: 'Mängutüübid',
            edit: 'Muuda',
            create: 'Loo',
            delete: 'Kustuta',
            name: 'Mängutüübi nimi',
            scoringType: 'Punktiarvestuse tüüp',
            labels: 'Sildid',
            teams: 'Meeskonnad',
            addLabel: 'Lisa silt',
            addTeam: 'Lisa meeskond',
            labelName: 'Sildi nimi',
            teamName: 'Meeskonna nimi',
            icon: 'Ikoon',
            update: 'Uuenda',
            cancel: 'Tühista'
        },
        scoring: {
            classic: 'Klassikaline lauamängu punktiarvestus',
            mafia: 'Meeskond vs meeskond, eraldi moderaator (Maffia)',
            custom: 'Skeemita - vaba punktisisestus',
            cooperative: 'Kõik mängijad võidavad või kaotavad',
            cooperative_with_moderator: 'Kõik mängijad võidavad või kaotavad, eraldi moderaator',
            team_vs_team: 'Meeskond vs meeskond'
        },
        auth: {
            login: 'Logi sisse',
            logout: 'Logi välja',
            loggingOut: 'Väljalogimine...'
        }
    }
}

export const i18n = createI18n({
    legacy: false,
    locale: 'en',
    fallbackLocale: 'en',
    messages
})