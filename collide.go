package main

func Collide(a, b Vector) bool {
	return !(a.X > b.X+EnemySizeX ||
		a.X+PlayerSizeX < b.X ||
		a.Y > b.Y+EnemySizeY ||
		a.Y+PlayerSizeY < b.Y)
}