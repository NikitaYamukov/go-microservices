FROM golang:1.25.7-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# копируем весь код
COPY . .

# если main.go в корне проекта
RUN go build -o app ./cmd

# --- финальный образ ---
FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/app .

CMD ["./app"]
