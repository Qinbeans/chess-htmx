package pieces

import (
	"errors"

	"github.com/flosch/pongo2/v6"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Square struct {
	Color string
	Piece int
}

type SerSquare struct {
	Color string `json:"color"`
	Piece string `json:"piece"`
}

const (
	NONE   = 0b0000
	PAWN   = 0b0001
	ROOK   = 0b0010
	KNIGHT = 0b0011
	BISHOP = 0b0100
	QUEEN  = 0b0101
	KING   = 0b0110
)

const (
	WHITE = 0b0000
	BLACK = 0b1000
)

var (
	PIECES = map[int]string{
		PAWN + BLACK:   "c/c7/Chess_pdt45.svg",
		ROOK + BLACK:   "f/ff/Chess_rdt45.svg",
		KNIGHT + BLACK: "e/ef/Chess_ndt45.svg",
		BISHOP + BLACK: "9/98/Chess_bdt45.svg",
		QUEEN + BLACK:  "4/47/Chess_qdt45.svg",
		KING + BLACK:   "f/f0/Chess_kdt45.svg",
		PAWN:           "4/45/Chess_plt45.svg",
		ROOK:           "7/72/Chess_rlt45.svg",
		KNIGHT:         "7/70/Chess_nlt45.svg",
		BISHOP:         "b/b1/Chess_blt45.svg",
		QUEEN:          "1/15/Chess_qlt45.svg",
		KING:           "4/42/Chess_klt45.svg",
	}

	STARTING_POSITION = [8][8]Square{
		{{Color: "white/35", Piece: ROOK}, {Color: "white/15", Piece: KNIGHT}, {Color: "white/35", Piece: BISHOP}, {Color: "white/15", Piece: QUEEN}, {Color: "white/35", Piece: KING}, {Color: "white/15", Piece: BISHOP}, {Color: "white/35", Piece: KNIGHT}, {Color: "white/15", Piece: ROOK}},
		{{Color: "white/15", Piece: PAWN}, {Color: "white/35", Piece: PAWN}, {Color: "white/15", Piece: PAWN}, {Color: "white/35", Piece: PAWN}, {Color: "white/15", Piece: PAWN}, {Color: "white/35", Piece: PAWN}, {Color: "white/15", Piece: PAWN}, {Color: "white/35", Piece: PAWN}},
		{{Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}},
		{{Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}},
		{{Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}},
		{{Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}, {Color: "white/15", Piece: NONE}, {Color: "white/35", Piece: NONE}},
		{{Color: "white/35", Piece: PAWN + BLACK}, {Color: "white/15", Piece: PAWN + BLACK}, {Color: "white/35", Piece: PAWN + BLACK}, {Color: "white/15", Piece: PAWN + BLACK}, {Color: "white/35", Piece: PAWN + BLACK}, {Color: "white/15", Piece: PAWN + BLACK}, {Color: "white/35", Piece: PAWN + BLACK}, {Color: "white/15", Piece: PAWN + BLACK}},
		{{Color: "white/15", Piece: ROOK + BLACK}, {Color: "white/35", Piece: KNIGHT + BLACK}, {Color: "white/15", Piece: BISHOP + BLACK}, {Color: "white/35", Piece: QUEEN + BLACK}, {Color: "white/15", Piece: KING + BLACK}, {Color: "white/35", Piece: BISHOP + BLACK}, {Color: "white/15", Piece: KNIGHT + BLACK}, {Color: "white/35", Piece: ROOK + BLACK}},
	}
)

type Games struct {
	Games map[string]*Game
}

func NewGames() *Games {
	return &Games{
		Games: make(map[string]*Game),
	}
}

func (g *Games) NewGame(c echo.Context) error {
	room := uuid.New().String()
	g.Games[room] = NewGame()
	ctx := pongo2.Context{
		"room":    room,
		"squares": g.Games[room].toSquareArray(),
	}
	return c.Render(200, "chess.dj", ctx)
}

type Game struct {
	Board [8][8]Square
}

func NewGame() *Game {
	return &Game{
		Board: STARTING_POSITION,
	}
}

func (g *Game) toSquareArray() []SerSquare {
	var board []SerSquare
	for _, row := range g.Board {
		for _, square := range row {
			board = append(board, SerSquare{Color: square.Color, Piece: PIECES[square.Piece]})
		}
	}
	return board
}

func (g *Game) movePiece(x1, y1, x2, y2 int) error {
	// Check target square
	if !g.checkIsLegalMove(x1, y1, x2, y2) {
		return errors.New("illegal move")
	}
	if g.Board[x2][y2].Piece != NONE && g.Board[x1][y1].Color == g.Board[x2][y2].Color {
		return errors.New("can't move to a square with a piece of the same color")
	}
	g.Board[x2][y2].Piece = g.Board[x1][y1].Piece
	return nil
}

func abs(x int) int {
	return x &^ (x >> 31)
}

func (g *Game) checkIsLegalMove(x1, y1, x2, y2 int) bool {
	// get piece
	piece := g.Board[x1][y1].Piece
	switch piece {
	case PAWN:
		color := piece & BLACK
		if color == WHITE {
			if x1 == 6 && x2 == 4 && y1 == y2 {
				return true
			}
			if x1-x2 == 1 && y1 == y2 {
				return true
			}
		} else {
			if x1 == 1 && x2 == 3 && y1 == y2 {
				return true
			}
			if x1-x2 == -1 && y1 == y2 {
				return true
			}
		}
	default:
		switch piece | BLACK {
		case ROOK:
			if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) {
				return true
			}
			// castling check
			switch piece & BLACK {
			case WHITE:
				if x1 == 7 && y1 == 4 && x2 == 7 && y2 == 6 {
					if g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE {
						return true
					}
				}
				if x1 == 7 && y1 == 4 && x2 == 7 && y2 == 2 {
					if g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
						return true
					}
				}
			case BLACK:
				if x1 == 0 && y1 == 4 && x2 == 0 && y2 == 6 {
					if g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE {
						return true
					}
				}
				if x1 == 0 && y1 == 4 && x2 == 0 && y2 == 2 {
					if g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
						return true
					}
				}
			}
		case KNIGHT:
			if (abs(x1-x2) == 2 && abs(y1-y2) == 1) || (abs(x1-x2) == 1 && abs(y1-y2) == 2) {
				return true
			}
		case BISHOP:
			if abs(x1-x2) == abs(y1-y2) {
				return true
			}
		case QUEEN:
			if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) || abs(x1-x2) == abs(y1-y2) {
				return true
			}
		case KING:
			if abs(x1-x2) <= 1 && abs(y1-y2) <= 1 {
				return true
			}
			// castling check
			switch piece & BLACK {
			case WHITE:
				if x1 == 7 && y1 == 4 && x2 == 7 && y2 == 6 {
					if g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE {
						return true
					}
				}
				if x1 == 7 && y1 == 4 && x2 == 7 && y2 == 2 {
					if g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
						return true
					}
				}
			case BLACK:
				if x1 == 0 && y1 == 4 && x2 == 0 && y2 == 6 {
					if g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE {
						return true
					}
				}
				if x1 == 0 && y1 == 4 && x2 == 0 && y2 == 2 {
					if g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
						return true
					}
				}
			}
		default:
			return false
		}
	}
	return false
}
