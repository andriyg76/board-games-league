package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ScoringType визначає тип підрахунку очок
type ScoringType string

const (
	ScoringTypeClassic           ScoringType = "classic"
	ScoringTypeMafia             ScoringType = "mafia"
	ScoringTypeCustom            ScoringType = "custom"
	ScoringTypeCooperative       ScoringType = "cooperative"
	ScoringTypeCoopWithModerator ScoringType = "cooperative_with_moderator"
	ScoringTypeTeamVsTeam        ScoringType = "team_vs_team"
)

var ScoringTypes = []ScoringType{
	ScoringTypeClassic,
	ScoringTypeCooperative,
	ScoringTypeCustom,
	ScoringTypeMafia,
	ScoringTypeCoopWithModerator,
	ScoringTypeTeamVsTeam,
}

// RoleType визначає тип ролі та обмеження кількості гравців
type RoleType string

const (
	RoleTypeOptional    RoleType = "optional"      // 0+ гравців
	RoleTypeOptionalOne RoleType = "optional_one"  // 0-1 гравець
	RoleTypeExactlyOne  RoleType = "exactly_one"   // рівно 1 гравець
	RoleTypeRequired    RoleType = "required"      // 1+ гравців
	RoleTypeMultiple    RoleType = "multiple"      // 2+ гравців
	RoleTypeModerator   RoleType = "moderator"     // модератор гри (рівно 1)
)

var RoleTypes = []RoleType{
	RoleTypeOptional,
	RoleTypeOptionalOne,
	RoleTypeExactlyOne,
	RoleTypeRequired,
	RoleTypeMultiple,
	RoleTypeModerator,
}

// Role визначає роль/групу/колір гравця в грі
type Role struct {
	Key      string            `bson:"key" json:"key"`             // унікальний ключ ролі
	Names    map[string]string `bson:"names" json:"names"`         // локалізовані назви {"en": "...", "uk": "..."}
	Color    string            `bson:"color" json:"color"`         // колір ролі
	Icon     string            `bson:"icon" json:"icon"`           // іконка ролі
	RoleType RoleType          `bson:"role_type" json:"role_type"` // тип ролі
}

// Label - застаріла структура, залишена для сумісності
// Deprecated: використовуйте Role
type Label struct {
	Name  string `bson:"name"`
	Color string `bson:"color"`
	Icon  string `bson:"icon"`
}

// GameType визначає тип настільної гри
type GameType struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Key         string             `bson:"key"`                    // унікальний ключ типу гри
	Names       map[string]string  `bson:"names"`                  // локалізовані назви
	Icon        string             `bson:"icon"`                   // іконка гри
	ScoringType ScoringType        `bson:"scoring_type"`           // тип підрахунку очок
	Roles       []Role             `bson:"roles"`                  // ролі/групи гравців
	MinPlayers  int                `bson:"min_players"`            // мінімальна кількість гравців
	MaxPlayers  int                `bson:"max_players"`            // максимальна кількість гравців
	BuiltIn     bool               `bson:"built_in"`               // вбудований тип (захищений від видалення)
	Version     int64              `bson:"version"`                // версія для оптимістичного локінгу
	CreatedAt   time.Time          `bson:"created_at"`             // дата створення
	UpdatedAt   time.Time          `bson:"updated_at"`             // дата оновлення

	// Застарілі поля для сумісності
	// Deprecated: використовуйте Names
	Name string `bson:"name,omitempty"`
	// Deprecated: використовуйте Roles
	Labels []Label `bson:"labels,omitempty"`
	// Deprecated: використовуйте Roles
	Teams []Label `bson:"teams,omitempty"`
}

// GetName повертає локалізовану назву гри
func (gt *GameType) GetName(lang string) string {
	if name, ok := gt.Names[lang]; ok && name != "" {
		return name
	}
	if name, ok := gt.Names["en"]; ok && name != "" {
		return name
	}
	// Fallback на старе поле Name
	if gt.Name != "" {
		return gt.Name
	}
	return gt.Key
}

// GetRoleName повертає локалізовану назву ролі
func (r *Role) GetName(lang string) string {
	if name, ok := r.Names[lang]; ok && name != "" {
		return name
	}
	if name, ok := r.Names["en"]; ok && name != "" {
		return name
	}
	return r.Key
}
