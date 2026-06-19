#!/usr/bin/env bash
# Manuelle Smoke-Tests gegen die laufende API (http://localhost:8080).
set -u
BASE="http://localhost:8080"

echo "=== A. Decimal als Zahl (erwartet 2.5 ohne Quotes) ==="
curl -s "$BASE/rest/1" | grep -o '"tarifProStunde":[^,]*'

echo "=== B. POST neu (erwartet 201 + Location) ==="
curl -s -i -X POST "$BASE/rest" -H 'Content-Type: application/json' \
  -d '{"name":"Parkhaus Test","kapazitaet":10,"tarifProStunde":2.5,"adresse":{"plz":"68159","ort":"Mannheim","strasse":"Hauptstr","hausnummer":"1"}}' \
  | grep -iE 'HTTP/|Location'

echo "=== C. POST Validierungsfehler (erwartet 422 + Problem Details mit path) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" -X POST "$BASE/rest" -H 'Content-Type: application/json' \
  -d '{"name":"","kapazitaet":0,"tarifProStunde":-1,"adresse":{"plz":"","ort":"","strasse":"","hausnummer":""}}'
curl -s -X POST "$BASE/rest" -H 'Content-Type: application/json' \
  -d '{"name":"","kapazitaet":0,"tarifProStunde":-1,"adresse":{"plz":"","ort":"","strasse":"","hausnummer":""}}' | head -c 400
echo ""

echo "=== D. POST Name-Duplikat (erwartet 422) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" -X POST "$BASE/rest" -H 'Content-Type: application/json' \
  -d '{"name":"Parkhaus Aachen","kapazitaet":10,"tarifProStunde":2.5,"adresse":{"plz":"52064","ort":"Aachen","strasse":"X","hausnummer":"1"}}'

echo "=== E. Suche unbekannter Param (erwartet 404) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" "$BASE/rest?foo=bar"

echo "=== F. Suche kapazitaet<=3 (lte) ==="
curl -s "$BASE/rest?kapazitaet=3" | grep -o '"totalElements":[0-9]*'

echo "=== G. PUT ohne If-Match (erwartet 428) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" -X PUT "$BASE/rest/1" -H 'Content-Type: application/json' \
  -d '{"name":"Parkhaus Aachen","kapazitaet":5,"tarifProStunde":3.0}'

echo "=== H. PUT mit ungueltigem If-Match (erwartet 412) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" -X PUT "$BASE/rest/1" -H 'Content-Type: application/json' -H 'If-Match: "abc"' \
  -d '{"name":"Parkhaus Aachen","kapazitaet":5,"tarifProStunde":3.0}'

echo "=== I. PUT mit korrektem If-Match \"0\" (erwartet 204 + ETag \"1\") ==="
curl -s -i -X PUT "$BASE/rest/1" -H 'Content-Type: application/json' -H 'If-Match: "0"' \
  -d '{"name":"Parkhaus Aachen","kapazitaet":5,"tarifProStunde":3.0}' | grep -iE 'HTTP/|Etag'

echo "=== J. GET mit If-None-Match (erwartet 304) ==="
ETAG=$(curl -s -i "$BASE/rest/2" | grep -i '^Etag:' | tr -d '\r' | awk '{print $2}')
echo "aktuelles ETag von /rest/2: $ETAG"
curl -s -o /dev/null -w "status=%{http_code}\n" "$BASE/rest/2" -H "If-None-Match: $ETAG"

echo "=== K. POST Auto hinzufuegen (erwartet 201 + Location) ==="
curl -s -i -X POST "$BASE/rest/2/autos" -H 'Content-Type: application/json' \
  -d '{"kennzeichen":"MA-XX-1","einfahrtszeit":"2025-06-19T10:00:00Z","kundentyp":"PREMIUM"}' | grep -iE 'HTTP/|Location'

echo "=== L. DELETE nicht-numerische ID (erwartet 204) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" -X DELETE "$BASE/rest/abc"

echo "=== M. Accept inkompatibel (erwartet 406) ==="
curl -s -o /dev/null -w "status=%{http_code}\n" "$BASE/rest/1" -H 'Accept: application/xml'

