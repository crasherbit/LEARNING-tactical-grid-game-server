package main

func CreateGrid(width, height int) Grid {
	cells := make([][]GridCell, width)
	for x := 0; x < width; x++ {
		cells[x] = make([]GridCell, height)
		for y := 0; y < height; y++ {
			cells[x][y] = GridCell{
				Position:  Vector2{X: x, Y: y},
				IsOccupied: false,
				Occupant:   nil,
			}
		}
	}
	return Grid{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}