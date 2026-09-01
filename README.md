# Tualo File Opener

Ein kleines Go-Programm, das Links wie `tualo-fs://OrdnerA` auf einen zuvor
konfigurierten lokalen Ordner abbildet. Es unterstuetzt macOS, Windows und Linux.
Das App-Icon kombiniert einen geoeffneten Ordner mit dem Tualo-Logo und wird bei
der Installation automatisch fuer die jeweilige Plattform eingerichtet.

Direkte Dateipfade aus Links werden absichtlich nicht akzeptiert. Nur lokale,
vom Benutzer eingerichtete Aliase koennen geoeffnet werden.

## Bauen und einrichten

Go 1.22 oder neuer wird benoetigt.

```sh
go build -o fileopener .
./fileopener install
./fileopener set OrdnerA "/lokaler/pfad/zu/OrdnerA"
```

Ohne weitere Angabe wird `tualo-fs` registriert. Ein eigenes Schema kann beim
Installieren optional angegeben werden:

```sh
./fileopener install meine-ordner
```

Der Browser-Link lautet dann `meine-ordner://OrdnerA`. Um das Schema spaeter zu
aendern, wird `install` einfach erneut mit dem neuen Namen ausgefuehrt. Erlaubt
sind RFC-konforme Schemas, die mit einem Buchstaben beginnen und danach
Buchstaben, Ziffern sowie `+`, `-` oder `.` enthalten.

Unter Windows werden die letzten beiden Befehle mit `fileopener.exe` ausgefuehrt.
`install` registriert das Schema nur fuer den aktuellen Benutzer und benoetigt
keine Administratorrechte. Nach einem Update muss `fileopener.exe install`
erneut ausgefuehrt werden. Die App erscheint danach unter **Einstellungen >
Apps > Standard-Apps** als `Tualo File Opener` und als Auswahl fuer den
registrierten Linktyp.

Danach kann folgender Link im Browser aufgerufen werden:

```text
tualo-fs://OrdnerA
tualo-fs://OrdnerA/Unterordner/Projekt%20mit%20Leerzeichen
```

Alles nach dem Alias wird als Unterpfad innerhalb des konfigurierten Ordners
geoeffnet. Pfade ausserhalb dieses Stammordners und Symlink-Ausbrueche werden
abgelehnt. Kann ein Link nicht geoeffnet werden, zeigt die Anwendung eine native
Fehlermeldung an; die Meldung wird zusaetzlich auf stderr geschrieben.

Der Browser fragt beim ersten Aufruf normalerweise nach einer Bestaetigung, dass
die externe Anwendung geoeffnet werden darf.

## Befehle

```text
fileopener install [Schema]
fileopener set <Alias> <Ordner>
fileopener list
fileopener open <Schema://Alias>
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

## App-Icon

Die bearbeitbare Vorlage liegt unter `assets/app-icon.svg`. Daraus wurden
`app-icon.png` fuer Linux, `app-icon.icns` fuer macOS und `app-icon.ico` fuer
Windows erzeugt. Der Windows-amd64-Build bindet das Icon ueber
`rsrc_windows_amd64.syso` automatisch auch in die EXE-Datei ein.