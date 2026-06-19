# Planning-Prompt: Portierung „parkhaus“ (TypeScript → Go)

> **Rolle:** Du bist ein **Planning Agent**. Erstelle aus diesem Dokument einen
> vollständigen, umsetzbaren **Implementierungsplan** für die Neu-Implementierung
> der bestehenden REST-Schnittstelle des Projekts `parkhaus` in **Go**.
> Anschließend soll der Plan von einem Coding-Agent **Schritt für Schritt** umgesetzt werden.

---

## 1. Auftrag (in einem Satz)

Portiere die **REST-Schnittstelle** des bestehenden TypeScript/Hono/Prisma-Projekts
„parkhaus“ nach **Go**, mit PostgreSQL als Datenbank — **funktional gleichwertig**
zur bestehenden REST-API.

## 2. Scope (verbindlich)

**IN Scope:**

- Komplette **REST-API** für Parkhäuser inkl. Adresse, Autos und Datei-Upload/-Download.
- **Persistenz** mit PostgreSQL (gleiches Schema wie unten beschrieben).
- **Pagination**, **flexible Suche** per Query-Parametern.
- **Optimistische Synchronisation** über `ETag` / `If-Match` / `If-None-Match` (Versionsnummer).
- **RFC 9457 Problem Details** für Fehlerantworten.
- **Konfiguration** (Port, DB-URL, Logging) über Umgebungsvariablen und/oder Config-Datei.
- **Strukturiertes Logging**.
- Optional: `docker-compose` für PostgreSQL + lauffähiges `Dockerfile` für die Go-App.

**OUT of Scope (NICHT implementieren):**

- ❌ **Keine Authentifizierung / Autorisierung** (kein Keycloak, kein JWT, keine Rollen).
  → Die im Original geschützten Endpunkte (POST/PUT/DELETE/Upload) werden hier **ohne**
  Auth-Middleware bereitgestellt.
- ❌ **Keine Tests** (weder Unit- noch Integrationstests).
- ❌ Kein GraphQL.
- ❌ Kein Mail-Versand (die „fire-and-forget“-Mail beim Anlegen entfällt).
- ❌ Keine Prometheus-/Metrics-Endpunkte.

## 3. Technologie-Entscheidung (vom Planning Agent zu treffen)

Wähle einen **idiomatischen Go-Stack** und **begründe** die Wahl kurz im Plan.
Triff explizit Entscheidungen zu:

1. **HTTP-Router/Framework** (z. B. `net/http` + `chi`, `gin`, `echo`, `fiber`).
2. **DB-Zugriff** (z. B. `pgx` + `sqlc`, `database/sql` + `squirrel`, oder ORM wie `GORM`/`Ent`).
   - Berücksichtige: das Original nutzt ein **separates Postgres-Schema `parkhaus`**,
     `DECIMAL(8,2)`, ein Postgres-**Enum** `kundentyp`, und `BYTEA` für Datei-Uploads.
3. **Validierung** (z. B. `go-playground/validator`, manuell).
4. **Logging** (z. B. `log/slog`, `zerolog`, `zap`).
5. **Konfiguration** (z. B. `env` via `os`/`viper`/`envconfig`, optional TOML).
6. **Decimal-Handling** für `tarifProStunde` (z. B. `shopspring/decimal`) — Geldbetrag, kein float-Rundungsfehler.

> Hinweis: Empfohlene, bewährte Default-Wahl (falls keine besseren Gründe dagegensprechen):
> `chi` + `pgx`/`sqlc` + `slog` + `shopspring/decimal` + `go-playground/validator`.
> Der Planning Agent darf abweichen, muss es aber begründen.

## 4. Vorgeschlagene Projektstruktur (Richtwert, Go-Konventionen)

```
parkhaus-go/
├── cmd/server/main.go            # Entry-Point: Config laden, DB verbinden, Server starten, Graceful Shutdown
├── internal/
│   ├── config/                   # Env/Config-Parsing (Port, DATABASE_URL, LogLevel)
│   ├── server/                   # Router-Setup, Middleware (Logging, Recover, CORS, SecureHeaders, Compression)
│   ├── parkhaus/
│   │   ├── handler/              # HTTP-Handler (entspricht TS „router“)
│   │   ├── service/              # Geschäftslogik Lesen + Schreiben (entspricht TS „service“)
│   │   ├── repository/           # DB-Zugriff (Prisma-Ersatz)
│   │   ├── model/                # Domänen-Structs + DTOs
│   │   └── apperr/               # Domänen-Fehlertypen
│   └── problemdetails/           # RFC 9457 Problem-Details-Helper
├── migrations/ oder db/          # SQL-Schema (create/drop), Seed-Daten
├── go.mod / go.sum
├── Dockerfile
└── compose.yml                   # PostgreSQL (+ optional App)
```

