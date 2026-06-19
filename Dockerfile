# --- Build-Stage ---
FROM golang:1.26 AS build
WORKDIR /src

# Abhängigkeiten cachen
COPY go.mod go.sum ./
RUN go mod download

# Quellcode kopieren und bauen
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

# --- Runtime-Stage ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]

