package pieces

import (
	"errors"
	"log"

	"github.com/Qinbeans/chess-htmx/utils"
	"github.com/gorilla/websocket"
)

// Square is the container for a square on the chess board
type Square struct {
	Color string
	Piece int
}

// SerSquare is the container for a square on the chess board for serialization
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

const (
	MAX_CLIENTS = 2
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

	PIECE_NAMES = map[int]string{
		PAWN + BLACK:   "B_PAWN",
		ROOK + BLACK:   "B_ROOK",
		KNIGHT + BLACK: "B_KNIGHT",
		BISHOP + BLACK: "B_BISHOP",
		QUEEN + BLACK:  "B_QUEEN",
		KING + BLACK:   "B_KING",
		PAWN:           "W_PAWN",
		ROOK:           "W_ROOK",
		KNIGHT:         "W_KNIGHT",
		BISHOP:         "W_BISHOP",
		QUEEN:          "W_QUEEN",
		KING:           "W_KING",
		NONE:           "<NONE>",
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

// Game is a struct for a game of chess
//   - Board: 8x8 array of Squares
//   - Clients: list of client ids
//   - Conn: websocket connection
type Game struct {
	Board        [8][8]Square
	Clients      map[string]*websocket.Conn
	ClientColors map[string]int
	Turn         int
}

// Creates a new game of chess
func NewGame(user1 string) *Game {
	return &Game{
		Board:        STARTING_POSITION,
		Clients:      map[string]*websocket.Conn{user1: nil},
		ClientColors: map[string]int{user1: WHITE},
		Turn:         WHITE,
	}
}

func (g *Game) ResetBoard() {
	g.Board = STARTING_POSITION
}

// toSquareArray converts the board to a 1D array of SerSquares
func (g *Game) toSquareArray() []SerSquare {
	var board []SerSquare
	// if len(g.Clients) < MAX_CLIENTS {
	// 	return board
	// }
	for _, row := range g.Board {
		for _, square := range row {
			board = append(board, SerSquare{Color: square.Color, Piece: PIECES[square.Piece]})
		}
	}
	return board
}

// movePiece moves a piece from one square to another
func (g *Game) movePiece(x1, y1, x2, y2 int) error {
	// Check target square
	piece := (g.Board[x1][y1].Piece | BLACK) - BLACK
	otherPiece := (g.Board[x2][y2].Piece | BLACK) - BLACK
	pieceColor := g.Board[x1][y1].Piece & BLACK
	otherPieceColor := g.Board[x2][y2].Piece & BLACK
	if !g.checkIsLegalMove(x1, y1, x2, y2) {
		return errors.New("illegal move")
	}
	if piece != NONE && otherPiece != NONE && pieceColor == otherPieceColor {
		return errors.New("can't move to a square with a piece of the same color")
	}
	g.Board[x2][y2].Piece, g.Board[x1][y1].Piece = g.Board[x1][y1].Piece, g.Board[x2][y2].Piece
	return nil
}

func (g *Game) takePiece(x1, y1, x2, y2 int) {
	g.Board[x2][y2].Piece = g.Board[x1][y1].Piece
	g.Board[x1][y1].Piece = NONE
}

func (g *Game) castle(x1, y1, x2, y2 int) (int, int, int, int, error) {
	if g.isCheck(x1, y1) || g.isCheck(x2, y2) {
		return -1, -1, -1, -1, errors.New("can't castle while in check")
	}
	// find out which piece is rook and which is king
	// (x1, y1) or (x2, y2) can be the king or the rook as the action is to drag one of the two pieces to the other
	var rookX, kingX, kingDY, rookDY int

	if y1 == 0 {
		// this is probably the rook
		rookX = x1
		kingX = x2
		kingDY = 1
		rookDY = 2
	} else if y1 == 7 {
		// this is probably the rook
		rookX = x1
		kingX = x2
		kingDY = 6
		rookDY = 5
	}

	if g.isCheck(kingX, kingDY) {
		return -1, -1, -1, -1, errors.New("can't castle into check")
	}

	g.Board[kingX][kingDY].Piece = g.Board[kingX][y1].Piece
	g.Board[rookX][rookDY].Piece = g.Board[rookX][y1].Piece
	g.Board[kingX][y1].Piece = NONE
	g.Board[rookX][y1].Piece = NONE
	return kingX, kingDY, rookX, rookDY, nil
}

func (g *Game) whereIsKing(color int) (int, int) {
	for i, row := range g.Board {
		for j, square := range row {
			if square.Piece == KING+color {
				return i, j
			}
		}
	}
	return -1, -1
}

// This function is used after we check legal moves
func (g *Game) isCastle(x1, y1, x2, y2 int) bool {
	piece := g.Board[x1][y1].Piece &^ BLACK
	otherPiece := g.Board[x2][y2].Piece &^ BLACK
	if piece == KING && otherPiece == ROOK {
		return true
	}
	return false
}

func (g *Game) checkPawnMovement(x1, y1, x2, y2 int) bool {
	piece := g.Board[x1][y1].Piece
	otherPiece := g.Board[x2][y2].Piece
	color := piece & BLACK
	if color == WHITE {
		if (x1 == 1 && x2 == 3 && y1 == y2) || (x1-x2 == -1 && y1 == y2) || (x1-x2 == -1 && utils.Abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
			if utils.Abs(y1-y2) == 2 {
				// check if any piece is in the way
				for i := 1; i < 3; i++ {
					if g.Board[x1-i][y1].Piece != NONE {
						return false
					}
				}
			}
			if otherPiece != NONE && otherPiece&BLACK == WHITE && y1 == y2 && x1-x2 == -1 {
				return false
			}
			return true
		}
	} else {
		if (x1 == 6 && x2 == 4 && y1 == y2) || (x1-x2 == 1 && y1 == y2) || (x1-x2 == 1 && utils.Abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
			if utils.Abs(y1-y2) == 2 {
				// check if any piece is in the way
				for i := 1; i < 3; i++ {
					if g.Board[x1+i][y1].Piece != NONE {
						return false
					}
				}
			}
			if otherPiece != NONE && otherPiece&BLACK == BLACK && y1 == y2 && x1-x2 == 1 {
				return false
			}
			return true
		}
	}
	return false
}

func (g *Game) checkRookMovement(x1, y1, x2, y2 int) bool {
	piece := g.Board[x1][y1].Piece
	// check if any piece is in the way, make sure to exclude the target square
	y_min := utils.Min(y1, y2)
	y_max := utils.Max(y1, y2)
	x_min := utils.Min(x1, x2)
	x_max := utils.Max(x1, x2)
	if x1 == x2 {
		for i := y_min + 1; i < y_max; i++ {
			if g.Board[x1][i].Piece != NONE {
				return false
			}
		}
	} else {
		for i := x_min + 1; i < x_max; i++ {
			if g.Board[i][y1].Piece != NONE {
				return false
			}
		}
	}
	if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) {
		return true
	}
	// castling check
	switch piece & BLACK {
	case WHITE:
		if (((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 6)) && g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE) || ((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 2)) && g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
			return true
		}
	case BLACK:
		if (((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 6)) && g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE) || ((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 2)) && g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
			return true
		}
	}
	return false
}

