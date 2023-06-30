package main

import (
	"errors"
	"image/color"

	// _ "image/png"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type GameState int

type Game struct {
	state             GameState
	player            *ebiten.Image
	background        *ebiten.Image
	enemy             *ebiten.Image
	door              *ebiten.Image
	jumpPlayer        *audio.Player
	menuPlayer        *audio.Player
	gamePlayer        *audio.Player
	winPlayer         *audio.Player
	losePlayer        *audio.Player
	hellPlayer        *audio.Player
	scrollOffset      int
	currentMenuOption int
	PlayerOrigin      Vector
	PlayerPosition    Vector
	PlayerVelocity    Vector
	PlayerScreenPos   Vector
	titleFont         font.Face
	miscFont          font.Face
}

func (g *Game) init() {
	//font
	coolFont, err := os.ReadFile("assets/fonts/Pixellettersfull.ttf")
	if err != nil {
		log.Fatal(err)
	}
	tt, err := opentype.Parse(coolFont)
	if err != nil {
		log.Fatal(err)
	}
	g.titleFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    100,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	g.miscFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    40,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	//img
	g.player, _, err = ebitenutil.NewImageFromFile("assets/img/player.png")
	if err != nil {
		log.Fatal(err)
	}
	g.background, _, err = ebitenutil.NewImageFromFile("assets/img/background.png")
	if err != nil {
		log.Fatal(err)
	}
	g.enemy, _, err = ebitenutil.NewImageFromFile("assets/img/enemy.png")
	if err != nil {
		log.Fatal(err)
	}
	g.door, _, err = ebitenutil.NewImageFromFile("assets/img/door.png")
	if err != nil {
		log.Fatal(err)
	}

	//sound
	audioContext := audio.NewContext(44100)
	fourfouronezerozero := 44100
	jumpSoundFile, err := os.Open("assets/sound/jump.mp3")
	if err != nil {
		log.Fatal(err)
	}
	jumpSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, jumpSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.jumpPlayer, err = audioContext.NewPlayer(jumpSound)
	if err != nil {
		log.Fatal(err)
	}
	menuSoundFile, err := os.Open("assets/sound/menuBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	menuSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, menuSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.menuPlayer, err = audioContext.NewPlayer(menuSound)
	if err != nil {
		log.Fatal(err)
	}
	gameSoundFile, err := os.Open("assets/sound/gameBGM_long.mp3")
	if err != nil {
		log.Fatal(err)
	}
	gameSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, gameSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.gamePlayer, err = audioContext.NewPlayer(gameSound)
	if err != nil {
		log.Fatal(err)
	}
	loseSoundFile, err := os.Open("assets/sound/loseBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	loseSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, loseSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.losePlayer, err = audioContext.NewPlayer(loseSound)
	if err != nil {
		log.Fatal(err)
	}
	winSoundFile, err := os.Open("assets/sound/winBGM.mp3")
	if err != nil {
		log.Fatal(err)
	}
	winSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, winSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.winPlayer, err = audioContext.NewPlayer(winSound)
	if err != nil {
		log.Fatal(err)
	}
	hellSoundFile, err := os.Open("assets/sound/gameBGM_hell.mp3")
	if err != nil {
		log.Fatal(err)
	}
	hellSound, err := mp3.DecodeWithSampleRate(fourfouronezerozero, hellSoundFile)
	if err != nil {
		log.Fatal(err)
	}
	g.hellPlayer, err = audioContext.NewPlayer(hellSound)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case GameStateMenu:
		g.DrawMenu(screen)
	case GameStatePlaying:
		g.DrawPlaying(screen)
		g.DrawEnemies(screen)
	case GameStateWin:
		g.DrawMenu(screen)
	case GameStateLose:
		g.DrawMenu(screen)
	}
}

func (g *Game) Update() error {
	switch g.state {
	case GameStateMenu:
		g.menuPlayer.Play()
		g.winPlayer.Pause()
		g.winPlayer.Seek(0)
		g.losePlayer.Pause()
		g.losePlayer.Seek(0)
		// g.updateMenu(screen)
		if err := g.UpdateMenu(); err != nil {
			return err
		}
	case GameStatePlaying:
		if g.currentMenuOption == 2 {
			g.hellPlayer.Play()
		} else {
			g.gamePlayer.Play()
		}
		g.menuPlayer.Pause()
		g.menuPlayer.Seek(0)
		g.UpdatePlaying()
		RemoveEnemies()
	case GameStateWin:
		g.gamePlayer.Pause()
		g.gamePlayer.Seek(0)
		g.hellPlayer.Pause()
		g.hellPlayer.Seek(0)
		g.winPlayer.Play()
		g.UpdateMenu()
	case GameStateLose:
		g.gamePlayer.Pause()
		g.gamePlayer.Seek(0)
		g.hellPlayer.Pause()
		g.hellPlayer.Seek(0)
		g.losePlayer.Play()
		g.UpdateMenu()
	}
	return nil
}

func (g *Game) UpdateMenu() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.currentMenuOption++
		if g.currentMenuOption == 3 {
			g.currentMenuOption++
		}
		if g.currentMenuOption >= len(menuOptions) {
			g.currentMenuOption = len(menuOptions) - 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.currentMenuOption--
		if g.currentMenuOption == 3 {
			g.currentMenuOption--
		}
		if g.currentMenuOption < 0 {
			g.currentMenuOption = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.state == GameStateMenu {
			enemySpeed = 5.0
			speed = 2.5
			switch g.currentMenuOption {
			case 0:
				enemySpeed += 2.5
			case 1:
				enemySpeed += 5.0
			case 2:
				enemySpeed += 11.0
				speed += 1.0
			case 4:
				// os.Exit(0) // doesn't close resources.. doesn't work with defer.. use error instead
				return errors.New("exit")
			}
			g.scrollOffset = 0
			g.state = GameStatePlaying
		} else if g.state == GameStateWin || g.state == GameStateLose {
			g.state = GameStateMenu
		}
	}
	return nil
}

func (g *Game) DrawMenu(screen *ebiten.Image) {
	clr := color.RGBA{R: 0, G: 0, B: 80, A: 150}
	// ebitenutil.DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 0, G: 0, B: 80, A: 150})
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, clr, false)

	if g.state == GameStateMenu {
		txt := "GoFlappy"
		x := ScreenWidth/2 - text.BoundString(g.titleFont, txt).Max.X/2
		y := ScreenHeight/2 - 80
		text.Draw(screen, txt, g.titleFont, x, y, color.White)

		for i, pick := range menuOptions {
			textX := ScreenWidth/2 - 100
			textY := ScreenHeight/2 + 50 + i*50
			if i == g.currentMenuOption {
				if i != 3 {
					text.Draw(screen, ">> "+pick, g.miscFont, textX, textY, color.White)
				}
			} else {
				text.Draw(screen, pick, g.miscFont, textX, textY, color.White)
			}
		}
	} else if g.state == GameStateLose && len(enemies) == 0 {
		enemies = []Enemy{}

		txt := "YOU ARE DEAD"
		txt2 := "Press Enter"
		x := ScreenWidth/2 - text.BoundString(g.titleFont, txt).Max.X/2
		y := ScreenHeight/2 - 60
		x2 := ScreenWidth/2 - text.BoundString(g.miscFont, txt2).Max.X/2
		y2 := ScreenHeight/2 + 30
		text.Draw(screen, txt, g.titleFont, x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		text.Draw(screen, txt2, g.miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = GameStateMenu
		}
	} else if g.state == GameStateWin && len(enemies) == 0 {
		txt := "YOU ARE LEGENDARY"
		txt2 := "Press Enter"
		x := ScreenWidth/2 - text.BoundString(g.titleFont, txt).Max.X/2
		y := ScreenHeight/2 - 60
		x2 := ScreenWidth/2 - text.BoundString(g.miscFont, txt2).Max.X/2
		y2 := ScreenHeight/2 + 30
		text.Draw(screen, txt, g.titleFont, x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		text.Draw(screen, txt2, g.miscFont, x2, y2, color.White)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = GameStateMenu
		}
	}
}

func (g *Game) DrawPlaying(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-g.scrollOffset), 0)
	screen.DrawImage(g.background, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.PlayerScreenPos.X, g.PlayerScreenPos.Y)
	screen.DrawImage(g.player, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(doorPosition.X-float64(g.scrollOffset), doorPosition.Y)
	screen.DrawImage(g.door, op)
}

func (g *Game) UpdatePlaying() {
	g.HandleGameInput()
	g.MovePlayer()

	if rand.Intn(100) < 7 {
		enemy := Enemy{
			Position: Vector{X: float64(BackgroundWidth), Y: rand.Float64() * ScreenHeight},
			Velocity: enemySpeed + rand.Float64()*3,
		}
		enemies = append(enemies, enemy)
	}

	for i, e := range enemies {
		enemies[i].Position.X -= enemies[i].Velocity
		if Collide(g.PlayerPosition, e.Position) {
			//log.Println("check: collision - enemy")
			enemies = nil
			g.state = GameStateLose
			g.PlayerPosition = g.PlayerOrigin

			break
		}
	}

	if Collide(g.PlayerPosition, doorPosition) {
		//log.Println("check: collision - door")
		enemies = nil
		g.state = GameStateWin
		g.PlayerPosition = g.PlayerOrigin

		return
	}
}

func (g *Game) HandleGameInput() {
	continuousMovement = 0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		continuousMovement = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		continuousMovement = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.PlayJumpSound()
		g.PlayerVelocity.Y = -float64(jumpPower)
	}
}
