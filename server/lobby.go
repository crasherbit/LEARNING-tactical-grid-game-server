package main

import (
	"fmt"
	"math/rand"
	"sync"
)

var lobbies = make(map[string]*Lobby)
var lobbyMutex sync.Mutex

func createLobby(player *Player) *Lobby {
	lobby := &Lobby{
		ID:              GenerateLobbyID(),
		Players:         [2]*Player{player, nil},
		CurrentTurnIndex: 0,
		Grid:            CreateGrid(8, 8),
	}
	lobbies[lobby.ID] = lobby
	player.CurrentLobbyID = lobby.ID
	fmt.Printf("Lobby %s creata per il giocatore %s\n", lobby.ID, player.ID)
	return lobby
}

func JoinLobby(player *Player) {
	lobbyMutex.Lock()
	defer lobbyMutex.Unlock()

	for _, lobby := range lobbies {
		if lobby.Players[1] == nil {
			lobby.Players[1] = player
			player.CurrentLobbyID = lobby.ID
			randomizeTurnOrder(lobby)
			notifyLobbyReady(lobby)
			return
		}
	}

	// Nessuna lobby disponibile, creane una nuova
	createLobby(player)
}

func randomizeTurnOrder(lobby *Lobby) {
	lobby.CurrentTurnIndex = rand.Intn(2)
	lobby.Players[lobby.CurrentTurnIndex].IsMyTurn = true
	lobby.Players[1-lobby.CurrentTurnIndex].IsMyTurn = false
}

func notifyLobbyReady(lobby *Lobby) {
	for _, player := range lobby.Players {
		if player != nil {
			NotifyPlayer(player.ID, Message{
				EventName: "lobby_ready",
				Data: map[string]any{
					"lobby_id":    lobby.ID,
					"players":     lobby.Players,
					"my_player_id": player.ID,
					"current_turn": lobby.Players[lobby.CurrentTurnIndex],
				},
			})
		}
	}
}