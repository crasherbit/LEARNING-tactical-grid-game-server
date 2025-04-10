package game

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// GameManager gestisce tutte le partite in corso
type GameManager struct {
	games map[string]*GameState
	mutex sync.RWMutex
}

// NewGameManager crea un nuovo GameManager
func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*GameState),
	}
}

// CreateGame crea una nuova partita e restituisce l'ID
func (gm *GameManager) CreateGame() string {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	// Genera un ID unico per la partita
	gameID := fmt.Sprintf("game-%d", time.Now().UnixNano())
	
	// Crea un nuovo stato di gioco
	game := &GameState{
		GameID:      gameID,
		Players:     make([]*Player, 0),
		CurrentTurn: 0,
		Status:      "WaitingForPlayers",
		Entities:    make([]Entity, 0),
		EntityData:  make([]BaseEntity, 0),
	}
	
	gm.games[gameID] = game
	return gameID
}

// GetGame restituisce lo stato di una partita per ID
func (gm *GameManager) GetGame(gameID string) (*GameState, error) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()
	
	game, exists := gm.games[gameID]
	if !exists {
		return nil, errors.New("partita non trovata")
	}
	return game, nil
}

// AddPlayer aggiunge un giocatore a una partita
func (gm *GameManager) AddPlayer(gameID string, playerID string, playerName string) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	game, exists := gm.games[gameID]
	if !exists {
		return errors.New("partita non trovata")
	}
	
	// Verifica se la partita è piena
	if len(game.Players) >= 2 {
		return errors.New("la partita è già al completo")
	}
	
	// Verifica se il giocatore è già nella partita
	for _, player := range game.Players {
		if player.ID == playerID {
			// Il giocatore è già nella partita, è un reconnect
			return nil
		}
	}
	
	// Determina l'ID del giocatore in base alla posizione
	// Per il primo giocatore: player1, per il secondo: player2
	actualPlayerID := playerID
	if len(game.Players) == 0 {
		actualPlayerID = "player1"
	} else {
		actualPlayerID = "player2"
	}
	
	// Crea il nuovo giocatore
	player := &Player{
		BaseEntity: BaseEntity{
			ID:        actualPlayerID,
			EntityType: "player",
			Health:    100,
			MaxHealth: 100,
		},
		Name:        playerName,
		ActionPoints: 5,
		Abilities:   []string{"fireball", "heal"},
	}
	
	// Posiziona il giocatore sulla griglia
	if len(game.Players) == 0 {
		player.Pos = Position{X: 0, Y: 0}
	} else {
		// Secondo giocatore all'angolo opposto
		player.Pos = Position{X: 9, Y: 9} // Assumendo una griglia 10x10
	}
	
	// Aggiunge il giocatore alla partita
	game.Players = append(game.Players, player)
	game.Entities = append(game.Entities, player)
	
	// Se ci sono due giocatori, inizia la partita
	if len(game.Players) == 2 {
		game.Status = "InProgress"
		game.CurrentTurn = 1
		game.CurrentPlayerID = game.Players[0].ID
		game.Players[0].ActionPoints = 5
	}
	
	// Aggiorna EntityData per la serializzazione
	game.updateEntityData()
	
	return nil
}

// ListGames restituisce un elenco di partite disponibili
func (gm *GameManager) ListGames() []map[string]interface{} {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()
	
	var gameList []map[string]interface{}
	
	for id, game := range gm.games {
		// Includi solo partite che stanno aspettando giocatori
		if game.Status == "WaitingForPlayers" {
			gameInfo := map[string]interface{}{
				"id":           id,
				"playerCount":  len(game.Players),
				"maxPlayers":   2,
				"status":       game.Status,
			}
			gameList = append(gameList, gameInfo)
		}
	}
	
	return gameList
}

// ProcessAction elabora un'azione di gioco
func (gm *GameManager) ProcessAction(action GameAction) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	game, exists := gm.games[action.GameID]
	if !exists {
		return errors.New("partita non trovata")
	}
	
	// Verifica che sia il turno del giocatore
	if game.CurrentPlayerID != action.PlayerID {
		return errors.New("non è il tuo turno")
	}
	
	// Trova il giocatore
	var player *Player
	for _, p := range game.Players {
		if p.ID == action.PlayerID {
			player = p
			break
		}
	}
	
	if player == nil {
		return errors.New("giocatore non trovato")
	}
	
	// Elabora l'azione in base al tipo
	switch action.Type {
	case ActionMove:
		return gm.processMove(game, player, action)
	case ActionUseAbility:
		return gm.processAbility(game, player, action)
	case ActionEndTurn:
		return gm.processEndTurn(game, player)
	default:
		return errors.New("tipo di azione non valido")
	}
}

