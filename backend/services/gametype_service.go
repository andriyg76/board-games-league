package services

import (
	"context"
	"embed"
	"github.com/andriyg76/bgl/models"
	"github.com/andriyg76/bgl/repositories"
	"github.com/andriyg76/glog"
	"gopkg.in/yaml.v3"
)

//go:embed data/games.yaml
var gamesYAML embed.FS

// GameTypeService сервіс для роботи з типами ігор
type GameTypeService interface {
	LoadBuiltInGames(ctx context.Context) error
}

type gameTypeService struct {
	repo repositories.GameTypeRepository
}

// NewGameTypeService створює новий сервіс типів ігор
func NewGameTypeService(repo repositories.GameTypeRepository) GameTypeService {
	return &gameTypeService{repo: repo}
}

// Структури для парсингу YAML

type gamesYAMLFile struct {
	Games []gameYAML `yaml:"games"`
}

type gameYAML struct {
	Key         string            `yaml:"key"`
	Names       map[string]string `yaml:"names"`
	Icon        string            `yaml:"icon"`
	ScoringType string            `yaml:"scoring_type"`
	MinPlayers  int               `yaml:"min_players"`
	MaxPlayers  int               `yaml:"max_players"`
	Roles       []roleYAML        `yaml:"roles"`
}

type roleYAML struct {
	Key      string            `yaml:"key"`
	Names    map[string]string `yaml:"names"`
	Color    string            `yaml:"color"`
	Icon     string            `yaml:"icon"`
	RoleType string            `yaml:"role_type"`
}

// LoadBuiltInGames завантажує вбудовані типи ігор з YAML файлу
func (s *gameTypeService) LoadBuiltInGames(ctx context.Context) error {
	data, err := gamesYAML.ReadFile("data/games.yaml")
	if err != nil {
		return glog.Error("failed to read embedded games.yaml: %w", err)
	}

	var file gamesYAMLFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return glog.Error("failed to parse games.yaml: %w", err)
	}

	for _, g := range file.Games {
		gameType := yamlToModel(g)
		if err := s.repo.Upsert(ctx, gameType); err != nil {
			glog.Warn("failed to upsert game type %s: %v", g.Key, err)
			continue
		}
		glog.Info("Loaded built-in game type: %s", g.Key)
	}

	return nil
}

func yamlToModel(g gameYAML) *models.GameType {
	roles := make([]models.Role, len(g.Roles))
	for i, r := range g.Roles {
		roles[i] = models.Role{
			Key:      r.Key,
			Names:    r.Names,
			Color:    r.Color,
			Icon:     r.Icon,
			RoleType: models.RoleType(r.RoleType),
		}
	}

	return &models.GameType{
		Key:         g.Key,
		Names:       g.Names,
		Icon:        g.Icon,
		ScoringType: models.ScoringType(g.ScoringType),
		Roles:       roles,
		MinPlayers:  g.MinPlayers,
		MaxPlayers:  g.MaxPlayers,
		BuiltIn:     true,
	}
}
