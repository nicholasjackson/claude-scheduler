package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"claude-schedule/internal/db"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("cannot find config directory: %v", err)
	}
	dbPath := filepath.Join(configDir, "claude-schedule", "claude-schedule.db")

	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}

	notifier := notifications.New()
	appService := NewApp(store, notifier)

	app := application.New(application.Options{
		Name:        "Claude Scheduler",
		Description: "Schedule tasks using Claude Code",
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(notifier),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Claude Scheduler",
		Width:            1024,
		Height:           768,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Hide window instead of destroying it on close so the app keeps running
	// in the systray.
	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	// System tray setup.
	systemTray := app.SystemTray.New()
	systemTray.SetTooltip("Claude Scheduler")

	if runtime.GOOS == "darwin" {
		systemTray.SetTemplateIcon(icons.SystrayMacTemplate)
	} else {
		systemTray.SetIcon(icons.SystrayLight)
		systemTray.SetDarkModeIcon(icons.SystrayDark)
	}

	menu := app.NewMenu()
	menu.Add("Open Claude Scheduler").OnClick(func(ctx *application.Context) {
		window.Show()
		window.Focus()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	systemTray.SetMenu(menu)

	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
