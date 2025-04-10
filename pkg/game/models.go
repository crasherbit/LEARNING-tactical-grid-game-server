package game

// Position rappresenta una posizione sulla griglia
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Entity è l'interfaccia base per tutte le entità di gioco
type Entity interface {
	GetID() string
	GetPosition() Position
	SetPosition(Position)
	GetHealth() int
	SetHealth(int)
	GetMaxHealth() int
}

// BaseEntity implementa i campi comuni per tutte le entità
type BaseEntity struct {
	ID        string   `json:"id"`
	EntityType string   `json:"type"`
	Pos       Position `json:"position"`
	Health    int      `json:"health"`
	MaxHealth int      `json:"maxHealth"`
	OwnerID   string   `json:"ownerId,omitempty"`
}

// GetID restituisce l'ID dell'entità
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetPosition restituisce la posizione dell'entità
func (e *BaseEntity) GetPosition() Position {
	return e.Pos
}

// SetPosition imposta la posizione dell'entità
func (e *BaseEntity) SetPosition(pos Position) {
	e.Pos = pos
}

// GetHealth restituisce la salute attuale dell'entità
func (e *BaseEntity) GetHealth() int {
	return e.Health
}

// SetHealth imposta la salute dell'entità
func (e *BaseEntity) SetHealth(health int) {
	e.Health = health
	if e.Health > e.MaxHealth {
		e.Health = e.MaxHealth
	}
	if e.Health < 0 {
		e.Health = 0
	}
}

// GetMaxHealth restituisce la salute massima dell'entità
func (e *BaseEntity) GetMaxHealth() int {
	return e.MaxHealth
}

// Player rappresenta un giocatore nel gioco
type Player struct {
	BaseEntity
	Name        string   `json:"name"`
	ActionPoints int      `json:"actionPoints"`
	Abilities   []string `json:"abilityIds"`
}

// Ability rappresenta un'abilità utilizzabile dai giocatori
type Ability struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	IconName       string     `json:"iconName"`
	MinRange       int        `json:"minRange"`
	MaxRange       int        `json:"maxRange"`
	ActionPointCost int        `json:"actionPointCost"`
	NeedsLineOfSight bool       `json:"needsLineOfSight"`
	RangeType      string     `json:"rangeType"`
	TargetType     string     `json:"targetType"`
	Damage         int        `json:"damage"`
	Description    string     `json:"description"`
}

// GameState rappresenta lo stato corrente di una partita
type GameState struct {
	GameID         string    `json:"gameId"`
	Players        []*Player `json:"players"`
	CurrentTurn    int       `json:"currentTurn"`
	CurrentPlayerID string    `json:"currentPlayerId"`
	Status         string    `json:"status"`
	Entities       []Entity  `json:"-"` // Non serializzato direttamente
	EntityData     []BaseEntity `json:"entities"` // Usato per la serializzazione
}

// ActionType definisce i tipi di azione possibili
type ActionType string

const (
	ActionMove     ActionType = "Move"
	ActionUseAbility ActionType = "UseAbility"
	ActionEndTurn    ActionType = "EndTurn"
)

// GameAction rappresenta un'azione di gioco inviata dal client
type GameAction struct {
	GameID         string     `json:"gameId"`
	PlayerID       string     `json:"playerId"`
	Type           ActionType `json:"type"`
	StartPosition  Position   `json:"startPosition"`
	TargetPosition Position   `json:"targetPosition"`
	AbilityID      string     `json:"abilityId"`
	TargetIDs      []string   `json:"targetIds"`
}