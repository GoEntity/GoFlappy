package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	EnemySizeX      = 49.0
	EnemySizeY      = 29.0
)

var (
	enemies            []Enemy
	enemySpeed         = 5.0
)

type Enemy struct {
	Position Vector
	Velocity float64
	dirty    bool
}

func (g *Game) DrawEnemies(screen *ebiten.Image) {
	for _, e := range enemies {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.Position.X-float64(g.scrollOffset), e.Position.Y)
		screen.DrawImage(g.enemy, op)

		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset)), int(e.Position.Y))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset) + 49), int(e.Position.Y))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset)), int(e.Position.Y + 29))
		// ebitenutil.DebugPrintAt(screen, "**", int(e.Position.X-float64(scrollOffset) + 49), int(e.Position.Y + 29))
	}
}

func RemoveEnemies() {
	for i := range enemies {
		if enemies[i].Position.X < 0 {
			enemies[i].dirty = true
		}
	}
	enemies = EnemyRecycler(enemies)
}

func EnemyRecycler(enemies []Enemy) []Enemy {
	recycled := enemies[:0]
	for _, e := range enemies {
		if !e.dirty {
			recycled = append(recycled, e)
		}
	}
	return recycled
}
