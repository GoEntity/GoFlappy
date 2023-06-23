package main

import (
	"image/color"
	"log"
	"os"

	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth     = 1280
	screenHeight    = 960
	backgroundWidth = 1280 * 4
	scrollSpeed     = 2

	menuOption_easy = iota
	menuOption_normal
	menuOption_difficult

	GameStateMenu GameState = iota
	GameStatePlaying

	DPI = 72

	playerSizeX = 50.0
	playerSizeY = 68.0
	enemySizeX  = 49.0
	enemySizeY  = 29.0

	gravity     = 0.5
	jumpPower   = 5
	playerSpeed = 2.5
	maxSpeed    = 5
)

var (
	player     *ebiten.Image
	background *ebiten.Image
	enemy      *ebiten.Image
	door       *ebiten.Image

	scrollOffset      = 0
	currentMenuOption = menuOption_easy

	playerOrigin    = vector{X: 100, Y: 350}
	playerPosition  = playerOrigin
	playerVelocity  = vector{}
	playerScreenPos = vector{}

	doorPosition = vector{X: backgroundWidth - 300, Y: screenHeight / 2}

	continuousMovement = 0

	jumpPlayer *audio.Player

	enemies       []Enemy
	enemyVelocity float64 = 5.0

	notPlaying bool = true
	win        bool = false

	titleFont font.Face
	miscFont  font.Face

	menuOptions = []string{"EASY", "NORMAL", "DIFFICULT"}
)

type vector struct {
	X, Y float64
}

type GameState int

type Game struct {
	state GameState
}

type Enemy struct {
	Position vector
	Velocity float64
	dirty    bool
}

