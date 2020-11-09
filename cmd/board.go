package main

// Board struct of board.
type Board struct {
	x, y   int
	w, h   int
	dx, dy int
}

// X get x value
func (b *Board) X() int {
	return b.x
}

// Y get y value
func (b *Board) Y() int {
	return b.y
}

// W get w value
func (b *Board) W() int {
	return b.w
}

// H get h value
func (b *Board) H() int {
	return b.h
}

// Motion is board motion
func (b *Board) Motion() {
	b.x -= b.dx
	b.y -= b.dy

	if b.x < -b.w {
		b.x = int(rc.Right)
	}
}
