package pieces

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/flosch/pongo2/v6"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	READSIZE  = 1024
	WRITESIZE = 1024
	ALL       = ""
)

type Server struct {
	Upgrader websocket.Upgrader
	Games    map[string]*Game
}

type Message struct {
	Author  string            `json:"author"`
	Content map[string]string `json:"content"`
}

// *****************************************************************************

// NewServer returns a new server
func NewServer() *Server {
	return &Server{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  READSIZE,
			WriteBufferSize: WRITESIZE,
		},
		Games: make(map[string]*Game),
	}
}

// Subscribe returns a unique id for the client to use to subscribe to the websocket
func (g *Server) SubscribeNewUser(room string) uuid.UUID {
	id := uuid.New()
	if g.Games[room] == nil {
		g.Games[room] = NewGame(id.String())
	} else {
		g.Games[room].Clients[id.String()] = nil
		g.Games[room].ClientColors[id.String()] = BLACK
	}
	return id
}

// Unsubscribe removes the client from the list of connections
func (g *Server) Unsubscribe(id uuid.UUID, room string) {
	delete(g.Games[room].Clients, id.String())
}

// GracefulDisconnect removes the client from the list of connections and broadcasts the intent to disconnect
func (g *Server) GracefulDisconnect(id uuid.UUID, room string) error {
	var err error
	var intnt []byte
	if g.Games[room] == nil {
		return nil
	}
	if len(g.Games[room].Clients) > 0 {
		// broadcast intent to disconnect
		intnt, err = json.Marshal(Message{
			Author: id.String(),
			Content: map[string]string{
				"type": "cmd",
				"msg":  "disconnected",
			},
		})
		g.Broadcast(id.String(), room, intnt)
	}
	if g.Games[room].Clients[id.String()] != nil {
		g.Games[room].Clients[id.String()].Close()
	}
	g.Unsubscribe(id, room)
	return err
}