Mappe die Struktur sinnvoll; die Schichten **Handler → Service → Repository** sollen
sauber getrennt sein (wie im Original `router → service → prismaClient`).

---

## 5. Datenmodell (1:1 zu übernehmen)

Schema-Name in PostgreSQL: **`parkhaus`**. Tabellen (Spaltennamen in snake_case):

### `parkhaus`
| Spalte             | Typ                | Constraints / Default |
|--------------------|--------------------|-----------------------|
| `id`               | integer (Identity) | PK, `GENERATED ALWAYS AS IDENTITY (START WITH 1000)` |
| `version`          | integer            | NOT NULL, DEFAULT `0` (optimistic locking) |
| `name`             | text               | NOT NULL, **UNIQUE** |
| `kapazitaet`       | integer            | NOT NULL, CHECK `>= 0` |
| `tarif_pro_stunde` | decimal(8,2)       | NOT NULL (Geldbetrag) |
| `erzeugt`          | timestamp          | NOT NULL, DEFAULT `NOW()` |
| `aktualisiert`     | timestamp          | NOT NULL, DEFAULT `NOW()` (bei Update aktualisieren) |

### `adresse` (1:1 zu Parkhaus)
| Spalte        | Typ     | Constraints |
|---------------|---------|-------------|
| `id`          | integer (Identity) | PK |
| `plz`         | text    | NOT NULL, CHECK Format `\d{5}` |
| `ort`         | text    | NOT NULL |
| `strasse`     | text    | NOT NULL |
| `hausnummer`  | text    | NOT NULL |
| `parkhaus_id` | integer | NOT NULL, **UNIQUE**, FK → `parkhaus(id)` ON DELETE CASCADE |

### `auto` (n:1 zu Parkhaus)
| Spalte          | Typ                | Constraints |
|-----------------|--------------------|-------------|
| `id`            | integer (Identity) | PK |
| `kennzeichen`   | text               | NOT NULL |
| `einfahrtszeit` | timestamp          | NOT NULL |
| `kundentyp`     | enum `kundentyp`   | NOT NULL — Werte: `PREMIUM`, `BASIS`, `ANWOHNER` |
| `parkhaus_id`   | integer            | NOT NULL, FK → `parkhaus(id)` ON DELETE CASCADE, Index |

### `parkhaus_file` (1:1 zu Parkhaus, Binärdatei)
| Spalte        | Typ     | Constraints |
|---------------|---------|-------------|
| `id`          | integer (Identity) | PK |
| `data`        | bytea   | NOT NULL |
| `filename`    | text    | NOT NULL |
| `mimetype`    | text    | NULLABLE |
| `parkhaus_id` | integer | NOT NULL, **UNIQUE**, FK → `parkhaus(id)` ON DELETE CASCADE |

> **Postgres-Enum:** `CREATE TYPE kundentyp AS ENUM ('PREMIUM', 'BASIS', 'ANWOHNER');`
> Referenz-SQL (Schema-Erzeugung & Drop) liegt im Original unter
> `src/config/resources/postgresql/{create-table,drop-table,copy-csv}.sql`.
> Seed-Daten optional aus CSV (parkhaus, adresse, auto) — kann als SQL-INSERTs umgesetzt werden.

**Wichtige Datentyp-Hinweise für Go:**

- `tarif_pro_stunde`: in der **JSON-Response als `number`** ausgeben (das TS-Projekt wandelt
  `Decimal` → `number` um). Intern mit Decimal-Typ rechnen, beim Marshalling als Zahl serialisieren.
- `version`: steuert ETag (`"<version>"`).
- Zeitstempel als RFC 3339/ISO 8601 serialisieren.

---

## 6. REST-Endpunkte (vollständige Spezifikation)

**Basis-Pfad:** `/rest`
Health: `/health/liveness`, `/health/readiness` (beide liefern `{"status":"up"}`).

### 6.1 Lesen (GET)

#### `GET /rest/:id` — Parkhaus nach ID
- `:id` muss numerisch sein → sonst **404**.
- **Accept**-Handling: ist `Accept` gesetzt und **weder** `*/*` **noch** `json`/`html`,
  dann **406** (leerer Body).
