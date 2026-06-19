# Programmierworkshop am 19.6.2026

## Namen

Jonah Doll, Kilian Schmidt

## Link zum Git-Repository

https://github.com/jonahdoll/parkhaus-2

## KI-Werkzeuge

Cline mit DeepSeek Chat

GitHub Copilot mit Claude Opus 4.8

### Agenten

Cline mit DeepSeek Chat

GitHub Copilot mit Claude Opus 4.8

### Chat-URLs, z.B. https://chatgpt.com

https://www.deepseek.com

https://www.anthropic.com/claude/opus

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
[testcontainers-go](https://golang.testcontainers.org) — echte PostgreSQL im Container für GET/POST.

## Datenmodell (ER-Diagramm)

Das ER-Diagramm ist unter docs/ER-Diagramm.plantuml

## Tests

Das Projekt enthält mehrere Unit- und Integrationstests.

### Unit-Tests
Die Service-Schicht wird mit gemocktem Repository getestet:
- `internal/parkhaus/service/parkhaus_read_test.go`
- `internal/parkhaus/service/parkhaus_write_test.go`
- `internal/parkhaus/service/mock_repository_test.go`

### Integrationstests
Der Handler wird mit einer echten PostgreSQL-Datenbank (via testcontainers-go) getestet:
- `internal/parkhaus/handler/get_endpoints_integration_test.go`

### Ausführen

```sh
make test
```

## Start

```sh
# 1. PostgreSQL inkl. Schema + Seed-Daten starten
docker compose -f extras/compose.yml up -d

# 2. Anwendung starten (Port 8080)
make run

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
1. **Plan-Modus:** Wir sollen eine REST-Schnittstelle mit der Programmiersprache Go entwickeln. Die Absicherung mit Keycloak ist dabei zunächst optional. Die genauen Anforderungen lauten:
   - REST-Schnittstelle (Lesen und Neuanlegen)
   - Validierung (nur beim Neuanlegen)
   - OR-Mapping (für PostgreSQL)
   - Optional: OIDC mit Keycloak
   - Einfacher Integrationstest
2. Bitte nicht zu kompliziert. Lass uns zunächst nur klären, welche Frameworks bzw. Produkte wir überhaupt benötigen.
3. Bevor wir weitermachen, solltest du wissen, dass dieser Server bereits in TypeScript implementiert existiert. Ich habe den dortigen Agenten beauftragt, das Projekt für dich zusammenzufassen, damit du genau weißt, welche Funktionen der Server bieten muss.
4. Ich soll dieses Projekt in Go nachbauen. Analysiere das gesamte Projekt und schreibe einen Prompt für einen Planning-Agenten, der das neue Projekt auf dieser Grundlage anschließend implementiert.

Ist die Grundlegende Struktur dieses GO Projekts passend? Funktionieren Dateien wie @/tools.go richtig und enhalten die Notwendingen Abhöngigkeiten für dev Dependencies einer RestSchnittstelle?

Plan: Behebe die Codestyle Fehler und überprüfe danach dass der korrekte Codestyle eingehalten wurde.

Plan: Wir vollen strukturierten und lesbaren Code. Teile den Service Layer und die Handler in seperierte Dateien für Read und Write auf.

Plan: Füge hinzu das Deutsche Kommentare nicht zu lint fehlern führen.

Plan: Schreibe mir REINE Unit-Tests für den Service Layer.

Wichtig:
1. Keine Integrationstests! Nutze gin.CreateTestContext und httptest.NewRecorder() für Tests rein im Arbeitsspeicher (kein Server-Start).
2. Falls der Code eine DB oder Services nutzt, erstelle einfache Mocks dafür.
3. Schreibe die Tests als typische Go "Table-Driven Tests" (Test-Strukturen in einer Schleife).
4. Teste den Erfolgsfall (200 OK) sowie Fehlerfälle (z.B. ungültiges JSON oder DB-Fehler).
5. Nutze github.com/stretchr/testify/assert für die Überprüfungen.

Plan: Füge zu der Make File hinzu, dass Unit-Tests, Integrations-Tests und beides ausgeührt werden kann

Plan: Fix die codestyle Fehler
