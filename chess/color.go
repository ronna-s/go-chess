package chess

type Color bool

const (
	White = Color(true)
	Black = Color(false)
)

func (c Color) String() string {
	if c {
		return "white"
	}
	return "black"
}