- Lädt Parkhaus **inkl. Adresse** (Autos nur optional; im Original Default ohne Autos bei diesem Endpunkt).
- **ETag**: setze Header `ETag: "<version>"`.
- **If-None-Match**: wenn Client `If-None-Match: "<version>"` schickt und es passt → **304** (leerer Body).
- Erfolg: **200** + JSON des Parkhauses.
- Nicht gefunden: **404**.

#### `GET /rest` — Suche / Liste mit Pagination
- **Accept**-Handling wie oben (406 bei inkompatiblem Accept).
- Query-Parameter:
  - `count-only` (Präsenz genügt): liefert nur `{"count": <n>}` (Anzahl aller Parkhäuser).
  - `page` (1-basiert in der API; intern 0-basiert), `size`.
  - **Such-Filter** (alle optional, kombinierbar):
    - `name` → exakte Übereinstimmung (`equals`).
    - `kapazitaet` → numerisch, Filter `kapazitaet <= wert` (**lte**).
    - `tarifProStunde` → numerisch, Filter `tarif_pro_stunde <= wert` (**lte**).
  - **Unbekannter Filter-Parameter** (z. B. `foo`) → **404** („Ungueltige Suchparameter“).
- Lädt Parkhäuser **inkl. Adresse**.
- Erfolg: **200** + **Page-Objekt** (Format siehe 6.4).
- Keine Treffer (auch bei ungültiger Seite) → **404**.

#### `GET /rest/file/:id` — Datei-Download zu Parkhaus
- `:id` numerisch, sonst **404**.
- Lädt `parkhaus_file` zu `parkhaus_id`.
- Setzt Header:
  - `Content-Type` = gespeicherter `mimetype` (Fallback `application/octet-stream`).
  - `Content-Disposition: attachment; filename="<filename>"`.
- Body = Bytes der Datei. Keine Datei → **404**.

### 6.2 Schreiben (POST/PUT/DELETE) — **ohne Auth** (wichtig: kein JWT/Rollen)

#### `POST /rest` — Neues Parkhaus anlegen
- Request-Body JSON (Validierung siehe 6.3), enthält `name`, `kapazitaet`,
  `tarifProStunde`, `adresse{...}`, optional `autos[]`.
- **Kapazitätsprüfung beim Anlegen:** Anzahl mitgelieferter Autos darf `kapazitaet`
  **nicht überschreiten** → sonst **422**.
- **Name-Eindeutigkeit:** existiert `name` bereits → **422** („Ein Parkhaus mit dem Namen … existiert bereits.“).
- Anlegen erfolgt **transaktional** inkl. Adresse + Autos.
- Erfolg: **201** + Header `Location: <basis-url>/<neueId>` (kein Body).

#### `PUT /rest/:id` — Parkhaus ändern
- `:id` numerisch, sonst **404**.
- **Pflicht-Header `If-Match`** mit `"<version>"`:
  - fehlt → **428 Precondition Required** (Problem Details, Detail enthält `If-Match`).
  - Format ungültig (Regex `^"\d{1,3}"`) → **412 Precondition Failed** („Versionsnummer … ist ungueltig“).
  - Version veraltet (kleiner als aktuelle DB-Version) → **412** („Versionsnummer … ist nicht aktuell“).
- Body: `name`, `kapazitaet`, `tarifProStunde` (nur diese drei Felder, siehe Update-Schema).
- Bei Erfolg: Version **+1**, **204 No Content** + Header `ETag: "<neueVersion>"`.
- Nicht gefunden → **404**.

#### `POST /rest/:id/autos` — Auto zu Parkhaus hinzufügen
- `:id` numerisch, sonst **404**.
- Body: ein Auto (`kennzeichen`, `einfahrtszeit`, `kundentyp`).
- **Kapazitätsprüfung:** aktuelle Anzahl Autos `>= kapazitaet` → **422**
  („… hat keine freie Kapazitaet mehr …“).
- Parkhaus existiert nicht → **404**.
- Erfolg: **201** + Header `Location: <url>/<autoId>` (Format `.../rest/<parkhausId>/autos/<autoId>`).

#### `DELETE /rest/:id` — Parkhaus löschen
- `:id` nicht numerisch → trotzdem **204** (idempotentes Verhalten im Original).
- Existiert nicht → **204** (kein Fehler).
- Existiert → transaktional löschen (Cascade entfernt Adresse/Autos/File) → **204**.

#### `POST /rest/:id` — Datei-Upload zu Parkhaus (multipart/form-data)
- `:id` numerisch, sonst **404**.
- Multipart-Form-Feld **`file`**:
  - fehlt oder mehrere Dateien → **400** (Problem Details).
  - kein gültiger Datei-Typ → **400**.