// processMove elabora un'azione di movimento
func (gm *GameManager) processMove(game *GameState, player *Player, action GameAction) error {
	// Verifica che il movimento sia valido (adiacente)
	startPos := player.GetPosition()
	targetPos := action.TargetPosition
	
	dx := math.Abs(float64(targetPos.X - startPos.X))
	dy := math.Abs(float64(targetPos.Y - startPos.Y))
	
	if dx+dy != 1 {
		return errors.New("movimento non valido: puoi muoverti solo in celle adiacenti")
	}
	
	// Verifica punti azione
	if player.ActionPoints <= 0 {
		return errors.New("punti azione insufficienti")
	}
	
	// Verifica collisioni con altre entità (semplificato per ora)
	for _, entity := range game.Entities {
		if entity.GetID() != player.ID {
			pos := entity.GetPosition()
			if pos.X == targetPos.X && pos.Y == targetPos.Y {
				return errors.New("posizione occupata")
			}
		}
	}
	
	// Esegui il movimento
	player.SetPosition(targetPos)
	player.ActionPoints--
	
	// Aggiorna EntityData per la serializzazione
	game.updateEntityData()
	
	return nil
}

// processAbility elabora un'azione di utilizzo abilità
func (gm *GameManager) processAbility(game *GameState, player *Player, action GameAction) error {
	// Trova l'abilità
	// Per ora simuliamo alcune abilità di base
	abilities := map[string]Ability{
		"fireball": {
			ID:              "fireball",
			Name:            "Fireball",
			ActionPointCost: 2,
			Damage:          20,
			MinRange:        2,
			MaxRange:        5,
		},
		"heal": {
			ID:              "heal",
			Name:            "Heal",
			ActionPointCost: 3,
			Damage:          -25, // Negativo per la cura
			MinRange:        0,
			MaxRange:        1,
		},
	}
	
	ability, exists := abilities[action.AbilityID]
	if !exists {
		return errors.New("abilità non trovata")
	}
	
	// Verifica che il giocatore abbia questa abilità
	hasAbility := false
	for _, id := range player.Abilities {
		if id == action.AbilityID {
			hasAbility = true
			break
		}
	}
	
	if !hasAbility {
		return errors.New("il giocatore non ha questa abilità")
	}
	
	// Verifica punti azione
	if player.ActionPoints < ability.ActionPointCost {
		return errors.New("punti azione insufficienti")
	}
	
	// Verifica range
	startPos := player.GetPosition()
	targetPos := action.TargetPosition
	
	dx := math.Abs(float64(targetPos.X - startPos.X))
	dy := math.Abs(float64(targetPos.Y - startPos.Y))
	distance := int(dx + dy)
	
	if distance < ability.MinRange || distance > ability.MaxRange {
		return errors.New("bersaglio fuori portata")
	}
	
	// Trova il bersaglio (semplice implementazione)
	var target Entity
	for _, entity := range game.Entities {
		pos := entity.GetPosition()
		if pos.X == targetPos.X && pos.Y == targetPos.Y {
			target = entity
			break
		}
	}
	
	// Applica l'effetto dell'abilità
	if target != nil {
		if ability.Damage > 0 {
			// Danno
			target.SetHealth(target.GetHealth() - ability.Damage)
		} else {
			// Cura
			target.SetHealth(target.GetHealth() - ability.Damage) // Negativo diventa positivo
		}
	}
	
	// Consuma punti azione
	player.ActionPoints -= ability.ActionPointCost
	
	// Aggiorna EntityData per la serializzazione
	game.updateEntityData()
	
	return nil
}

// processEndTurn elabora la fine del turno
func (gm *GameManager) processEndTurn(game *GameState, player *Player) error {
	// Passa al giocatore successivo
	nextPlayerIndex := 0
	for i, p := range game.Players {
		if p.ID == player.ID {
			nextPlayerIndex = (i + 1) % len(game.Players)
			break
		}
	}
	
	// Incrementa il contatore dei turni se torniamo al primo giocatore
	if nextPlayerIndex == 0 {
		game.CurrentTurn++
	}
	
	// Imposta il nuovo giocatore corrente
	game.CurrentPlayerID = game.Players[nextPlayerIndex].ID
	
	// Ripristina i punti azione
	game.Players[nextPlayerIndex].ActionPoints = 5
	
	// Aggiorna EntityData per la serializzazione
	game.updateEntityData()
	
	return nil
}

// updateEntityData aggiorna EntityData per la serializzazione JSON
func (gs *GameState) updateEntityData() {
	gs.EntityData = make([]BaseEntity, 0, len(gs.Entities))
	for _, entity := range gs.Entities {
		if playerEntity, ok := entity.(*Player); ok {
			gs.EntityData = append(gs.EntityData, playerEntity.BaseEntity)
		} else if baseEntity, ok := entity.(*BaseEntity); ok {
			gs.EntityData = append(gs.EntityData, *baseEntity)
		}
	}
}