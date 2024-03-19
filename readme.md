[![Build Status](https://github.com/Qinbeans/chess-htmx/actions/workflows/docker-image.yml/badge.svg)](https://github.com/Qinbeans/chess-htmx/actions/workflows/docker-image.yml)

# Revision 1

What I realized after reading further into the documentation was that HTMX was less a framework and more a component for a backend framework. My problems with HTMX make so much more sense now as I treated HTMX like frontend frameworks commonly used. I also chose to switch templating engines, which I did because Pongo2 hadn't been updated recently.

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
    "method": "GET",
    "path": "/*",
    "name": "github.com/labstack/echo/v4.StaticDirectoryHandler.func1"
  },
  {
    "method": "GET",
    "path": "/",
    "name": "main.menu"
  },
  {
    "method": "GET",
    "path": "/chat_menu",
    "name": "main.chat_menu"
  },
  {
    "method": "GET",
    "path": "/chess_menu",
    "name": "main.chess_menu"
  }
]
```
