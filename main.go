//flappy not floppy

package main

import (
	"image/color"
	"log"
	_ "time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/wav"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	screenWidth  = 1280
	screenHeight = 960

	backgroundWidth = 1280 * 4
	scrollSpeed     = 2

	menuOption_normal = iota
	menuOption_fast
	menuOption_faster

	GameStateMenu GameState = iota
	GameStatePlaying
)

var (
	player            *ebiten.Image
	background        *ebiten.Image
	scrollOffset      = 0
	currentMenuOption = menuOption_normal

	playerPosition  = vector{X: 100, Y: 350}
	playerVelocity  = vector{}
	playerScreenPos = vector{}

	continuousMovement = 0
	gravity            = 0.3
	jumpPower          = 10
	playerSpeed        = 2
	maxSpeed           = 5

	jumpPlayer   *audio.Player
)

var menuOptions = []string{"NORMAL", "FAST", "FASTER"}

type vector struct {
	X, Y float64
}

type GameState int

type Game struct {
	state GameState
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case GameStateMenu:
		g.drawMenu(screen)
	case GameStatePlaying:
		g.drawPlaying(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 80, A: 150})
	ebitenutil.DebugPrintAt(screen, "GOPHER FLAPPY", screenWidth/2-57, screenHeight/2-80)
	for i, pick := range menuOptions {
		textX := screenWidth/2 - len(pick)*6
		textY := screenHeight/2 + (i * 50)
		if i == currentMenuOption {
			ebitenutil.DebugPrintAt(screen, pick+" <=", textX, textY)
		} else {
			ebitenutil.DebugPrintAt(screen, pick, textX, textY)
		}
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-scrollOffset), 0)
	screen.DrawImage(background, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(playerScreenPos.X, playerScreenPos.Y)
	screen.DrawImage(player, op)
}

func (g *Game) updateMenu(screen *ebiten.Image) {

	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		currentMenuOption++
		if currentMenuOption >= len(menuOptions) {
			currentMenuOption = len(menuOptions) - 1
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		currentMenuOption--
		if currentMenuOption < 0 {
			currentMenuOption = 0
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch currentMenuOption {
		case 0:
			g.state = GameStatePlaying
			// playerPosition = playerPosition

		case 1:
			g.state = GameStatePlaying
			// playerPosition = playerPosition
			gravity *= 2
			jumpPower *= 2
			playerSpeed *= 2
			maxSpeed *= 2
		case 2:
			g.state = GameStatePlaying
			// playerPosition = playerPosition
			gravity *= 3
			jumpPower *= 3
			playerSpeed *= 3
			maxSpeed *= 3
		}
	}
}

func (g *Game) updatePlaying() {
	handleGameInput()
	movePlayer()
}

func handleGameInput() {
	continuousMovement = 0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		continuousMovement = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		continuousMovement = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		playJumpSound()
		// time.Sleep(10 * time.Millisecond)
		playerVelocity.Y = -float64(jumpPower)
	}
}

func movePlayer() {
	playerVelocity.Y += float64(gravity)
	if playerVelocity.Y > float64(maxSpeed) {
		playerVelocity.Y = float64(maxSpeed)
	}
	playerVelocity.X = float64(continuousMovement) * float64(playerSpeed)

	targetPlayerPosition := vector{
		X: playerPosition.X + playerVelocity.X,
		Y: playerPosition.Y + playerVelocity.Y,
	}

	playerPosition = targetPlayerPosition

	playerScreenPos.X = playerPosition.X - float64(scrollOffset)
	playerScreenPos.Y = playerPosition.Y

	if playerPosition.X < 0 {
		playerPosition.X = 0
	} else if playerPosition.X > float64(backgroundWidth-player.Bounds().Dx()) {
		playerPosition.X = float64(backgroundWidth - player.Bounds().Dx())
	}

	if playerPosition.Y > screenHeight-70 {
		playerPosition.Y = screenHeight - 70
		playerVelocity.Y = 0
	} else if playerPosition.Y < 0 {
		playerPosition.Y = 0
		playerVelocity.Y = 0
	}

	scrollOffset = int(playerPosition.X) - screenWidth/2
	if scrollOffset < 0 {
		scrollOffset = 0
	} else if scrollOffset > backgroundWidth-screenWidth {
		scrollOffset = backgroundWidth - screenWidth
	}

}

func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case GameStateMenu:
		g.updateMenu(screen)
	case GameStatePlaying:
		g.updatePlaying()
	}

	g.Draw(screen)

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func playJumpSound() {
	jumpPlayer.Play()
	jumpPlayer.Rewind()
}

func main() {
	var err error
	player, _, err = ebitenutil.NewImageFromFile("player.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	background, _, err = ebitenutil.NewImageFromFile("background.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	currentMenuOption = 0

	audioContext, err := audio.NewContext(12800)
	if err != nil {
		log.Fatal(err)
	}

	jumpSoundFile, err := ebitenutil.OpenFile("jump.wav")
	if err != nil {
		log.Fatal(err)
	}

	jumpSound, err := wav.Decode(audioContext, jumpSoundFile)
	if err != nil {
		log.Fatal(err)
	}

	jumpPlayer, err = audio.NewPlayer(audioContext, jumpSound)
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		state: GameStateMenu,
	}

	ebiten.SetWindowTitle("GOPHER FLAPPY")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
