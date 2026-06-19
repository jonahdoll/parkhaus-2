-- Seeding aus den CSV-Dateien (Semikolon-getrennt, mit Header).
-- Die CSVs liegen im selben Init-Verzeichnis und werden vom Postgres-Container
-- unter /docker-entrypoint-initdb.d/csv bereitgestellt.
SET search_path TO parkhaus;

-- parkhaus: feste IDs aus CSV übernehmen (OVERRIDING SYSTEM VALUE wegen GENERATED ALWAYS).
COPY parkhaus (id, version, name, kapazitaet, tarif_pro_stunde, erzeugt, aktualisiert)
    FROM '/docker-entrypoint-initdb.d/csv/parkhaus.csv'
    WITH (FORMAT csv, DELIMITER ';', HEADER true);

COPY adresse (id, plz, ort, strasse, hausnummer, parkhaus_id)
    FROM '/docker-entrypoint-initdb.d/csv/adresse.csv'
    WITH (FORMAT csv, DELIMITER ';', HEADER true);

COPY auto (id, kennzeichen, einfahrtszeit, kundentyp, parkhaus_id)
    FROM '/docker-entrypoint-initdb.d/csv/auto.csv'
    WITH (FORMAT csv, DELIMITER ';', HEADER true);

-- Identity-Sequenzen nach dem Import anpassen, damit neue Datensätze
-- kollisionsfrei vergeben werden. parkhaus startet gemäß Spezifikation bei >= 1000.
SELECT setval(pg_get_serial_sequence('parkhaus', 'id'), GREATEST(1000, (SELECT MAX(id) FROM parkhaus)));
SELECT setval(pg_get_serial_sequence('adresse', 'id'), (SELECT MAX(id) FROM adresse));
SELECT setval(pg_get_serial_sequence('auto', 'id'), (SELECT MAX(id) FROM auto));

