package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"parkhaus-2/internal/db"
	"parkhaus-2/internal/parkhaus/model"
	"parkhaus-2/internal/server"
)

// router ist der geteilte Gin-Router gegen die echte Test-Datenbank.
// Er wird einmalig in TestMain aufgebaut, da alle GET-Tests rein lesend
// arbeiten und sich deshalb nicht gegenseitig beeinflussen können.
var router http.Handler

// initSQLDir ist der Pfad zu den vorhandenen DDL-/Seed-Skripten (relativ zu dieser Datei).
const initSQLDir = "../../../extras/compose/postgres/init/parkhaus"

// TestMain startet einmalig einen PostgreSQL-Container, spielt die statischen
// CSV-Seed-Daten via der vorhandenen Init-Skripte ein und baut den Router auf.
func TestMain(m *testing.M) {
	// decimal.Decimal als JSON-Zahl (wie in der Produktion via main.go).
	model.InitDecimalMarshal()

	ctx := context.Background()

	container, database, err := startPostgres(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Setup fehlgeschlagen: %v\n", err)
		os.Exit(1)
	}

	router = server.New(database)

	code := m.Run()

	// Aufräumen (Fehler hier sind für das Testergebnis unkritisch).
	_ = container.Terminate(context.Background())
	db.Close(database)

	os.Exit(code)
}

