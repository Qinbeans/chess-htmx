package pieces

import (
	"errors"
	"fmt"
	"log"

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

type Games struct {
	Games map[string]*Game
}

func NewGames() *Games {
	return &Games{
		Games: make(map[string]*Game),
	}
}

type Game struct {
	Board   [8][8]Square
	Clients []string
}

func NewGame(user1 string) *Game {
	return &Game{
		Board:   STARTING_POSITION,
		Clients: []string{user1},
	}
}

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

func abs(x int) int {
	return x &^ (x >> 31)
}

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
			if (x1 == 1 && x2 == 3 && y1 == y2) || (x1-x2 == -1 && y1 == y2) || (x1-x2 == -1 && abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
				return true
			}
		} else {
			fmt.Println("BLACK")
			if (x1 == 6 && x2 == 4 && y1 == y2) || (x1-x2 == 1 && y1 == y2) || (x1-x2 == 1 && abs(y1-y2) == 1 && otherPiece != NONE && otherPiece&BLACK == WHITE) {
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
			if (abs(x1-x2) == 2 && abs(y1-y2) == 1) || (abs(x1-x2) == 1 && abs(y1-y2) == 2) {
				return true
			}
		case BISHOP:
			log.Println("BISHOP")
			if abs(x1-x2) == abs(y1-y2) {
				return true
			}
		case QUEEN:
			log.Println("QUEEN")
			if (x1 == x2 && y1 != y2) || (x1 != x2 && y1 == y2) || abs(x1-x2) == abs(y1-y2) {
				return true
			}
		case KING:
			log.Print("KING ")
			if abs(x1-x2) <= 1 && abs(y1-y2) <= 1 {
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

func (g *Games) NewGame(c echo.Context) error {
	room := uuid.New().String()
	client := uuid.New().String()
	g.Games[room] = NewGame(client)
	return c.JSON(200, map[string]string{
		"room": room,
		"id":   client,
		"type": "chess",
	})
}

func (g *Games) ConnectToRoom(c echo.Context) error {
	var room map[string]string
	err := c.Bind(&room)
	if err != nil {
		log.Println(err)
		return c.JSON(200, map[string]string{
			"message": "bad request",
			"type":    "chess",
		})
	}
	room_id := room["room"]
	client := room["client"]
	if _, ok := g.Games[room_id]; !ok {
		return c.JSON(200, map[string]string{
			"message": "room not found",
			"type":    "chess",
		})
	}
	if len(g.Games[room_id].Clients) >= MAX_CLIENTS {
		return c.JSON(200, map[string]string{
			"message": "room full",
			"type":    "chess",
		})
	}
	g.Games[room_id].Clients = append(g.Games[room_id].Clients, client)
	return c.JSON(200, map[string]string{
		"room": room_id,
		"id":   client,
		"type": "chess",
	})
}

func (g *Games) Room(c echo.Context) error {
	params := c.QueryParams()
	room := params.Get("room")
	if room == "" {
		return c.Redirect(302, "/")
	}
	client := params.Get("user")
	if client == "" {
		return c.Redirect(302, "/")
	}
	if _, ok := g.Games[room]; !ok {
		return c.Redirect(302, "/")
	}
	return c.Render(200, "chess.dj", pongo2.Context{
		"room":   room,
		"client": client,
		"board":  g.Games[room].toSquareArray(),
	})
}

func (g *Games) MovePiece(c echo.Context) error {
	var move map[string]interface{}
	err := c.Bind(&move)
	if err != nil {
		log.Println(err)
		return c.JSON(200, []byte("{\"message\": \"bad request\"}"))
	}
	room := move["room"].(string)
	//client := move["client"].(string)
	src_pos := int(move["src_pos"].(float64))
	trg_pos := int(move["trg_pos"].(float64))
	log.Printf("src_pos: %d, trg_pos: %d\n", src_pos, trg_pos)
	x1, y1 := src_pos/8, src_pos%8
	x2, y2 := trg_pos/8, trg_pos%8
	if _, ok := g.Games[room]; !ok {
		return c.JSON(200, map[string]string{
			"message": "room not found",
		})
	}
	err = g.Games[room].movePiece(x1, y1, x2, y2)
	if err != nil {
		log.Println(err)
		return c.JSON(200, map[string]string{
			"message": "illegal move",
		})
	}
	return c.JSON(200, map[string]string{
		"message": "ok",
	})
}