func init() {
	//font
	coolFont, err := os.ReadFile("assets/fonts/Pixellettersfull.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(coolFont)
	if err != nil {
		log.Fatal(err)
	}
	titleFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    100,
		DPI:     DPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	miscFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    40,
		DPI:     DPI,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	//img
	player, _, err = ebitenutil.NewImageFromFile("assets/img/player.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	background, _, err = ebitenutil.NewImageFromFile("assets/img/background.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	enemy, _, err = ebitenutil.NewImageFromFile("assets/img/enemy.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	door, _, err = ebitenutil.NewImageFromFile("assets/img/door.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	//sound
	audioContext, err := audio.NewContext(44100)
	if err != nil {
		log.Fatal(err)
	}
	jumpSoundFile, err := ebitenutil.OpenFile("assets/sound/jump.mp3")
	if err != nil {
		log.Fatal(err)
	}
	jumpSound, err := mp3.Decode(audioContext, jumpSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	jumpPlayer, err = audio.NewPlayer(audioContext, jumpSound)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case GameStateMenu:
		g.drawMenu(screen)
	case GameStatePlaying:
		g.drawPlaying(screen)
		drawEnemies(screen)
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case GameStateMenu:
		g.updateMenu(screen)
	case GameStatePlaying:
		g.updatePlaying()
		removeEnemies()
	}
	g.Draw(screen)

	return nil
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 80, A: 150})

	if notPlaying {
		txt := "GoFlappy"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 80
		text.Draw(screen, txt, titleFont, x, y, color.White)

		for i, pick := range menuOptions {
			textX := screenWidth/2 - 100
			textY := screenHeight/2 + 50 + i*50
			if i == currentMenuOption {
				text.Draw(screen, ">> "+pick, miscFont, textX, textY, color.White)
			} else {
				text.Draw(screen, pick, miscFont, textX, textY, color.White)
			}
		}
	} else if !win && len(enemies) == 0 {
		enemies = []Enemy{}

		txt := "YOU ARE DEAD"
		txt2 := "Press Enter"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 60
		x2 := screenWidth/2 - text.BoundString(miscFont, txt2).Max.X/2
		y2 := screenHeight/2 + 30
		text.Draw(screen, txt, titleFont, x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		text.Draw(screen, txt2, miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			notPlaying = true
			g.drawMenu(screen)
		}
	} else if win {
		txt := "YOU ARE LEGENDARY"
		txt2 := "Press Enter"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 60
		x2 := screenWidth/2 - text.BoundString(miscFont, txt2).Max.X/2
		y2 := screenHeight/2 + 30
		text.Draw(screen, txt, titleFont, x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		text.Draw(screen, txt2, miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			notPlaying = true
			g.drawMenu(screen)
		}
	}
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
		if notPlaying {
			enemyVelocity = 5.0
			switch currentMenuOption {
			case 0:
			case 1:
				enemyVelocity += 2.5
			case 2:
				enemyVelocity += 5.0
			}
			scrollOffset = 0
			notPlaying = false
			g.state = GameStatePlaying
		} else if enemies == nil {
			g.state = GameStateMenu
			notPlaying = true
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
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(doorPosition.X-float64(scrollOffset), doorPosition.Y)
	screen.DrawImage(door, op)
}

func (g *Game) updatePlaying() {
	handleGameInput()
	g.movePlayer()

	if rand.Intn(100) < 7 { //spawn percent chance
		enemy := Enemy{
			Position: vector{X: float64(backgroundWidth), Y: rand.Float64() * screenHeight},
			Velocity: enemyVelocity + rand.Float64()*3,
		}
		enemies = append(enemies, enemy)
	}

	for _, e := range enemies {
		if collide(playerPosition, e.Position) {
			// log.Println("check: collision - enemy")

			enemies = nil
			g.state = GameStateMenu
			win = false

			playerPosition = playerOrigin

			break
		}
	}

	if collide(playerPosition, doorPosition) {
		// log.Println("check: collision - door")

		enemies = nil
		g.state = GameStateMenu
		win = true

		playerPosition = playerOrigin

		return
	}
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
		playerVelocity.Y = -float64(jumpPower)
	}
}

func (g *Game) movePlayer() {
	playerVelocity.Y += float64(gravity)

	if playerVelocity.Y > float64(maxSpeed) {
		playerVelocity.Y = float64(maxSpeed)
	}

	playerVelocity.X = float64(continuousMovement) * float64(playerSpeed)

	forPlayerPosition := vector{
		X: playerPosition.X + playerVelocity.X,
		Y: playerPosition.Y + playerVelocity.Y,
	}
	forPlayerScreenPos := vector{
		X: forPlayerPosition.X - float64(scrollOffset),
		Y: playerScreenPos.Y,
	}

	if forPlayerScreenPos.X < 0 || forPlayerScreenPos.X > screenWidth {
		playerVelocity.X = -2 * playerVelocity.X
	} else {
		playerPosition.X = forPlayerPosition.X
		playerScreenPos.X = forPlayerScreenPos.X
	}

	forPlayerScreenPos = vector{X: playerScreenPos.X, Y: forPlayerPosition.Y}
	if forPlayerScreenPos.Y < 0 || forPlayerScreenPos.Y > screenHeight-40 {
		playerVelocity.Y = -2 * playerVelocity.Y
	} else {
		playerPosition.Y = forPlayerPosition.Y
		playerScreenPos.Y = forPlayerScreenPos.Y
	}

	playerPosition.X += playerVelocity.X
	playerPosition.Y += playerVelocity.Y

	playerScreenPos.X = playerPosition.X - float64(scrollOffset)
	playerScreenPos.Y = playerPosition.Y

	if playerScreenPos.X < 0 {
		if playerVelocity.X < 0 {
			playerVelocity.X = -2 * playerVelocity.X
		}
	} else if playerScreenPos.X > screenWidth {
		if playerVelocity.X > 0 {
			playerVelocity.X = -2 * playerVelocity.X
		}
	}

	if playerScreenPos.Y < 0 {
		if playerVelocity.Y < 0 {
			playerVelocity.Y = -2 * playerVelocity.Y
		}
	} else if playerScreenPos.Y > screenHeight-40 {
		if playerVelocity.Y > 0 {
			playerVelocity.Y = -2 * playerVelocity.Y
		}
	}

	scrollOffset = int(playerPosition.X) - screenWidth/2

	if scrollOffset < 0 {
		scrollOffset = 0
	} else if scrollOffset > backgroundWidth-screenWidth {
		scrollOffset = backgroundWidth - screenWidth
	}

}

func playJumpSound() {
	jumpPlayer.Play()
	jumpPlayer.Rewind()
}

func drawEnemies(screen *ebiten.Image) {
	for _, e := range enemies {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.Position.X-float64(scrollOffset), e.Position.Y)
		screen.DrawImage(enemy, op)

		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset)), int(e.Position.Y))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset) + 49), int(e.Position.Y))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset)), int(e.Position.Y + 29))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset) + 49), int(e.Position.Y + 29))
	}
}

func removeEnemies() {
	for i := range enemies {
		enemies[i].Position.X -= enemies[i].Velocity
		if enemies[i].Position.X < -20 {
			enemies[i].dirty = true
		}
	}
	enemies = enemyRecycler(enemies)
}

func enemyRecycler(enemies []Enemy) []Enemy {
	filtered := enemies[:0]
	for _, e := range enemies {
		if !e.dirty {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func collide(a, b vector) bool {
	return !(a.X > b.X+enemySizeX ||
		a.X+playerSizeX < b.X ||
		a.Y > b.Y+enemySizeY ||
		a.Y+playerSizeY < b.Y)
}

func main() {
	defer jumpPlayer.Close()
	currentMenuOption = 0

	game := &Game{
		state: GameStateMenu,
	}

	ebiten.SetWindowTitle("GoFlappy")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(60)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
