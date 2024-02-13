package main

import (
	"github.com/Qinbeans/chess/pieces"
	"github.com/Qinbeans/chess/template"
	"github.com/Qinbeans/chess/websockets"
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
	// init Echo
	server := echo.New()
	// set renderer to our template
	server.Renderer = template.New()
	// gorilla/websocket middleware
	ws := websockets.NewWSServer()
	defer ws.Close()
	chess := pieces.NewGames()
	// Middleware is a function that takes a handler and returns a handler
	// server.Use(ws.Middleware)
	server.Static("/", "build")
	server.POST("/getroom", ws.GetRoom)
	server.POST("/joinroom", ws.ConnectToRoom)
	server.GET("/menu", menu)
	server.GET("/room", room)
	server.GET("/room/ws", ws.WSHandler)
	server.GET("/newgame", chess.NewGame)
	server.Logger.Fatal(server.Start(":8090"))
}