func (g *Game) checkKnightMovement(x1, y1, x2, y2 int) bool {
	x_comp := utils.Abs(x1 - x2)
	y_comp := utils.Abs(y1 - y2)
	if y_comp == 2 && x_comp == 1 || x_comp == 2 && y_comp == 1 {
		return true
	}
	return false
}

func (g *Game) checkBishopMovement(x1, y1, x2, y2 int) bool {
	// check if any piece is in the way, make sure to exclude the target square
	x_min := utils.Min(x1, x2)
	x_max := utils.Max(x1, x2)
	if x1 < x2 && y1 < y2 {
		for i := 1; i < x_max-x_min; i++ {
			if g.Board[x1+i][y1+i].Piece != NONE {
				return false
			}
		}
	} else if x1 < x2 && y1 > y2 {
		for i := 1; i < x_max-x_min; i++ {
			if g.Board[x1+i][y1-i].Piece != NONE {
				return false
			}
		}
	} else if x1 > x2 && y1 < y2 {
		for i := 1; i < x_max-x_min; i++ {
			if g.Board[x1-i][y1+i].Piece != NONE {
				return false
			}
		}
	} else {
		for i := 1; i < x_max-x_min; i++ {
			if g.Board[x1-i][y1-i].Piece != NONE {
				return false
			}
		}
	}
	if utils.Abs(x1-x2) == utils.Abs(y1-y2) {
		return true
	}
	return false
}

func (g *Game) checkKingMovement(x1, y1, x2, y2 int) bool {
	if utils.Abs(x1-x2) <= 1 && utils.Abs(y1-y2) <= 1 {
		return true
	}
	// castling check
	switch g.Board[x1][y1].Piece & BLACK {
	case WHITE:
		if (((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 6)) && g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE) || ((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 2)) && g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
			return true
		}
	case BLACK:
		if (((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 6)) && g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE) || ((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 2)) && g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
			return true
		}
	}
	return false
}

func (g *Game) isCheck(x1, y1 int) bool {
	// check if the king is in check
	kingColor := g.Board[x1][y1].Piece & BLACK
	for i, row := range g.Board {
		for j, square := range row {
			if square.Piece&BLACK != kingColor {
				if g.checkIsLegalMove(i, j, x1, y1) {
					return true
				}
			}
		}
	}
	return false
}

func (g *Game) isCheckmate(color int) bool {
	// check if the king is in check
	kingX, kingY := g.whereIsKing(color)
	if !g.isCheck(kingX, kingY) {
		return false
	}
	// check if the king can move
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i == 0 && j == 0 {
				continue
			}
			if kingX+i >= 0 && kingX+i < 8 && kingY+j >= 0 && kingY+j < 8 {
				if g.checkIsLegalMove(kingX, kingY, kingX+i, kingY+j) {
					return false
				}
			}
		}
	}
	// check if any piece can take the attacking piece
	for i, row := range g.Board {
		for j, square := range row {
			if square.Piece&BLACK == color {
				for k := 0; k < 8; k++ {
					for l := 0; l < 8; l++ {
						if g.checkIsLegalMove(i, j, k, l) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

// checkIsLegalMove checks if a move is legal
func (g *Game) checkIsLegalMove(x1, y1, x2, y2 int) bool {
	// get piece
	log.Printf("%s<%d, %d> -> %s<%d, %d> ", PIECE_NAMES[g.Board[x1][y1].Piece], x1, y1, PIECE_NAMES[g.Board[x2][y2].Piece], x2, y2)
	piece := g.Board[x1][y1].Piece
	switch piece &^ BLACK {
	case PAWN:
		return g.checkPawnMovement(x1, y1, x2, y2)
	case ROOK:
		return g.checkRookMovement(x1, y1, x2, y2)
	case KNIGHT:
		return g.checkKnightMovement(x1, y1, x2, y2)
	case BISHOP:
		return g.checkBishopMovement(x1, y1, x2, y2)
	case QUEEN:
		return g.checkRookMovement(x1, y1, x2, y2) || g.checkBishopMovement(x1, y1, x2, y2)
	case KING:
		return g.checkKingMovement(x1, y1, x2, y2)
	default:
		return false
	}
}
