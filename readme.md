[![Build Status](https://github.com/Qinbeans/chess-htmx/actions/workflows/docker-image.yml/badge.svg)](https://github.com/Qinbeans/chess-htmx/actions/workflows/docker-image.yml)

# Chess In Go and HTMX

Can be found [here](https://chess-htmx.onrender.com/)

This is a test on how far I can take my skills as a web developer. I'm quite confident in my abilities with SvelteKit and Go, but I don't really make websites with Go and Javascript alone. I read that HTMX is a wonderful framework for frontend and I believe it's worth a shot.

## Building

There's really no build process aside from building the classes from TailwindCSS. I wrote a custom config for Air (live reload for Go) which automatically builds the styles.

## One command to rule them all

That's somewhat a lie as `air` is for development--you can install air [here](https://github.com/cosmtrek/air).

Here are the steps to build and deploy:
- `docker build . --tag="chess-htmx"`
- `docker run chess-htmx`

or

- `docker compose up`

## Endpoints

You can find the endpoints in `routes.json`.

```json
[
  {
    "method": "POST",
    "path": "/getroom",
    "name": "github.com/Qinbeans/chess-htmx/websockets.(*WSServer).GetRoom-fm"
  },
  {
    "method": "GET",
    "path": "/room/ws",
    "name": "github.com/Qinbeans/chess-htmx/websockets.(*WSServer).WSHandler-fm"
  },
  {
    "method": "POST",
    "path": "/chess/new",
    "name": "github.com/Qinbeans/chess-htmx/pieces.(*Server).NewGame-fm"
  },
  {
    "method": "POST",
    "path": "/chess/join",
    "name": "github.com/Qinbeans/chess-htmx/pieces.(*Server).ConnectToRoom-fm"
  },
  {
    "method": "GET",
    "path": "/chess",
    "name": "github.com/Qinbeans/chess-htmx/pieces.(*Server).Room-fm"
  },
  {
    "method": "GET",
    "path": "/*",
    "name": "github.com/labstack/echo/v4.StaticDirectoryHandler.func1"
  },
  {
    "method": "POST",
    "path": "/joinroom",
    "name": "github.com/Qinbeans/chess-htmx/websockets.(*WSServer).ConnectToRoom-fm"
  },
  {
    "method": "GET",
    "path": "/",
    "name": "main.menu"
  },
  {
    "method": "GET",
    "path": "/room",
    "name": "main.room"
  },
  {
    "method": "GET",
    "path": "/chess/ws",
    "name": "github.com/Qinbeans/chess-htmx/pieces.(*Server).WSHandler-fm"
  }
]
```
