.PHONY: run fmt lint security tidy test-unit test-integration test all help

# Pfad zum lokalen Linter
LOCAL_LINT_BIN := ./bin/golangci-lint


all: fmt tidy lint security run

# Startet den Gin-Server
run:
	@echo "🚀 Starte den REST-API Server..."
	go run cmd/server/main.go

# Formatiert den gesamten Go-Code nach dem offiziellen Standard
fmt:
	@echo "🎨 Formatiere Go-Code (go fmt)..."
	go fmt ./...

# Bereinigt und aktualisiert die go.mod Abhängigkeiten
tidy:
	@echo "📦 Bereinige Go-Module (go mod tidy)..."
	go mod tidy

# Führt den Linter aus (prüft flexibel lokal oder global)
lint:
	@echo "🔍 Starte Code-Analyse (golangci-lint)..."
	@if [ -f $(LOCAL_LINT_BIN) ]; then \
		$(LOCAL_LINT_BIN) run ./...; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "❌ Fehler: golangci-lint wurde weder im ./bin-Ordner noch global im PATH gefunden!"; \
		exit 1; \
	fi

# Führt den offiziellen Sicherheits-Scanner aus
security:
	@echo "🛡️ Prüfe Abhängigkeiten auf Sicherheitslücken (govulncheck)..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "💡 Tipp: Installiere govulncheck für Security-Scans mit:"; \
		echo "   go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Führt die reinen Unit-Tests (Service-Layer, kein Server-Start, keine DB) aus
test-unit:
	@echo "🧪 Führe Unit-Tests aus (Service-Layer)..."
	go test -count=1 -short ./internal/parkhaus/service/

# Führt die Integrationstests (Handler, benötigt PostgreSQL-Container) aus
test-integration:
	@echo "🧪 Führe Integrationstests aus (Handler, benötigt Docker)..."
	go test -count=1 ./internal/parkhaus/handler/

# Führt zuerst Unit-Tests, dann Integrationstests aus
test: test-unit test-integration

# Zeigt eine Übersicht aller verfügbaren Befehle
help:
	@echo "Verfügbare Makefile-Befehle:"
	@echo "  make run      - Startet die REST-API (cmd/server/main.go)"
	@echo "  make fmt      - Formatiert alle .go Dateien"
	@echo "  make lint     - Führt die statische Code-Analyse aus"
	@echo "  make security - Scannt das Projekt auf bekannte Sicherheitslücken"
	@echo "  make tidy     - Räumt ungenutzte Dependencies in go.mod auf"
	@echo "  make test-unit        - Führt nur die Unit-Tests aus (schnell, keine DB)"
	@echo "  make test-integration - Führt nur die Integrationstests aus (braucht Docker)"
	@echo "  make test             - Führt Unit- und Integrationstests nacheinander aus"
	@echo "  make all      - Führt fmt, tidy, lint, security aus und startet dann den Server"
