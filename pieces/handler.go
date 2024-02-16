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
)

type Server struct {
	Upgrader websocket.Upgrader
	Games    map[string]*Game
	Mode     string
}

type Message struct {
	Author  string            `json:"author"`
	Content map[string]string `json:"content"`
}

// *****************************************************************************

// NewServer returns a new server
func NewServer(mode string) *Server {
	return &Server{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  READSIZE,
			WriteBufferSize: WRITESIZE,
		},
		Games: make(map[string]*Game),
		Mode:  mode,
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

// Broadcast sends a message to all clients in a room
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
	protoc := "ws"
	if g.Mode == "release" {
		protoc = "wss"
	}
	return c.Render(200, "chess.dj", pongo2.Context{
		"protoc": protoc,
		"room":   room,
		"client": client,
		"board":  g.Games[room].toSquareArray(),
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
				errorMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type":      "error",
						"msg":       "not your piece",
						"src":       fmt.Sprintf("%d", src_pos),
						"src_color": g.Games[room].Board[x1][y1].Color,
						"src_piece": PIECES[g.Games[room].Board[x1][y1].Piece],
						"dst":       fmt.Sprintf("%d", dst_pos),
						"dst_color": g.Games[room].Board[x2][y2].Color,
						"dst_piece": PIECES[g.Games[room].Board[x2][y2].Piece],
					},
				})
				g.SendErrorBytes(user, room, errorMsg)
				continue
			} else if g.Games[room].Turn != g.Games[room].ClientColors[user] {
				errorMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type":      "error",
						"msg":       "not your turn",
						"src":       fmt.Sprintf("%d", src_pos),
						"src_color": g.Games[room].Board[x1][y1].Color,
						"src_piece": PIECES[g.Games[room].Board[x1][y1].Piece],
						"dst":       fmt.Sprintf("%d", dst_pos),
						"dst_color": g.Games[room].Board[x2][y2].Color,
						"dst_piece": PIECES[g.Games[room].Board[x2][y2].Piece],
					},
				})
				g.SendErrorBytes(user, room, errorMsg)
				continue
			}
			err = g.Games[room].movePiece(x1, y1, x2, y2)
			if err != nil {
				errorMsg, _ := json.Marshal(Message{
					Author: user,
					Content: map[string]string{
						"type":      "error",
						"msg":       err.Error(),
						"src":       fmt.Sprintf("%d", src_pos),
						"src_color": g.Games[room].Board[x1][y1].Color,
						"src_piece": PIECES[g.Games[room].Board[x1][y1].Piece],
						"dst":       fmt.Sprintf("%d", dst_pos),
						"dst_color": g.Games[room].Board[x2][y2].Color,
						"dst_piece": PIECES[g.Games[room].Board[x2][y2].Piece],
					},
				})
				g.SendErrorBytes(user, room, errorMsg)
				continue
			}
			moveMsg, _ := json.Marshal(Message{
				Author: user,
				Content: map[string]string{
					"type": "move",
					"src":  fmt.Sprintf("%d", src_pos),
					"dst":  fmt.Sprintf("%d", dst_pos),
				},
			})
			g.Broadcast(user, room, moveMsg)
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
