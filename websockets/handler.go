package websockets

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	READSIZE  = 1024
	WRITESIZE = 1024
)

type WSServer struct {
	Upgrader    websocket.Upgrader
	Connections map[string]*websocket.Conn
	Rooms       map[string][]string
}

type Message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

type Room struct {
	Members []string
	Lock    sync.Mutex
}

func NewWSServer() *WSServer {
	return &WSServer{
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  READSIZE,
			WriteBufferSize: WRITESIZE,
		},
		Connections: make(map[string]*websocket.Conn),
		Rooms:       make(map[string][]string),
	}
}

// Subscribe returns a unique id for the client to use to subscribe to the websocket
func (ws *WSServer) SubscribeNewUser(room string) uuid.UUID {
	id := uuid.New()
	if ws.Rooms[room] == nil {
		ws.Rooms[room] = []string{id.String()}
	} else {
		ws.Rooms[room] = append(ws.Rooms[room], id.String())
	}
	ws.Connections[id.String()] = nil
	return id
}

// Unsubscribe removes the client from the list of connections
func (ws *WSServer) Unsubscribe(id uuid.UUID, room string) {
	delete(ws.Connections, id.String())
	// remove user from room
	for i, v := range ws.Rooms[room] {
		if v == id.String() {
			ws.Rooms[room] = append(ws.Rooms[room][:i], ws.Rooms[room][i+1:]...)
		}
	}
}

func (ws *WSServer) GracefulDisconnect(room, user string, conn *websocket.Conn) {
	conn.Close()
	ws.Unsubscribe(uuid.MustParse(user), room)
	if len(ws.Rooms[room]) == 0 {
		delete(ws.Rooms, room)
	}
}

// handleConnection is a function that takes a websocket connection and handles it
func (ws *WSServer) handleConnection(conn *websocket.Conn, user string, room string) {
	jsonJoin, err := json.Marshal(Message{
		Author:  user,
		Content: "[joined the room]",
	})
	if err != nil {
		log.Println(err)
		ws.GracefulDisconnect(room, user, conn)
		return
	}
	ws.Broadcast(user, room, jsonJoin)
	for {
		mtype, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("\tUser %s: %s - %d\n", user, room, mtype)
		if strings.Contains(string(msg), "/quit") {
			jsonMsg, err := json.Marshal(Message{
				Author:  user,
				Content: "[left the room]",
			})
			if err != nil {
				log.Println(err)
			} else {
				ws.Broadcast(user, room, jsonMsg)
			}
			break
		}
		var marsh_msg map[string]string
		json.Unmarshal(msg, &marsh_msg)
		jsonMsg, err := json.Marshal(Message{
			Author:  user,
			Content: marsh_msg["chatm"],
		})
		if err != nil {
			ws.SendError(user, err)
			continue
		}
		ws.Broadcast(user, room, jsonMsg)
	}
	ws.GracefulDisconnect(room, user, conn)
}

func (ws *WSServer) Close() {
	// send every connection a close message (8)
	log.Println("Graceful shutdown initiated...")
	for _, v := range ws.Connections {
		if v != nil {
			v.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
	// close the connections
	for k, v := range ws.Connections {
		if v != nil {
			v.Close()
		}
		delete(ws.Connections, k)
	}
	// delete the rooms
	for k := range ws.Rooms {
		delete(ws.Rooms, k)
	}
	log.Println("Graceful shutdown complete")
}

func (ws *WSServer) Broadcast(user, room string, msg []byte) {
	for _, v := range ws.Rooms[room] {
		if v != user {
			if ws.Connections[v] != nil {
				ws.Connections[v].WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}

// SendError is a function that takes a websocket connection and sends an error message
func (ws *WSServer) SendError(user string, err error) {
	jsonErr, _ := json.Marshal(Message{
		Author:  "user",
		Content: err.Error(),
	})
	log.Println(string(jsonErr))
	ws.Connections[user].WriteMessage(websocket.TextMessage, jsonErr)
}

// WSHandler is a function that takes a websocket connection and handles it
func (ws *WSServer) WSHandler(c echo.Context) error {
	if c.Request().Header.Get("Connection") == "Upgrade" {
		params := c.QueryParams()
		room := params.Get("room")
		if room == "" {
			log.Println("Room parameter is required")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "room parameter is required",
			})
		}
		user := params.Get("user")
		if user == "" {
			log.Println("User parameter is required")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "user parameter is required",
			})
		}
		// check if the room exists
		if ws.Rooms[room] == nil {
			log.Println("Room does not exist")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "room does not exist",
			})
		}
		// check if the user is in the room
		check := false
		for _, v := range ws.Rooms[room] {
			if v == user {
				check = true
			}
		}
		if !check {
			log.Println("User is not in the room")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "user is not in the room",
			})
		}
		log.Println("Upgrading connection")
		conn, err := ws.Upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Println(err)
			return err
		}
		ws.Connections[user] = conn
		go ws.handleConnection(conn, user, room)
		log.Printf("User %s connected to room %s\n", user, room)
	}
	log.Println("Invalid request")
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": "invalid request",
	})
}

// getroom is a function that takes a websocket connection and handles it
func (ws *WSServer) GetRoom(c echo.Context) error {
	// generate a unique id for the room
	room := uuid.New()
	// subscribe the user to the room
	id := ws.SubscribeNewUser(room.String())
	return c.JSON(http.StatusOK, map[string]string{
		"room": room.String(),
		"id":   id.String(),
		"type": "chat",
	})
}

func (ws *WSServer) ConnectToRoom(c echo.Context) error {
	// generate a unique id for the room
	room := c.FormValue("room")
	// check if the room exists
	if ws.Rooms[room] == nil {
		log.Println("Room does not exist")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "room does not exist",
		})
	}
	// subscribe the user to the room
	id := ws.SubscribeNewUser(room)
	return c.JSON(http.StatusOK, map[string]string{
		"room": room,
		"id":   id.String(),
		"type": "chat",
	})
}