- Vorhandene Datei zum Parkhaus wird **ersetzt** (alte löschen, neue speichern) — transaktional.
- Parkhaus existiert nicht → **404**.
- Erfolg: **204** + Header `Location: <basis-url>/file/<id>`.

### 6.3 Validierungsregeln (entspricht Zod-Schemas im Original)

**Parkhaus (Neu — POST /rest):**
- `name`: String, Regex `^\w.*`, max. **100** Zeichen.
- `kapazitaet`: Integer, **> 0**.
- `tarifProStunde`: Zahl, **>= 0**.
- `adresse` (Pflicht):
  - `plz`: String, 1–10 Zeichen.
  - `ort`: String, 1–100 Zeichen.
  - `strasse`: String, 1–100 Zeichen.
  - `hausnummer`: String, 1–10 Zeichen.
- `autos` (optional): Array von Auto.

**Auto:**
- `kennzeichen`: String, 1–20 Zeichen.
- `einfahrtszeit`: Datum (ISO 8601 / coercible).
- `kundentyp`: Enum `PREMIUM` | `BASIS` | `ANWOHNER`.

**Parkhaus (Update — PUT /rest/:id):**
- Nur `name`, `kapazitaet`, `tarifProStunde` (gleiche Regeln wie oben).

**Bei Validierungsfehler → HTTP 422** mit Problem Details, wobei `detail` ein Array von
Fehlern ist, die jeweils einen **`path`** enthalten (das TS-Projekt liefert Zod-`issues`;
in Go ein äquivalentes Array mit Feldnamen, z. B. `[{"path":["name"], "message":"..."}]`).
Die Feldnamen (`name`, `kapazitaet`, `tarifProStunde`, `kennzeichen`, `einfahrtszeit`,
`kundentyp`) müssen erkennbar sein.

### 6.4 Antwort-Formate

**Page-Objekt** (für `GET /rest`):
```json
{
  "content": [ /* Array von Parkhaus-DTOs (inkl. adresse) */ ],
  "page": {
    "size": 5,
    "number": 0,
    "totalElements": 42,
    "totalPages": 9
  }
}
```
- `number` ist die **0-basierte** interne Seitennummer.
- `totalPages = ceil(totalElements / size)`.

**Pagination-Defaults:**
- `DEFAULT_PAGE_SIZE = 5`, `MAX_PAGE_SIZE = 100`, `DEFAULT_PAGE_NUMBER = 0`.
- API-`page` ist 1-basiert: intern `number = page - 1` (bei ungültig/<=0 → 0).
- `size`: bei ungültig oder außerhalb 1..100 → Default 5.

**Parkhaus-DTO (JSON):** wie DB-Modell, aber:
- `tarifProStunde` als **Zahl** (nicht String).
- Property-Namen im JSON **camelCase**: `tarifProStunde`, `parkhausId` etc.
  (Achtung: DB ist snake_case, JSON-Tags entsprechend setzen.)
- Adresse eingebettet als `adresse`.

### 6.5 Problem Details (RFC 9457)

Antwort-Header bei Fehlern: `Content-Type: application/problem+json`.
Body-Schema:
```json
{ "title": "...", "statusCode": 422, "detail": <string|array> }
```
Titel-Mapping nach Statuscode:
- 400 → „Bad Request“
- 401 → „Unauthorized“ (im Scope hier i. d. R. nicht nötig)
- 403 → „Forbidden“ (nicht nötig)
- 412 → „Precondition Failed“
- 422 → „Unprocessable Content“
- 428 → „Precondition Required“
- sonst → „Client Error“

### 6.6 Fehler-Mapping (Domänenfehler → HTTP)

| Domänenfehler (Original)             | HTTP-Status | Problem-Title             |
|--------------------------------------|-------------|---------------------------|
| NotFound                             | 404         | (Standard 404, kein PD)   |
| Validierungsfehler                   | 422         | Unprocessable Content     |
| ParkhausExists (Name doppelt)        | 422         | Unprocessable Content     |
| KapazitaetUeberschritten             | 422         | Unprocessable Content     |
| VersionInvalid / VersionOutdated     | 412         | Precondition Failed       |
| (If-Match fehlt)                     | 428         | Precondition Required     |
| Sonstiger interner Fehler            | 500         | (Plain „Interner Fehler“) |

---

## 7. Cross-Cutting Concerns

- **Middleware (Reihenfolge wie Original sinnvoll):** Secure Headers
  (`X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`), CORS, Compression,
  Recover/Panic-Handler, Request-Logging (nur im Debug-Level).
