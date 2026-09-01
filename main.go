package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Fehler:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return nil
	case "install":
		if len(args) != 1 {
			return fmt.Errorf("Verwendung: fileopener install")
		}
		if err := installProtocol(); err != nil {
			return err
		}
		fmt.Printf("Schema %s:// wurde fuer diesen Benutzer registriert.\n", urlScheme)
		return nil
	case "uninstall":
		if len(args) != 1 {
			return fmt.Errorf("Verwendung: fileopener uninstall")
		}
		if err := unregisterProtocol(); err != nil {
			return err
		}
		fmt.Printf("Registrierung fuer %s:// wurde entfernt.\n", urlScheme)
		return nil
	case "set":
		if len(args) != 3 {
			return fmt.Errorf("Verwendung: fileopener set <Alias> <Ordner>")
		}
		cfg, err := loadConfig(path)
		if err != nil {
			return err
		}
		folder, err := setFolder(&cfg, args[1], args[2])
		if err != nil {
			return err
		}
		if err := saveConfig(path, cfg); err != nil {
			return err
		}
		fmt.Printf("%s -> %s\n", args[1], folder)
		return nil
	case "remove":
		if len(args) != 2 {
			return fmt.Errorf("Verwendung: fileopener remove <Alias>")
		}
		cfg, err := loadConfig(path)
		if err != nil {
			return err
		}
		if err := removeFolder(&cfg, args[1]); err != nil {
			return err
		}
		return saveConfig(path, cfg)
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("Verwendung: fileopener list")
		}
		cfg, err := loadConfig(path)
		if err != nil {
			return err
		}
		for _, alias := range sortedAliases(cfg) {
			fmt.Printf("%s -> %s\n", alias, cfg.Folders[alias])
		}
		return nil
	case "open":
		if len(args) != 2 {
			return fmt.Errorf("Verwendung: fileopener open <%s://Alias>", urlScheme)
		}
		return openConfiguredURL(path, args[1])
	default:
		if len(args) == 1 {
			return openConfiguredURL(path, args[0])
		}
		return fmt.Errorf("unbekannter Befehl %q", args[0])
	}
}

func openConfiguredURL(configFile, rawURL string) error {
	alias, err := aliasFromURL(rawURL)
	if err != nil {
		return err
	}
	cfg, err := loadConfig(configFile)
	if err != nil {
		return err
	}
	folder, err := folderForAlias(cfg, alias)
	if err != nil {
		return err
	}
	return openFolder(folder)
}

func printUsage() {
	fmt.Printf(`Tualo File Opener

Verwendung:
  fileopener install
  fileopener set <Alias> <Ordner>
  fileopener list
  fileopener open <%s://Alias>
  fileopener remove <Alias>
  fileopener uninstall
`, urlScheme)
}
