package player

type Movements struct {
	Up    bool
	Down  bool
	Left  bool
	Right bool
}

type Player struct {
	Id        string
	X         float64
	Y         float64
	Speed     float64
	Movements Movements
}
