# Programmierworkshop am 19.6.2026

## Namen

Jonah Doll, Kilian Schmitt

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

- **REST-Schnittstelle:** Gin
- **Validierung:** go-playground/validator (bereits in Gin integriert)
- **OR-Mapping:** GORM
- **Integrationstest:** testcontainers-go

### REST-Schnittstelle (Lesen, Neuanlegen, Ändern, Löschen, Datei-Upload/-Download)
[Gin](https://github.com/gin-gonic/gin) (`github.com/gin-gonic/gin`) — Routing, JSON-Handling, Middleware.
Alle fachlichen Endpunkte liegen unter `/rest`, die Health-Checks unter `/health/liveness` und `/health/readiness` (liefern jeweils `{"status":"up"}`).

#### Endpunkt-Übersicht

| Methode & Pfad | Beschreibung |
|---|---|
| `GET /rest/:id` | Parkhaus nach ID (inkl. Adresse), setzt `ETag` |
| `GET /rest` | Suche/Liste mit Pagination und Filtern (`name`, `kapazitaet`, `tarifProStunde`, `page`, `size`); `count-only` liefert nur die Anzahl |
| `GET /rest/file/:id` | Datei-Download zu einem Parkhaus |
| `POST /rest` | Neues Parkhaus anlegen (transaktional inkl. Adresse + Autos), `Location`-Header |
| `PUT /rest/:id` | Parkhaus ändern (erfordert `If-Match`), setzt neuen `ETag` |
| `POST /rest/:id/autos` | Auto zu einem Parkhaus hinzufügen (Kapazitätsprüfung) |
| `POST /rest/:id` | Datei-Upload (multipart/form-data, Feld `file`) |
| `DELETE /rest/:id` | Parkhaus löschen (idempotent, immer `204`) |
| `GET /health/liveness`, `GET /health/readiness` | Health-Checks |

**Optimistische Synchronisation (ETag / If-Match / If-None-Match):**
Lese-Antworten enthalten den Header `ETag: "<version>"`. Beim Lesen kann mit `If-None-Match` ein `304 Not Modified` ausgelöst werden. Änderungen via `PUT` erfordern den Header `If-Match`; fehlt er, wird `428 Precondition Required` zurückgegeben, bei ungültiger oder veralteter Versionsnummer `412 Precondition Failed`. Erfolgreiche Änderungen erhöhen die Version um 1 und liefern `204 No Content` mit neuem `ETag`.

**Datei-Upload und -Download:** Über die REST-Schnittstelle kann genau **eine** Datei pro Parkhaus hochgeladen und wieder heruntergeladen werden (1:1-Beziehung). Der Upload erfolgt per `POST /rest/:id` (multipart/form-data, Feld `file`); eine bereits vorhandene Datei wird dabei ersetzt. Der Download erfolgt per `GET /rest/file/:id` mit den Headern `Content-Type` (gespeicherter MIME-Typ, Fallback `application/octet-stream`) und `Content-Disposition: attachment`. Die Dateien werden in der Datenbank (Tabelle `parkhaus_file`, Spalte `data` als `bytea`) gespeichert.

### Validierung (Neuanlegen und Ändern)
[go-playground/validator](https://github.com/go-playground/validator) (`github.com/go-playground/validator/v10`, in Gin integriert).
Validierung der Eingabe-DTOs beim Neuanlegen (`POST`) und Ändern (`PUT`); Fehler werden als RFC-9457 Problem Details (HTTP 422) mit Feld-Pfaden zurückgegeben.

### OR-Mapping (für PostgreSQL)
[GORM](https://gorm.io) (`gorm.io/gorm` + `gorm.io/driver/postgres`).
Schema `parkhaus` mit Tabellen `parkhaus`, `adresse`, `auto`, `parkhaus_file`. Geldbetrag `tarifProStunde` via [shopspring/decimal](https://github.com/shopspring/decimal) (in der JSON-Response als Zahl serialisiert). DDL/Seeding erfolgt über SQL-Init-Skripte (`extras/compose/postgres/init/parkhaus`, `01-create.sql` + `02-copy-csv.sql`), **nicht** über GORM-AutoMigrate.

### Optional: OIDC mit Keycloak
Nicht in die Anwendung integriert (out of scope). Für eine spätere Anbindung sind unter `extras/keycloak/` bereits Infrastruktur-Artefakte (Compose, TLS) vorbereitet; das Paket `internal/security/` ist als Platzhalter vorgesehen.

### Einfacher Integrationstest
[testcontainers-go](https://golang.testcontainers.org) — echte PostgreSQL im Container für die GET-Endpunkte.

## Datenmodell (ER-Diagramm)

Das ER-Diagramm ist unter docs/ER-Diagramm.plantuml

## Konfiguration

Die Anwendung wird über Umgebungsvariablen konfiguriert (sinnvolle Defaults sind hinterlegt):

| Variable | Default | Beschreibung |
|---|---|---|
| `PORT` | `8080` | HTTP-Port des Servers |
| `DATABASE_URL` | `postgres://parkhaus:parkhaus@localhost:5432/parkhaus?sslmode=disable&search_path=parkhaus` | PostgreSQL-Connection-String (Schema `parkhaus`) |
| `LOG_LEVEL` | `info` | Log-Level (`debug`, `info`, `warn`, `error`) |

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
curl http://localhost:8080/rest/1000                # Lesen nach ID (liefert ETag-Header)
curl "http://localhost:8080/rest?size=5"            # Liste (paginiert)
curl "http://localhost:8080/rest?count-only"        # nur Anzahl
curl -X POST http://localhost:8080/rest \
  -H "Content-Type: application/json" \
  -d '{"name":"Parkhaus Neu","kapazitaet":10,"tarifProStunde":2.5,"adresse":{"plz":"68159","ort":"Mannheim","strasse":"Hauptstr","hausnummer":"1"}}'

# Ändern (optimistische Sperre über If-Match mit der aktuellen Version)
curl -X PUT http://localhost:8080/rest/1000 \
  -H "Content-Type: application/json" \
  -H 'If-Match: "0"' \
  -d '{"name":"Parkhaus Neu","kapazitaet":20,"tarifProStunde":3.0}'
```

### Docker-Image der Anwendung

Das Projekt enthält ein mehrstufiges `Dockerfile` (Build mit `golang:1.26`, schlankes
distroless-Runtime-Image). Image bauen und starten:

```sh
docker build -t parkhaus-2 .
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://parkhaus:parkhaus@host.docker.internal:5432/parkhaus?sslmode=disable&search_path=parkhaus" \
  parkhaus-2
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
