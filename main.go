package main

import (
	"image/color"
	"log"

	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/audio"
	"github.com/hajimehoshi/ebiten/audio/mp3"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	forFont "github.com/goentity/fonts"
	forImage "github.com/goentity/image"
	forSound "github.com/goentity/sound"
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
	player     *ebiten.Image
	background *ebiten.Image
	enemy      *ebiten.Image
	door       *ebiten.Image

	scrollOffset      = 0
	currentMenuOption = menuOption_normal

	playerPosition  = vector{X: 100, Y: 350}
	playerVelocity  = vector{}
	playerScreenPos = vector{}

	doorPosition = vector{X: backgroundWidth - 300, Y: screenHeight / 2}

	continuousMovement = 0
	gravity            = 0.5
	jumpPower          = 5
	playerSpeed        = 2.5
	maxSpeed           = 5

	jumpPlayer *audio.Player

	enemies       []Enemy
	enemyVelocity float64 = 5.0

	firstLaunch bool = true
	win         bool = false

	titleFont font.Face
	miscFont  font.Face
)

func init() {
	tt, err := opentype.Parse(fonts.Pixellettersfull-BnJ5_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	titleFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    200,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	miscFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    100,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// func forCustomFace(textSize float64) font.Face {
// 	customFaceLocal := truetype.NewFace(customFont, &truetype.Options{Size: textSize})
// 	return customFaceLocal
// }

var menuOptions = []string{"EASY", "NORMAL", "DIFFICULT"}

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

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 80, A: 150})

	if firstLaunch {
		txt := "GoFlappy"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 80
		text.Draw(screen, txt, titleFont, x, y, color.White)

		// text.Draw(screen, txt, customFace, x, y, color.White)
		// textSize = 30.0
		for i, pick := range menuOptions {
			textX := screenWidth/2 - 100
			textY := screenHeight/2 + 50 + i*50
			if i == currentMenuOption {
				// ebitenutil.DebugPrintAt(screen, pick+" <=", textX, textY)
				// text.Draw(screen, ">> "+pick, forCustomFace(60.0), textX, textY, color.White)
				text.Draw(screen, ">> "+pick, miscFont, textX, textY, color.White)
			} else {
				// ebitenutil.DebugPrintAt(screen, pick, textX, textY)
				// text.Draw(screen, pick, forCustomFace(60.0), textX, textY, color.White)
				text.Draw(screen, pick, miscFont, textX, textY, color.White)
			}
		}
	} else if !win && len(enemies) == 0 {
		enemies = []Enemy{}
		// ebitenutil.DebugPrintAt(screen, "Don't give up...", screenWidth/2-40, screenHeight/2-60)
		// ebitenutil.DebugPrintAt(screen, "Press Enter", screenWidth/2-30, screenHeight/2+30)
		txt := "YOU ARE DEAD"
		txt2 := "Press Enter"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 60
		x2 := screenWidth/2 - text.BoundString(miscFont, txt2).Max.X/2
		y2 := screenHeight/2 + 30
		text.Draw(screen, txt, titleFont, x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		text.Draw(screen, txt2, miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			firstLaunch = true
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
			firstLaunch = true
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
		if firstLaunch {
			enemyVelocity = 5.0
			switch currentMenuOption {
			case 0:
				g.state = GameStatePlaying
				firstLaunch = false
			case 1:
				enemyVelocity += 2.5
				g.state = GameStatePlaying
				firstLaunch = false
				// temp
				// gravity *= 2
				// jumpPower *= 2
				// playerSpeed *= 2
				// maxSpeed *= 2
			case 2:
				enemyVelocity += 5.0
				g.state = GameStatePlaying
				firstLaunch = false
				// temp
				// gravity *= 3
				// jumpPower *= 3
				// playerSpeed *= 3
				// maxSpeed *= 3
			}
			scrollOffset = 0
		} else if enemies == nil {
			g.state = GameStateMenu
			firstLaunch = true
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

			playerPosition = vector{X: 100, Y: 350}

			break
		}
	}

	if collide(playerPosition, doorPosition) {
		// log.Println("check: collision - door")

		enemies = nil
		g.state = GameStateMenu
		win = true

		playerPosition = vector{X: 100, Y: 350}

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

	forPlayerPosition := vector{X: playerPosition.X + playerVelocity.X, Y: playerPosition.Y + playerVelocity.Y}

	forPlayerScreenPos := vector{X: forPlayerPosition.X - float64(scrollOffset), Y: playerScreenPos.Y}
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
		playerPosition.X = 0
		playerScreenPos.X = 0
		if playerVelocity.X < 0 {
			playerVelocity.X = -2 * playerVelocity.X
		}
	} else if playerScreenPos.X > screenWidth {
		playerPosition.X = screenWidth + float64(scrollOffset)
		playerScreenPos.X = screenWidth
		if playerVelocity.X > 0 {
			playerVelocity.X = -2 * playerVelocity.X
		}
	}

	if playerScreenPos.Y < 0 {
		playerPosition.Y = 0
		playerScreenPos.Y = 0
		if playerVelocity.Y < 0 {
			playerVelocity.Y = -2 * playerVelocity.Y
		}
	} else if playerScreenPos.Y > screenHeight-40 {
		playerPosition.Y = screenHeight - 40
		playerScreenPos.Y = screenHeight - 40
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
	playerSizeX := 50.0
	playerSizeY := 68.0
	targetSizeX := 49.0 //going with enemy dimension for all for now.. cuz im lazy
	targetSizeY := 29.0

	return !(a.X > b.X+targetSizeX ||
		a.X+playerSizeX < b.X ||
		a.Y > b.Y+targetSizeY ||
		a.Y+playerSizeY < b.Y)
}

func main() {
	var err error
	player, _, err = ebitenutil.NewImageFromFile("img/player.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	background, _, err = ebitenutil.NewImageFromFile("img/background.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	enemy, _, err = ebitenutil.NewImageFromFile("img/enemy.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	door, _, err = ebitenutil.NewImageFromFile("img/door.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	currentMenuOption = 0

	audioContext, err := audio.NewContext(12800)
	if err != nil {
		log.Fatal(err)
	}

	jumpSoundFile, err := ebitenutil.OpenFile("sound/jump.mp3")
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

	defer jumpPlayer.Close()

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
