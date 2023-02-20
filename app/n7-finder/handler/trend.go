package handler

type Direction int

const (
	UP Direction = iota
	DOWN
	HORIZONTAL
)

type Trend interface {
	Observe() Direction
}
