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

	GameStateMenu = iota
	GameStatePlaying
	GameStateWin
	GameStateLose

	playerSizeX = 50.0
	playerSizeY = 68.0
	enemySizeX  = 49.0
	enemySizeY  = 29.0
	PlayerGravity   = 0.5
	PlayerJumpPower = 5
	PlayerSpeed     = 2.5
	PlayerMaxSpeed  = 5
	EnemySpeed float64 = 5.0
)

var (
	player     *ebiten.Image
	background *ebiten.Image
	enemy      *ebiten.Image
	door       *ebiten.Image
	jumpPlayer *audio.Player
	menuPlayer *audio.Player
	gamePlayer *audio.Player
	winPlayer  *audio.Player
	losePlayer *audio.Player
	hellPlayer *audio.Player	

	gravity   = PlayerGravity
	jumpPower = PlayerJumpPower
	speed     = PlayerSpeed
	maxSpeed  = PlayerMaxSpeed

	doorPosition = vector{X: backgroundWidth - 300, Y: screenHeight / 2}

	continuousMovement = 0

	enemies    []Enemy
	enemySpeed = EnemySpeed

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
	scrollOffset      int
	currentMenuOption int
	PlayerOrigin    vector
	PlayerPosition  vector
	PlayerVelocity  vector
	PlayerScreenPos vector
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
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	miscFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    40,
		DPI:     72,
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
	menuSoundFile, err := ebitenutil.OpenFile("assets/sound/menuBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	menuSound, err := mp3.Decode(audioContext, menuSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	menuPlayer, err = audio.NewPlayer(audioContext, menuSound)
	if err != nil {
		log.Fatal(err)
	}
	gameSoundFile, err := ebitenutil.OpenFile("assets/sound/gameBGM_long.mp3")
	if err != nil {
		log.Fatal(err)
	}
	gameSound, err := mp3.Decode(audioContext, gameSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	gamePlayer, err = audio.NewPlayer(audioContext, gameSound)
	if err != nil {
		log.Fatal(err)
	}
	loseSoundFile, err := ebitenutil.OpenFile("assets/sound/loseBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	loseSound, err := mp3.Decode(audioContext, loseSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	losePlayer, err = audio.NewPlayer(audioContext, loseSound)
	if err != nil {
		log.Fatal(err)
	}
	winSoundFile, err := ebitenutil.OpenFile("assets/sound/winBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	winSound, err := mp3.Decode(audioContext, winSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	winPlayer, err = audio.NewPlayer(audioContext, winSound)
	if err != nil {
		log.Fatal(err)
	}
	hellSoundFile, err := ebitenutil.OpenFile("assets/sound/gameBGM_hell.mp3")
	if err != nil {
		log.Fatal(err)
	}
	hellSound, err := mp3.Decode(audioContext, hellSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	hellPlayer, err = audio.NewPlayer(audioContext, hellSound)
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
		g.drawEnemies(screen)
	case GameStateWin:
		g.drawMenu(screen)
	case GameStateLose:
		g.drawMenu(screen)
	}
}

func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case GameStateMenu:
		menuPlayer.Play()
		winPlayer.Pause()
		winPlayer.Seek(0)
		losePlayer.Pause()
		losePlayer.Seek(0)
		g.updateMenu(screen)
	case GameStatePlaying:
		if g.currentMenuOption == 2 {
			hellPlayer.Play()
		} else {
			gamePlayer.Play()
		}
		menuPlayer.Pause()
		menuPlayer.Seek(0)
		g.updatePlaying()
		removeEnemies()
	case GameStateWin:
		gamePlayer.Pause()
		gamePlayer.Seek(0)
		hellPlayer.Pause()
		hellPlayer.Seek(0)
		winPlayer.Play()
		g.updateMenu(screen)
	case GameStateLose:
		gamePlayer.Pause()
		gamePlayer.Seek(0)
		hellPlayer.Pause()
		hellPlayer.Seek(0)
		losePlayer.Play()
		g.updateMenu(screen)
	}
	return nil
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{R: 0, G: 0, B: 80, A: 150})

	if g.state == GameStateMenu {
		txt := "GoFlappy"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 80
		text.Draw(screen, txt, titleFont, x, y, color.White)

		for i, pick := range menuOptions {
			textX := screenWidth/2 - 100
			textY := screenHeight/2 + 50 + i*50
			if i == g.currentMenuOption {
				text.Draw(screen, ">> "+pick, miscFont, textX, textY, color.White)
			} else {
				text.Draw(screen, pick, miscFont, textX, textY, color.White)
			}
		}
	} else if g.state == GameStateLose && len(enemies) == 0 {
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
			g.state = GameStateMenu
		}
	} else if g.state == GameStateWin && len(enemies) == 0 {
		txt := "YOU ARE LEGENDARY"
		txt2 := "Press Enter"
		x := screenWidth/2 - text.BoundString(titleFont, txt).Max.X/2
		y := screenHeight/2 - 60
		x2 := screenWidth/2 - text.BoundString(miscFont, txt2).Max.X/2
		y2 := screenHeight/2 + 30
		text.Draw(screen, txt, titleFont, x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		text.Draw(screen, txt2, miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = GameStateMenu
		}
	}
}

func (g *Game) updateMenu(screen *ebiten.Image) {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.currentMenuOption++
		if g.currentMenuOption >= len(menuOptions) {
			g.currentMenuOption = len(menuOptions) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.currentMenuOption--
		if g.currentMenuOption < 0 {
			g.currentMenuOption = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.state == GameStateMenu {
			enemySpeed = EnemySpeed
			speed = PlayerSpeed
			switch g.currentMenuOption {
			case 0:
				enemySpeed += 2.5
			case 1:
				enemySpeed += 5.0
			case 2:
				enemySpeed += 11.0
				speed += 1.0
			}
			g.scrollOffset = 0
			g.state = GameStatePlaying
		} else if g.state == GameStateWin || g.state == GameStateLose {
			g.state = GameStateMenu
		}
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-g.scrollOffset), 0)
	screen.DrawImage(background, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.PlayerScreenPos.X, g.PlayerScreenPos.Y)
	screen.DrawImage(player, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(doorPosition.X-float64(g.scrollOffset), doorPosition.Y)
	screen.DrawImage(door, op)
}

func (g *Game) updatePlaying() {
	g.handleGameInput()
	g.movePlayer()

	if rand.Intn(100) < 7 { //spawn percent chance
		enemy := Enemy{
			Position: vector{X: float64(backgroundWidth), Y: rand.Float64() * screenHeight},
			Velocity: enemySpeed + rand.Float64()*3,
		}
		enemies = append(enemies, enemy)
	}

	for _, e := range enemies {
		if collide(g.PlayerPosition, e.Position) {
			//log.Println("check: collision - enemy")
			enemies = nil
			g.state = GameStateLose
			g.PlayerPosition = g.PlayerOrigin

			break
		}
	}

	if collide(g.PlayerPosition, doorPosition) {
		//log.Println("check: collision - door")
		enemies = nil
		g.state = GameStateWin
		g.PlayerPosition = g.PlayerOrigin

		return
	}
}

func (g *Game) handleGameInput() {
	continuousMovement = 0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		continuousMovement = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		continuousMovement = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		playJumpSound()
		g.PlayerVelocity.Y = -float64(jumpPower)
	}
}

func (g *Game) movePlayer() {
	g.PlayerVelocity.Y += float64(gravity)

	if g.PlayerVelocity.Y > float64(maxSpeed) {
		g.PlayerVelocity.Y = float64(maxSpeed)
	}

	g.PlayerVelocity.X = float64(continuousMovement) * float64(speed)

	forPlayerPosition := vector{
		X: g.PlayerPosition.X + g.PlayerVelocity.X,
		Y: g.PlayerPosition.Y + g.PlayerVelocity.Y,
	}
	forPlayerScreenPos := vector{
		X: forPlayerPosition.X - float64(g.scrollOffset),
		Y: g.PlayerScreenPos.Y,
	}

	if forPlayerScreenPos.X < 0 || forPlayerScreenPos.X > screenWidth {
		g.PlayerVelocity.X = -2 * g.PlayerVelocity.X
	} else {
		g.PlayerPosition.X = forPlayerPosition.X
		g.PlayerScreenPos.X = forPlayerScreenPos.X
	}

	forPlayerScreenPos = vector{X: g.PlayerScreenPos.X, Y: forPlayerPosition.Y}
	if forPlayerScreenPos.Y < 0 || forPlayerScreenPos.Y > screenHeight-40 {
		g.PlayerVelocity.Y = -2 * g.PlayerVelocity.Y
	} else {
		g.PlayerPosition.Y = forPlayerPosition.Y
		g.PlayerScreenPos.Y = forPlayerScreenPos.Y
	}

	g.PlayerPosition.X += g.PlayerVelocity.X
	g.PlayerPosition.Y += g.PlayerVelocity.Y

	g.PlayerScreenPos.X = g.PlayerPosition.X - float64(g.scrollOffset)
	g.PlayerScreenPos.Y = g.PlayerPosition.Y

	if g.PlayerScreenPos.X < 0 {
		if g.PlayerVelocity.X < 0 {
			g.PlayerVelocity.X = -2 * g.PlayerVelocity.X
		}
	} else if g.PlayerScreenPos.X > screenWidth {
		if g.PlayerVelocity.X > 0 {
			g.PlayerVelocity.X = -2 * g.PlayerVelocity.X
		}
	}

	if g.PlayerScreenPos.Y < 0 {
		if g.PlayerVelocity.Y < 0 {
			g.PlayerVelocity.Y = -2 * g.PlayerVelocity.Y
		}
	} else if g.PlayerScreenPos.Y > screenHeight-40 {
		if g.PlayerVelocity.Y > 0 {
			g.PlayerVelocity.Y = -2 * g.PlayerVelocity.Y
		}
	}

	g.scrollOffset = int(g.PlayerPosition.X) - screenWidth/2

	if g.scrollOffset < 0 {
		g.scrollOffset = 0
	} else if g.scrollOffset > backgroundWidth-screenWidth {
		g.scrollOffset = backgroundWidth - screenWidth
	}

}

func playJumpSound() {
	jumpPlayer.Play()
	jumpPlayer.Rewind()
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	for _, e := range enemies {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.Position.X-float64(g.scrollOffset), e.Position.Y)
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
	game := &Game{
		state: GameStateMenu,
		scrollOffset: 0,
		currentMenuOption: menuOption_easy,
		PlayerOrigin    : vector{X: 100, Y: 350},
		PlayerPosition  : vector{X: 100, Y: 350},
		PlayerVelocity  : vector{},
		PlayerScreenPos : vector{},
	}

	game.currentMenuOption = 0

	ebiten.SetWindowTitle("GoFlappy")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(60)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}