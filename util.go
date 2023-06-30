package main

func (g *Game) PlayJumpSound() {
	g.jumpPlayer.Play()
	g.jumpPlayer.Rewind()
}