// Broadcast sends a message to all clients in a room; empty user means broadcast to all
func (g *Server) Broadcast(user, room string, message []byte) {
	game := g.Games[room]
	for id, conn := range game.Clients {
		if id != user && conn != nil {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Close closes the server
func (g *Server) Close() {
	for _, game := range g.Games {
		for id, conn := range game.Clients {
			if conn != nil {
				conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				conn.Close()
			}
			delete(game.Clients, id)
		}
	}
	for room := range g.Games {
		delete(g.Games, room)
	}
}

// SendError sends an error message to a client
func (g *Server) SendError(user, room string, err error) {
	errorMsg, _ := json.Marshal(Message{
		Author: user,
		Content: map[string]string{
			"type": "error",
			"msg":  err.Error(),
		},
	})
	g.Games[room].Clients[user].WriteMessage(websocket.TextMessage, errorMsg)
}

func (g *Server) SendErrorBytes(user, room string, err []byte) {
	g.Games[room].Clients[user].WriteMessage(websocket.TextMessage, err)
}

// SendMoveError sends a move error message to a client
func (g *Server) SendMoveError(user, room string, err error, src, dst int) {
	x1, y1 := src/8, src%8
	x2, y2 := dst/8, dst%8
	errorMsg, _ := json.Marshal(Message{
		Author: user,
		Content: map[string]string{
			"type":      "error",
			"msg":       err.Error(),
			"src":       fmt.Sprintf("%d", src),
			"src_color": g.Games[room].Board[x1][y1].Color,
			"src_piece": PIECES[g.Games[room].Board[x1][y1].Piece],
			"dst":       fmt.Sprintf("%d", dst),
			"dst_color": g.Games[room].Board[x2][y2].Color,
			"dst_piece": PIECES[g.Games[room].Board[x2][y2].Piece],
		},
	})
	g.SendErrorBytes(user, room, errorMsg)
}

func (g *Server) SendTakeAck(user, room string, src, dst int) {
	moveMsg, _ := json.Marshal(Message{
		Author: user,
		Content: map[string]string{
			"type": "take-ack",
			"src":  fmt.Sprintf("%d", src),
			"dst":  fmt.Sprintf("%d", dst),
		},
	})
	g.Games[room].Clients[user].WriteMessage(websocket.TextMessage, moveMsg)
}

func (g *Server) SendCastle(user, room string, k_src, r_src, k_dst, r_dst int) {
	moveMsg, _ := json.Marshal(Message{
		Author: user,
		Content: map[string]string{
			"type":  "castle",
			"k_src": fmt.Sprintf("%d", k_src),
			"r_src": fmt.Sprintf("%d", r_src),
			"k_dst": fmt.Sprintf("%d", k_dst),
			"r_dst": fmt.Sprintf("%d", r_dst),
		},
	})
	g.Broadcast(ALL, room, moveMsg)
}

// *****************************************************************************

// NewGame is a callback for creating a new game of chess
func (g *Server) NewGame(c echo.Context) error {
	room := uuid.New().String()
	client := uuid.New().String()
	g.Games[room] = NewGame(client)
	return c.JSON(200, map[string]string{
		"room": room,
		"id":   client,
		"type": "chess",
	})
}

// ConnectToRoom is a callback for connecting to a room
func (g *Server) ConnectToRoom(c echo.Context) error {
	room_id := c.FormValue("room")
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
	client := g.SubscribeNewUser(room_id)
	return c.JSON(200, map[string]string{
		"room": room_id,
		"id":   client.String(),
		"type": "chess",
	})
}

// Room is a callback for rendering the chess room
func (g *Server) Room(c echo.Context) error {
	params := c.QueryParams()
	room := params.Get("room")
	if room == "" {
		log.Print("Room parameter is required")
		return c.Redirect(302, "/")
	}
	client := params.Get("user")
	if client == "" {
		log.Print("User parameter is required")
		return c.Redirect(302, "/")
	}
	if _, ok := g.Games[room]; !ok {
		log.Print("Room does not exist")
		return c.Redirect(302, "/")
	}
	return c.Render(200, "chess.dj", pongo2.Context{
		"title":       "Let's play chess",
		"description": "Play chess with a friend",
		"room":        room,
		"client":      client,
		"board":       g.Games[room].toSquareArray(),
	})
}

func (g *Server) WSHandler(c echo.Context) error {
	if c.Request().Header.Get("Connection") == "Upgrade" {
		params := c.QueryParams()
		room := params.Get("room")
		if room == "" {
			log.Println("Room parameter is required")
			return c.JSON(400, map[string]string{
				"error": "room parameter is required",
			})
		}
		user := params.Get("user")
		if user == "" {
			log.Println("User parameter is required")
			return c.JSON(400, map[string]string{
				"error": "user parameter is required",
			})
		}
		if g.Games[room] == nil {
			log.Println("Room does not exist")
			return c.JSON(400, map[string]string{
				"error": "room does not exist",
			})
		}
		check := false
		for id := range g.Games[room].Clients {
			if id == user {
				check = true
			}
		}
		if !check {
			log.Println("User is not in the room")
			return c.JSON(400, map[string]string{
				"error": "user is not in the room",
			})
		}
		conn, err := g.Upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Println(err)
			return err
		}
		g.Games[room].Clients[user] = conn
		go g.handleConnection(conn, user, room)
		log.Printf("User %s connected to room %s\n", user, room)
		return nil
	}
	return c.JSON(400, map[string]string{
		"error": "invalid request",
	})
}

func (g *Server) handleConnection(conn *websocket.Conn, user, room string) {
	defer g.GracefulDisconnect(uuid.MustParse(user), room)
	joinMsg, err := json.Marshal(Message{
		Author: user,
		Content: map[string]string{
			"type": "cmd",
			"msg":  "connected",
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
	g.Broadcast(user, room, joinMsg)
	for {
		_, raw_message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		var message map[string]interface{}
		err = json.Unmarshal(raw_message, &message)
		if err != nil {
			log.Println(err)
			continue
		}
		switch message["type"].(string) {
		case "cmd":
			switch message["msg"].(string) {
			case "quit":
				log.Println("User quit")
				return
			case "acknowledge":
				log.Println("User acknowledged")
				ackMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type": "cmd",
						"msg":  "acknowledge",
					},
				})
				g.Broadcast(user, room, ackMsg)
			case "reset-req":
				log.Println("User requested reset")
				if len(g.Games[room].Clients) < 2 {
					resetMsg, _ := json.Marshal(Message{
						Author: user,
						Content: map[string]string{
							"type": "error",
							"msg":  "reset denied, not enough players",
						},
					})
					g.SendErrorBytes(user, room, resetMsg)
				}
				resetMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type": "cmd",
						"msg":  "reset-req",
					},
				})
				g.Broadcast(user, room, resetMsg)
			case "reset-ack":
				log.Println("User acknowledged reset")
				g.Games[room].ResetBoard()
				board, _ := json.Marshal(g.Games[room].toSquareArray())
				resetMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type":  "cmd",
						"msg":   "reset-ack",
						"board": string(board),
					},
				})
				g.Broadcast(user, room, resetMsg)
			default:
				log.Println(message["msg"])
			}
		case "move":
			src_pos := int(message["from"].(float64))
			dst_pos := int(message["to"].(float64))
			x1, y1 := src_pos/8, src_pos%8
			x2, y2 := dst_pos/8, dst_pos%8
			// Check if src is client's color
			if g.Games[room].Board[x1][y1].Piece&BLACK != g.Games[room].ClientColors[user] {
				g.SendMoveError(user, room, fmt.Errorf("not your piece"), src_pos, dst_pos)
				continue
			} else if g.Games[room].Turn != g.Games[room].ClientColors[user] {
				g.SendMoveError(user, room, fmt.Errorf("not your turn"), src_pos, dst_pos)
				continue
			}
			err = g.Games[room].movePiece(x1, y1, x2, y2)
			if err != nil {
				g.SendMoveError(user, room, err, src_pos, dst_pos)
				continue
			}
			if g.Games[room].Board[x1][y1].Piece == KING && g.Games[room].isCheck(x2, y2) {
				g.SendMoveError(user, room, fmt.Errorf("cannot move into check"), src_pos, dst_pos)
				continue
			}
			if g.Games[room].Board[x1][y1].Piece != KING {
				// check if king is in check
				src_color := g.Games[room].Board[x1][y1].Piece &^ BLACK
				kingX, kingY := g.Games[room].whereIsKing(src_color)
				if g.Games[room].isCheck(kingX, kingY) {
					g.SendMoveError(user, room, fmt.Errorf("king is in check"), src_pos, dst_pos)
					continue
				}
			}
			if g.Games[room].isCastle(x1, y1, x2, y2) {
				var oKingX, oKingY, oRookX, oRookY int
				if g.Games[room].Board[x1][y1].Piece != KING {
					oKingX, oKingY = x2, y2
					oRookX, oRookY = x1, y1
				} else {
					oKingX, oKingY = x1, y1
					oRookX, oRookY = x2, y2
				}
				kingX, kingY, rookX, rookY, err := g.Games[room].castle(x1, y1, x2, y2)
				if err != nil {
					g.SendMoveError(user, room, err, src_pos, dst_pos)
					continue
				}
				o_king_pos := oKingX*8 + oKingY
				o_rook_pos := oRookX*8 + oRookY
				king_pos := kingX*8 + kingY
				rook_pos := rookX*8 + rookY
				g.SendCastle(user, room, o_king_pos, o_rook_pos, king_pos, rook_pos)
			}
			taken := false
			// check if piece exists at dst
			if g.Games[room].Board[x2][y2].Piece != NONE {
				g.Games[room].takePiece(x1, y1, x2, y2)
				taken = true
			}
			moveMsg, _ := json.Marshal(Message{
				Author: user,
				Content: map[string]string{
					"type":  "move",
					"src":   fmt.Sprintf("%d", src_pos),
					"dst":   fmt.Sprintf("%d", dst_pos),
					"taken": fmt.Sprintf("%t", taken),
				},
			})
			// check if opponent is in checkmate
			color := g.Games[room].Board[x2][y2].Piece &^ BLACK
			opp_color := color ^ BLACK
			g.Broadcast(user, room, moveMsg)
			if g.Games[room].isCheckmate(opp_color) {
				moveMsg, _ = json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type":  "checkmate",
						"color": g.Games[room].Board[x2][y2].Color,
					},
				})
				g.Broadcast(ALL, room, moveMsg)
			}
			switch g.Games[room].Turn {
			case WHITE:
				g.Games[room].Turn = BLACK
			case BLACK:
				g.Games[room].Turn = WHITE
			}
		default:
			log.Println("Unknown message type")
		}
	}
}
