package main

type Player struct {
	ID             string
	CurrentLobbyID string
	Name           string
	HealthPoints   int
	ActionPoints   int
	MovementPoints int
	IsMyTurn       bool
	PositionX      float32
	PositionY      float32
}

type Lobby struct {
	ID               string
	Players          [2]*Player
	CurrentTurnIndex int
	Grid             Grid
}

type Grid struct {
	Width  int
	Height int
	Cells  [][]GridCell
}

type GridCell struct {
	Position   Vector2
	IsOccupied bool
	Occupant   *string
}

type Vector2 struct {
	X int
	Y int
}

type Message struct {
	EventName string `json:"eventName"`
	Data      any    `json:"data"`
}
