package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Qinbeans/chess-htmx/static"
	"github.com/Qinbeans/chess-htmx/template"
	"github.com/Qinbeans/chess-htmx/websockets"
	"github.com/labstack/echo/v4"
)

func menu(c echo.Context) error {
	// check if this a boost request using header. Hx-Boosted should be true
	if c.Request().Header.Get("Hx-Boosted") == "true" {
		return c.Render(200, "home-comp", map[string]any{
			"title":       "Menu",
			"description": "Choose what you want to do at Chess-HTMX",
			"items": []map[string]any{
				{"selected": true, "label": "Home"},
				{"selected": false, "url": "/chat_menu", "label": "Chat"},
				{"selected": false, "url": "/chess_menu", "label": "Chess"},
			},
		})
	}
	return c.Render(200, "home-page", map[string]any{
		"title":       "Menu",
		"description": "Choose what you want to do at Chess-HTMX",
		"items": []map[string]any{
			{"selected": true, "label": "Home"},
			{"selected": false, "url": "/chat_menu", "label": "Chat"},
			{"selected": false, "url": "/chess_menu", "label": "Chess"},
		},
	})
}

// func chat(c echo.Context) error {
// 	params := c.QueryParams()
// 	room := params.Get("room")
// 	if room == "" {
// 		return c.Redirect(302, "/")
// 	}
// 	client := params.Get("user")
// 	if client == "" {
// 		return c.Redirect(302, "/")
// 	}
// 	return c.Render(200, "chat-page", map[string]any{
// 		"title":       "Chat Room",
// 		"description": "Chat with a friend",
// 		"room":        room,
// 		"client":      client,
// 	})
// }

func chat_menu(c echo.Context) error {
	// check if this a boost request using header. Hx-Boosted should be true
	if c.Request().Header.Get("Hx-Boosted") == "true" {
		return c.Render(200, "chat-comp", map[string]any{
			"title":       "Chat",
			"description": "Chat with a friend",
			"items": []map[string]any{
				{"selected": false, "url": "/", "label": "Home"},
				{"selected": true, "label": "Chat"},
				{"selected": false, "url": "/chess_menu", "label": "Chess"},
			},
		})
	}
	return c.Render(200, "chat-page", map[string]any{
		"title":       "Chat",
		"description": "Chat with a friend",
		"items": []map[string]any{
			{"selected": false, "url": "/", "label": "Home"},
			{"selected": true, "label": "Chat"},
			{"selected": false, "url": "/chess_menu", "label": "Chess"},
		},
	})
}

func chess_menu(c echo.Context) error {
	// check if this a boost request using header. Hx-Boosted should be true
	if c.Request().Header.Get("Hx-Boosted") == "true" {
		return c.Render(200, "chess-comp", map[string]any{
			"title":       "Chess",
			"description": "Play chess with a friend",
			"items": []map[string]any{
				{"selected": false, "url": "/", "label": "Home"},
				{"selected": false, "url": "/chat_menu", "label": "Chat"},
				{"selected": true, "label": "Chess"},
			},
		})
	}
	return c.Render(200, "chess-page", map[string]any{
		"title":       "Chess",
		"description": "Play chess with a friend",
		"items": []map[string]any{
			{"selected": false, "url": "/", "label": "Home"},
			{"selected": false, "url": "/chat_menu", "label": "Chat"},
			{"selected": true, "label": "Chess"},
		},
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
	server.Use(static.Middleware())
	// set renderer to our template
	server.Renderer = template.New()
	// gorilla/websocket middleware
	ws := websockets.NewWSServer()
	defer ws.Close()
	// chess := pieces.NewServer()
	// Chat
	server.Static("/", "build")
	// server.POST("/getroom", ws.GetRoom)
	// server.POST("/joinroom", ws.ConnectToRoom)
	server.GET("/", menu)
	server.GET("/chat_menu", chat_menu)
	// server.GET("/room/ws", ws.WSHandler)
	// Chess
	// server.POST("/chess/new", chess.NewGame)
	// server.POST("/chess/join", chess.ConnectToRoom)
	server.GET("/chess_menu", chess_menu)
	// server.GET("/chess/ws", chess.WSHandler)
	data, err := json.MarshalIndent(server.Routes(), "", "  ")
	if err != nil {
		server.Logger.Fatal(err)
	}
	os.WriteFile("routes.json", data, 0644)
	server.Logger.Fatal(server.Start(address))
}
