package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Qinbeans/chess-htmx/pieces"
	"github.com/Qinbeans/chess-htmx/template"
	"github.com/Qinbeans/chess-htmx/websockets"
	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
)

func menu(c echo.Context) error {
	return c.Render(200, "menu.dj", nil)
}

func room(c echo.Context) error {
	params := c.QueryParams()
	room := params.Get("room")
	if room == "" {
		return c.Redirect(302, "/menu")
	}
	client := params.Get("user")
	if client == "" {
		return c.Redirect(302, "/menu")
	}
	return c.Render(200, "room.dj", pongo2.Context{
		"room":   room,
		"client": client,
	})
}

func main() {
	mode := os.Getenv("Mode")

	address := ":8090"

	log.Println(mode)

	if mode == "release" {
		address = "0.0.0.0:80"
	}

	// init Echo
	server := echo.New()
	// set renderer to our template
	server.Renderer = template.New()
	// gorilla/websocket middleware
	ws := websockets.NewWSServer()
	defer ws.Close()
	chess := pieces.NewServer()
	// Chat
	server.Static("/", "build")
	server.POST("/getroom", ws.GetRoom)
	server.POST("/joinroom", ws.ConnectToRoom)
	server.GET("/", menu)
	server.GET("/room", room)
	server.GET("/room/ws", ws.WSHandler)
	// Chess
	server.POST("/chess/new", chess.NewGame)
	server.POST("/chess/join", chess.ConnectToRoom)
	server.GET("/chess", chess.Room)
	server.GET("/chess/ws", chess.WSHandler)
	data, err := json.MarshalIndent(server.Routes(), "", "  ")
	if err != nil {
		server.Logger.Fatal(err)
	}
	os.WriteFile("routes.json", data, 0644)
	server.Logger.Fatal(server.Start(address))
}
