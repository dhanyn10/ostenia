package main

import (
	"embed"
	"os"
	"ostenia/internal/backend"
	"ostenia/internal/network"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check for admin helper command
	if len(os.Args) >= 4 && os.Args[1] == "--add-host" {
		ip := os.Args[2]
		hostname := os.Args[3]
		err := network.AddHost(ip, hostname)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Create an instance of the app structure
	app := backend.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ostenia",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Frameless:        true,
		EnableFileDrop:   true,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
