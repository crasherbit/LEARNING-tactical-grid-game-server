package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func NotifyPlayer(playerID string, response Message) {
	conn, exists := playerConnections[playerID]
	if exists {
		responseJSON, _ := json.Marshal(response)
		err := conn.WriteMessage(websocket.TextMessage, responseJSON)
		if err != nil {
			fmt.Printf("Errore inviando il messaggio al giocatore %s: %v\n", playerID, err)
		}
	}
}

func GenerateLobbyID() string {
	return fmt.Sprintf("lobby-%d", len(lobbies)+1)
}

func SendLobbyState(lobby *Lobby) {
	for _, player := range lobby.Players {
		if player != nil {
			NotifyPlayer(player.ID, Message{
				EventName: "lobby_state",
				Data: map[string]any{
					"lobby_id":     lobby.ID,
					"grid":         lobby.Grid,
					"players":      lobby.Players,
					"current_turn": lobby.Players[lobby.CurrentTurnIndex].ID,
				},
			})
		}
	}
}