// startPostgres fährt den Container hoch und liefert eine GORM-Verbindung.
// Die DDL-Skripte (01-create.sql, 02-copy-csv.sql) werden in alphabetischer
// Reihenfolge ausgeführt; die CSVs werden zusätzlich an den vom COPY-Befehl
// erwarteten Pfad /docker-entrypoint-initdb.d/csv gemountet.
func startPostgres(ctx context.Context) (*postgres.PostgresContainer, *gorm.DB, error) {
	createSQL, err := filepath.Abs(filepath.Join(initSQLDir, "01-create.sql"))
	if err != nil {
		return nil, nil, err
	}
	copySQL, err := filepath.Abs(filepath.Join(initSQLDir, "02-copy-csv.sql"))
	if err != nil {
		return nil, nil, err
	}

	// Die CSVs werden einzeln mit world-readable FileMode gemountet, damit der
	// postgres-Prozess sie beim COPY lesen kann (Verzeichnis-Mounts erben die
	// Leserechte nicht zuverlässig -> sonst "Permission denied").
	csvFiles, err := csvContainerFiles()
	if err != nil {
		return nil, nil, err
	}

	container, err := postgres.Run(ctx,
		"postgres:16",
		postgres.WithDatabase("parkhaus"),
		postgres.WithUsername("parkhaus"),
		postgres.WithPassword("parkhaus"),
		postgres.WithOrderedInitScripts(createSQL, copySQL),
		// CSVs an den Pfad mounten, den 02-copy-csv.sql erwartet.
		testcontainers.WithFiles(csvFiles...),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("container start: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable", "search_path=parkhaus")
	if err != nil {
		return nil, nil, fmt.Errorf("connection string: %w", err)
	}

	database, err := db.Connect(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("db connect: %w", err)
	}
	// GORM-Logger im Test stummschalten: Die 404-Tests lösen bewusst
	// "record not found" aus; diese Warnungen sollen den Test-Output nicht verrauschen.
	database.Logger = gormlogger.Default.LogMode(gormlogger.Silent)
	return container, database, nil
}

// csvContainerFiles liefert die einzelnen CSV-Seed-Dateien als ContainerFiles,
// gemountet unter /docker-entrypoint-initdb.d/csv mit world-readable Rechten.
func csvContainerFiles() ([]testcontainers.ContainerFile, error) {
	names := []string{"parkhaus.csv", "adresse.csv", "auto.csv"}
	files := make([]testcontainers.ContainerFile, 0, len(names))
	for _, name := range names {
		host, err := filepath.Abs(filepath.Join(initSQLDir, "csv", name))
		if err != nil {
			return nil, err
		}
		files = append(files, testcontainers.ContainerFile{
			HostFilePath:      host,
			ContainerFilePath: "/docker-entrypoint-initdb.d/csv/" + name,
			FileMode:          0o644,
		})
	}
	return files, nil
}

// doGET führt einen GET-Request gegen den geteilten Router aus.
func doGET(t *testing.T, target string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// --- GET /rest/:id ---

func TestFindByID_Erfolg(t *testing.T) {
	rec := doGET(t, "/rest/1", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	if etag := rec.Header().Get("ETag"); etag != `"0"` {
		t.Errorf("ETag = %q, erwartet %q", etag, `"0"`)
	}

	var dto model.ParkhausDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if dto.ID != 1 {
		t.Errorf("id = %d, erwartet 1", dto.ID)
	}
	if dto.Name != "Parkhaus Aachen" {
		t.Errorf("name = %q, erwartet %q", dto.Name, "Parkhaus Aachen")
	}
	if dto.Kapazitaet != 3 {
		t.Errorf("kapazitaet = %d, erwartet 3", dto.Kapazitaet)
	}
	if got := dto.TarifProStunde.StringFixed(2); got != "2.50" {
		t.Errorf("tarifProStunde = %s, erwartet 2.50", got)
	}
	if dto.Adresse == nil {
		t.Fatal("adresse fehlt, erwartet eingebettete Adresse")
	}
	if dto.Adresse.Ort != "Aachen" {
		t.Errorf("adresse.ort = %q, erwartet %q", dto.Adresse.Ort, "Aachen")
	}
}

func TestFindByID_IfNoneMatch_NotModified(t *testing.T) {
	rec := doGET(t, "/rest/1", map[string]string{"If-None-Match": `"0"`})

	if rec.Code != http.StatusNotModified {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotModified)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("Body sollte leer sein, war %d Bytes", rec.Body.Len())
	}
}

func TestFindByID_UnbekannteID_NotFound(t *testing.T) {
	rec := doGET(t, "/rest/999999", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

func TestFindByID_NichtNumerischeID_NotFound(t *testing.T) {
	rec := doGET(t, "/rest/abc", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

func TestFindByID_InkompatiblerAccept_NotAcceptable(t *testing.T) {
	rec := doGET(t, "/rest/1", map[string]string{"Accept": "application/xml"})

	if rec.Code != http.StatusNotAcceptable {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotAcceptable)
	}
}

// --- GET /rest (Suche / Liste / Count) ---

func TestSearch_Liste_DefaultPagination(t *testing.T) {
	rec := doGET(t, "/rest", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}

	var page model.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if page.Page.TotalElements != 6 {
		t.Errorf("totalElements = %d, erwartet 6", page.Page.TotalElements)
	}
	if page.Page.Size != model.DefaultPageSize {
		t.Errorf("size = %d, erwartet %d", page.Page.Size, model.DefaultPageSize)
	}
	if page.Page.Number != 0 {
		t.Errorf("number = %d, erwartet 0", page.Page.Number)
	}
	// Default size = 5 -> ceil(6/5) = 2 Seiten, erste Seite hat 5 Einträge.
	if page.Page.TotalPages != 2 {
		t.Errorf("totalPages = %d, erwartet 2", page.Page.TotalPages)
	}
	if len(page.Content) != model.DefaultPageSize {
		t.Errorf("content-Länge = %d, erwartet %d", len(page.Content), model.DefaultPageSize)
	}
}

func TestSearch_ZweiteSeite(t *testing.T) {
	// API-page ist 1-basiert: page=2 -> interne number 1, Rest von 6 Einträgen = 1.
	rec := doGET(t, "/rest?page=2&size=5", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	var page model.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if page.Page.Number != 1 {
		t.Errorf("number = %d, erwartet 1", page.Page.Number)
	}
	if len(page.Content) != 1 {
		t.Errorf("content-Länge = %d, erwartet 1", len(page.Content))
	}
}

func TestSearch_NameTeilstring(t *testing.T) {
	// Teilstring-Suche: "Berlin" findet "Parkhaus Berlin".
	rec := doGET(t, "/rest?name=Berlin", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	var page model.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if page.Page.TotalElements != 1 {
		t.Fatalf("totalElements = %d, erwartet 1", page.Page.TotalElements)
	}
	if page.Content[0].Name != "Parkhaus Berlin" {
		t.Errorf("name = %q, erwartet %q", page.Content[0].Name, "Parkhaus Berlin")
	}
}

func TestSearch_KapazitaetLte(t *testing.T) {
	// kapazitaet<=3 -> nur "Parkhaus Aachen" (Kapazität 3).
	rec := doGET(t, "/rest?kapazitaet=3", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	var page model.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if page.Page.TotalElements != 1 {
		t.Errorf("totalElements = %d, erwartet 1", page.Page.TotalElements)
	}
}

func TestSearch_TarifProStundeLte(t *testing.T) {
	// tarifProStunde<=3.00 -> Aachen (2.50), Berlin (3.00), Köln (3.00) = 3 Treffer.
	rec := doGET(t, "/rest?tarifProStunde=3.00", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	var page model.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if page.Page.TotalElements != 3 {
		t.Errorf("totalElements = %d, erwartet 3", page.Page.TotalElements)
	}
}

func TestSearch_CountOnly(t *testing.T) {
	rec := doGET(t, "/rest?count-only", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Count int64 `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Body nicht parsebar: %v", err)
	}
	if body.Count != 6 {
		t.Errorf("count = %d, erwartet 6", body.Count)
	}
}

func TestSearch_UnbekannterParameter_NotFound(t *testing.T) {
	rec := doGET(t, "/rest?foo=bar", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

func TestSearch_KeineTreffer_NotFound(t *testing.T) {
	rec := doGET(t, "/rest?name=GibtEsNicht", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

func TestSearch_InkompatiblerAccept_NotAcceptable(t *testing.T) {
	rec := doGET(t, "/rest", map[string]string{"Accept": "application/xml"})

	if rec.Code != http.StatusNotAcceptable {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotAcceptable)
	}
}

// --- GET /rest/file/:id ---

func TestDownloadFile_KeineDatei_NotFound(t *testing.T) {
	// Der CSV-Seed enthält keine Dateien -> 404.
	rec := doGET(t, "/rest/file/1", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

func TestDownloadFile_NichtNumerischeID_NotFound(t *testing.T) {
	rec := doGET(t, "/rest/file/abc", nil)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Status = %d, erwartet %d", rec.Code, http.StatusNotFound)
	}
}

// --- Health ---

func TestHealth(t *testing.T) {
	for _, path := range []string{"/health/liveness", "/health/readiness"} {
		rec := doGET(t, path, nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: Status = %d, erwartet %d", path, rec.Code, http.StatusOK)
		}
		var body struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: Body nicht parsebar: %v", path, err)
		}
		if body.Status != "up" {
			t.Errorf("%s: status = %q, erwartet %q", path, body.Status, "up")
		}
	}
}
