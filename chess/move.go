package chess

type (
	Move struct {
		from, to int
	}
	Promotion struct {
		Move
		newPiece string
	}
)

func newMove(from, to int) Move {
	return Move{from: from, to: to}

}
func newPromotion(from, to int, newPiece string) Promotion {
	return Promotion{
		newPiece: newPiece,
		Move: Move{
			from: from,
			to:   to,
		},
	}
}

func (m Move) From() int {
	return m.from
}
func (m Move) To() int {
	return m.from
}
