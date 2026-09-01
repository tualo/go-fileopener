# Tualo File Opener

Ein kleines Go-Programm, das Links wie `tualo-fs://OrdnerA` auf einen zuvor
konfigurierten lokalen Ordner abbildet. Es unterstuetzt macOS, Windows und Linux.

Direkte Dateipfade aus Links werden absichtlich nicht akzeptiert. Nur lokale,
vom Benutzer eingerichtete Aliase koennen geoeffnet werden.

## Bauen und einrichten

Go 1.22 oder neuer wird benoetigt.

```sh
go build -o fileopener .
./fileopener install
./fileopener set OrdnerA "/lokaler/pfad/zu/OrdnerA"
```

Unter Windows werden die letzten beiden Befehle mit `fileopener.exe` ausgefuehrt.
`install` registriert das Schema nur fuer den aktuellen Benutzer und benoetigt
keine Administratorrechte.

Danach kann folgender Link im Browser aufgerufen werden:

```text
tualo-fs://OrdnerA
```

Der Browser fragt beim ersten Aufruf normalerweise nach einer Bestaetigung, dass
die externe Anwendung geoeffnet werden darf.

## Befehle

```text
fileopener install
fileopener set <Alias> <Ordner>
fileopener list
fileopener open <tualo-fs://Alias>
fileopener remove <Alias>
fileopener uninstall
```

Aliase werden ohne Beachtung der Gross-/Kleinschreibung gefunden. `uninstall`
entfernt die Schema-Registrierung, behaelt aber die Alias-Konfiguration bei.

## Builds fuer alle Plattformen

```sh
mkdir -p dist/darwin/arm64
mkdir -p dist/darwin/amd64
mkdir -p dist/windows/amd64
mkdir -p dist/linux/amd64

GOOS=darwin GOARCH=arm64 go build -o dist/darwin/arm64/fileopener .
GOOS=darwin GOARCH=amd64 go build -o dist/darwin/amd64/fileopener .
GOOS=windows GOARCH=amd64 go build -o dist/windows/amd64/fileopener.exe .
GOOS=linux GOARCH=amd64 go build -o dist/linux/amd64/fileopener .

```

Unter Linux muessen `xdg-open` und `xdg-mime` vorhanden sein. Die meisten
Desktop-Distributionen liefern beide ueber `xdg-utils` aus.