package pieces

import (
	"errors"
	"fmt"
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

// checkIsLegalMove checks if a move is legal
func (g *Game) checkIsLegalMove(x1, y1, x2, y2 int) bool {
	// get piece
	log.Printf("<%d, %d> -> <%d, %d> ", x1, y1, x2, y2)
	piece := g.Board[x1][y1].Piece
	otherPiece := g.Board[x2][y2].Piece
	switch piece {
	case PAWN:
		log.Print("PAWN ")
		color := piece & BLACK
		if color == WHITE {
			fmt.Println("WHITE")
			if (x1 == 1 && x2 == 3 && y1 == y2) || (x1-x2 == -1 && y1 == y2) || (x1-x2 == -1 && utils.Abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
				return true
			}
		} else {
			fmt.Println("BLACK")
			if (x1 == 6 && x2 == 4 && y1 == y2) || (x1-x2 == 1 && y1 == y2) || (x1-x2 == 1 && utils.Abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
				return true
			}
		}
	default:
		switch piece &^ BLACK {
		case ROOK:
			log.Print("ROOK ")
			if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) {
				return true
			}
			// castling check
			switch piece & BLACK {
			case WHITE:
				fmt.Println("WHITE")
				if (((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 6)) && g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE) || ((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 2)) && g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
					return true
				}
			case BLACK:
				fmt.Println("BLACK")
				if (((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 6)) && g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE) || ((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 2)) && g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
					return true
				}
			}
		case KNIGHT:
			log.Println("KNIGHT")
			if (utils.Abs(x1-x2) == 2 && utils.Abs(y1-y2) == 1) || (utils.Abs(x1-x2) == 1 && utils.Abs(y1-y2) == 2) {
				return true
			}
		case BISHOP:
			log.Println("BISHOP")
			if utils.Abs(x1-x2) == utils.Abs(y1-y2) {
				return true
			}
		case QUEEN:
			log.Println("QUEEN")
			if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) || utils.Abs(x1-x2) == utils.Abs(y1-y2) {
				return true
			}
		case KING:
			log.Print("KING ")
			if utils.Abs(x1-x2) <= 1 && utils.Abs(y1-y2) <= 1 {
				return true
			}
			// castling check
			switch piece & BLACK {
			case WHITE:
				fmt.Println("WHITE")
				if (((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 6)) && g.Board[7][7].Piece == ROOK && g.Board[7][5].Piece == NONE) || ((x1 == 7 && y1 == 4) || (x2 == 7 && y2 == 2)) && g.Board[7][0].Piece == ROOK && g.Board[7][3].Piece == NONE {
					return true
				}
			case BLACK:
				fmt.Println("BLACK")
				if (((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 6)) && g.Board[0][7].Piece == ROOK && g.Board[0][5].Piece == NONE) || ((x1 == 0 && y1 == 4) || (x2 == 0 && y2 == 2)) && g.Board[0][0].Piece == ROOK && g.Board[0][3].Piece == NONE {
					return true
				}
			}
		default:
			return false
		}
	}
	return false
}
