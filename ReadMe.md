# Programmierworkshop am 19.6.2026

## Namen

## Link zum Git-Repository

## KI-Werkzeuge

### Agenten

### Chat-URLs, z.B. https://chatgpt.com

## Frameworks und Bibliotheken
Rest-Schnittstelle: Gin
Validierung go-playground/validator (bereits in Gin integriert)
OR-Mapping: GORM
Integrationstest: testcontainers-go

### REST-Schnittstelle (Lesen, Neuanlegen, Datei-Upload/-Download)
[Gin](https://github.com/gin-gonic/gin) (`github.com/gin-gonic/gin`) — Routing, JSON-Handling, Middleware.
Endpunkte unter `/rest` (GET lesen/suchen, POST anlegen, PUT ändern, DELETE), Health unter `/health/liveness` und `/health/readiness`.

**Datei-Upload und -Download:** Über die REST-Schnittstelle können Dateien zu einem Parkhaus hochgeladen und wieder heruntergeladen werden. Der Upload erfolgt per `POST /rest/{id}/files` (multipart/form-data), der Download per `GET /rest/{id}/files/{fileId}`. Die Dateien werden in der Datenbank (Tabelle `parkhaus_file`) gespeichert.

### Validierung (nur Neuanlegen)
[go-playground/validator](https://github.com/go-playground/validator) (`github.com/go-playground/validator/v10`, in Gin integriert).
Validierung der Eingabe-DTOs beim Neuanlegen/Ändern; Fehler werden als RFC-9457 Problem Details (HTTP 422) mit Feld-Pfaden zurückgegeben.

### OR-Mapping (für PostgreSQL)
[GORM](https://gorm.io) (`gorm.io/gorm` + `gorm.io/driver/postgres`).
Schema `parkhaus` mit Tabellen `parkhaus`, `adresse`, `auto`, `parkhaus_file`. Geldbetrag `tarifProStunde` via [shopspring/decimal](https://github.com/shopspring/decimal). DDL/Seeding über SQL-Init-Skripte (`extras/compose/postgres/init/parkhaus`).

### Optional: OIDC mit Keycloak
Nicht implementiert (out of scope).

### Einfacher Integrationstest
[testcontainers-go](https://golang.testcontainers.org) — wird im nächsten Schritt ergänzt (echte PostgreSQL im Container für GET/POST).

## Start

```sh
# 1. PostgreSQL inkl. Schema + Seed-Daten starten
docker compose -f extras/compose.yml up -d

# 2. Anwendung starten (Port 8080)
go run ./cmd/server

# Beispiele
curl http://localhost:8080/rest/1000                # Lesen nach ID
curl "http://localhost:8080/rest?size=5"            # Liste (paginiert)
curl "http://localhost:8080/rest?count-only"        # nur Anzahl
curl -X POST http://localhost:8080/rest \
  -H "Content-Type: application/json" \
  -d '{"name":"Parkhaus Neu","kapazitaet":10,"tarifProStunde":2.5,"adresse":{"plz":"68159","ort":"Mannheim","strasse":"Hauptstr","hausnummer":"1"}}'
```

## Makefile-Befehle (Linter, Formatter & Dependency-Check)

Das Projekt enthält ein Makefile mit nützlichen Entwickler-Werkzeugen:

| Befehl | Beschreibung |
|---|---|
| `make fmt` | Formatiert den gesamten Go-Code nach dem offiziellen Standard (`go fmt`) |
| `make lint` | Führt eine statische Code-Analyse mit **golangci-lint** durch (prüft zuerst lokal in `./bin/`, dann global im PATH) |
| `make security` | Scannt Abhängigkeiten auf bekannte Sicherheitslücken mit **govulncheck** |
| `make tidy` | Bereinigt und aktualisiert die `go.mod`-Abhängigkeiten (`go mod tidy`) |
| `make run` | Startet den REST-API-Server |
| `make all` | Führt nacheinander `fmt`, `tidy`, `lint`, `security` aus und startet dann den Server |
| `make help` | Zeigt eine Übersicht aller verfügbaren Befehle |

**Hinweis:** Das Makefile setzt eine Unix-ähnliche Shell-Umgebung voraus. Auf Windows wird Git Bash, MSYS2 oder WSL empfohlen.

## Prompts/Requests an KI-Agent/en
1. Plan Mode:Wir sollen eine Rest Schnittstelle mit der Programmiersprache Go entwickeln. Dabei ist Security mit Keykloak erstmal optional. Die genauen Anforderungen sind "Frameworks und Bibliotheken
   REST-Schnittstelle (Lesen und Neuanlegen)
   Validierung (nur Neuanlegen)
   OR-Mapping (für PostgreSQL)
   Optional: OIDC mit Keycloak
   Einfacher Integrationstest"
2. Nicht so kompliziert. Erstmal nur welche Frameworks und so wie brauchen also welche Produkte.
3. Bevor wir weitermachen solltest du wissen das dieser Server bereits in TypeScript implementiert exisitiert. Ich habe dem Agenten dort die Aufgabe gegeben das Projekt für dich zusammen zu fassen so das du genau weißt was der Server können muss.
4. Ich soll dieses Projekt so in Go bauen. Analysiere das gesamte Projekt und schreib einen Prompt für einen Planing Agent der basierend darauf dann das neue Proekt IMplementiert
