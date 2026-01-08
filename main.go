package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	frameless := true
	var macOptions *mac.Options
	if runtime.GOOS == "darwin" {
		// Use native window chrome to get true rounded corners on macOS.
		frameless = false
		macOptions = &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
		}
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ksight",
		Width:  1024,
		Height: 768,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 255},
			Mac:             macOptions,
			OnStartup:       app.startup,
			Frameless:       frameless,
			CSSDragProperty: "--wails-draggable",
			CSSDragValue:    "drag",
			Bind: []interface{}{
				app,
			},
		})

	if err != nil {
		println("Error:", err.Error())
	}
}