- **CORS-Origins** (aus Original übernehmen): `https://localhost:4200`, `http://localhost:4200`,
  `https://localhost:5173`, `http://localhost:5173`, `http://localhost:8843`, `http://localhost:8880`.
  Methoden: `GET, HEAD, POST, PUT, DELETE`. Exponierte Header u. a.: `ETag`, `Location`, `Content-Type`.
- **Logging:** strukturiert, Log-Level via Env/Config (`debug`, `info`, `warn`, `error`).
- **Konfiguration (Env-Variablen):**
  - `DATABASE_URL` (PostgreSQL-Connection-String; Schema `parkhaus`).
  - Server-Port (Default **3000**; HTTP-Variante optional).
  - Log-Level.
  - (Original nutzt zusätzlich TOML `app.toml` + TLS-Zertifikate — **TLS ist optional**;
    für den Port kann HTTP genügen, sofern nicht anders gewünscht.)
- **Graceful Shutdown:** auf `SIGINT`/`SIGTERM` DB-Pool sauber schließen, Server beenden.
- **DB-Seeding (optional, Dev):** Schema anlegen + Beispieldaten laden (analog `db_populate`).
  Da Auth out-of-scope ist, kann dies als **CLI-Flag/Startup-Option** statt geschütztem
  `/dev`-Endpunkt erfolgen.

---

## 8. Akzeptanzkriterien (Definition of Done)

Der Go-Port gilt als fertig, wenn:

1. Alle Endpunkte aus **Abschnitt 6** existieren und das **gleiche Verhalten** (Statuscodes,
   Header `ETag`/`Location`/`Content-Disposition`, Body-Formate) wie oben spezifiziert zeigen.
2. **Pagination** (Page-Objekt, Defaults, 1→0-Index-Mapping) korrekt funktioniert.
3. **Suche** (`name` equals, `kapazitaet`/`tarifProStunde` als `lte`, unbekannter Param → 404,
   `count-only`) korrekt funktioniert.
4. **Optimistische Sperre** über `If-Match`/`ETag` inkl. der Statuscodes **428/412/204** korrekt ist.
5. **Validierung** die Felder gemäß 6.3 prüft und Fehler als **422 Problem Details** mit Feld-Pfaden liefert.
6. **Datei-Upload/-Download** (multipart, BYTEA, Replace-Semantik) funktioniert.
7. **Kapazitätslogik** beim Anlegen und beim Hinzufügen von Autos greift (**422**).
8. Projekt **kompiliert** (`go build ./...`) und **startet** gegen eine PostgreSQL-DB
   (per `compose.yml` bereitstellbar); `go vet ./...` ist sauber.
9. **Kein** Code für Auth, Tests, GraphQL, Mail, Metrics vorhanden (Scope eingehalten).

---

## 9. Aufgabe an den Planning Agent (Output-Erwartung)

Erzeuge einen **strukturierten Implementierungsplan** mit:

1. **Tech-Stack-Entscheidung** inkl. kurzer Begründung pro Komponente (siehe Abschnitt 3).
2. **Finale Verzeichnis-/Paketstruktur** (konkret, dateigenau).
3. **Geordnete Arbeitspakete / Milestones** (z. B. M1 Projekt-Setup & Config,
   M2 DB-Schema & Repository, M3 Domänenmodelle & Fehler/Problem-Details,
   M4 Service-Schicht inkl. Validierung/Kapazität/Versionierung, M5 HTTP-Handler & Routing,
   M6 Middleware/Logging/CORS, M7 Datei-Upload/Download, M8 Docker/Compose & Seeding,
   M9 Doku/README). Jedes Paket mit konkreten Schritten und betroffenen Dateien.
4. **Mapping-Tabelle**: TS-Datei/Konzept → Go-Pendant (für Nachvollziehbarkeit).
5. **Risiken/Edge-Cases** (z. B. Decimal-Serialisierung als Zahl, Postgres-Enum-Mapping,
   Identity-Sequenz `START WITH 1000`, snake_case↔camelCase, `count-only` ohne Wert,
   DELETE-Idempotenz bei nicht-numerischer ID, `If-Match`-Regex).
6. Hinweis, dass der Plan anschließend **iterativ** vom Coding-Agent umgesetzt wird,
   inkl. Validierung per `go build`/`go vet` nach jedem Milestone.

> Beginne mit der Tech-Stack-Entscheidung, dann die Struktur, dann die Milestones.
> Halte dich strikt an den **Scope** (keine Auth, keine Tests).

