package main

import (
	"log"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	GameStateMenu = iota
	GameStatePlaying
	GameStateWin
	GameStateLose
)

const (
	ScreenWidth     = 1280
	ScreenHeight    = 960
	BackgroundWidth = 1280 * 4
	ScrollSpeed     = 2
)

var(
	doorPosition       = Vector{X: BackgroundWidth - 300, Y: ScreenHeight / 2}
	menuOptions        = []string{"EASY", "NORMAL", "DIFFICULT", "", "EXIT"}
)

func main() {
	game := &Game{
		state:          GameStateMenu,
		PlayerOrigin:   Vector{X: 100, Y: 350},
		PlayerPosition: Vector{X: 100, Y: 350},
	}

	game.currentMenuOption = 0

	game.init()

	ebiten.SetWindowTitle("GoFlappy")
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.ActualFPS()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
