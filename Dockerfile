FROM golang:1.20 as builder

# Copy the local package files to the container's workspace.
WORKDIR /app
COPY ./main.go /app/main.go
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
COPY ./websockets /app/websockets
COPY ./pieces /app/pieces
COPY ./utils /app/utils
COPY ./public /app/public
COPY ./template /app/template
# Build the binary.
RUN go build -tags netgo -ldflags '-s -w' -o app

FROM node:18 as style_builder

WORKDIR /app
COPY ./package.json /app/package.json
COPY ./tailwind.config.js /app/tailwind.config.js
COPY ./postcss.config.js /app/postcss.config.js
COPY ./pnpm-lock.yaml /app/pnpm-lock.yaml
# To check for dj classes
COPY ./public /app/public
COPY ./styles /app/styles

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
USER nobody
WORKDIR /app
COPY --from=builder /app/app /app/app
COPY --from=style_builder /app/build /app/build
COPY ./public /app/public

CMD [ "./app" ]
