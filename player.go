package main

const (
	PlayerSizeX     = 50.0
	PlayerSizeY     = 68.0
)

var (
	continuousMovement = 0
	gravity            = 0.5
	jumpPower          = 5
	speed              = 2.5
	maxSpeed           = 5
)

func (g *Game) MovePlayer() {
	g.PlayerVelocity.Y += float64(gravity)

	if g.PlayerVelocity.Y > float64(maxSpeed) {
		g.PlayerVelocity.Y = float64(maxSpeed)
	}

	g.PlayerVelocity.X = float64(continuousMovement) * float64(speed)

	forPlayerPosition := Vector{
		X: g.PlayerPosition.X + g.PlayerVelocity.X,
		Y: g.PlayerPosition.Y + g.PlayerVelocity.Y,
	}
	forPlayerScreenPos := Vector{
		X: forPlayerPosition.X - float64(g.scrollOffset),
		Y: g.PlayerScreenPos.Y,
	}

	if forPlayerScreenPos.X < 0 || forPlayerScreenPos.X > ScreenWidth {
		g.PlayerVelocity.X = -2 * g.PlayerVelocity.X
	} else {
		g.PlayerPosition.X = forPlayerPosition.X
		g.PlayerScreenPos.X = forPlayerScreenPos.X
	}

	forPlayerScreenPos = Vector{X: g.PlayerScreenPos.X, Y: forPlayerPosition.Y}
	if forPlayerScreenPos.Y < 0 || forPlayerScreenPos.Y > ScreenHeight-40 {
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
	} else if g.PlayerScreenPos.X > ScreenWidth {
		if g.PlayerVelocity.X > 0 {
			g.PlayerVelocity.X = -2 * g.PlayerVelocity.X
		}
	}

	if g.PlayerScreenPos.Y < 0 {
		if g.PlayerVelocity.Y < 0 {
			g.PlayerVelocity.Y = -2 * g.PlayerVelocity.Y
		}
	} else if g.PlayerScreenPos.Y > ScreenHeight-40 {
		if g.PlayerVelocity.Y > 0 {
			g.PlayerVelocity.Y = -2 * g.PlayerVelocity.Y
		}
	}

	g.scrollOffset = int(g.PlayerPosition.X) - ScreenWidth/2

	if g.scrollOffset < 0 {
		g.scrollOffset = 0
	} else if g.scrollOffset > BackgroundWidth-ScreenWidth {
		g.scrollOffset = BackgroundWidth - ScreenWidth
	}
}