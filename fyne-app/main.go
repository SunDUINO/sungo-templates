/*
 * Project: ${projectName}
 * Template: Fyne Desktop App – SunGo Project Manager
 *
 * Features:
 *   - Main window with toolbar (New, Open, Save, About)
 *   - Tab container (Editor tab + Log tab)
 *   - Status bar with clock
 *   - Theme toggle (Light / Dark)
 *   - About dialog
 */

package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ── App state ─────────────────────────────────────────────────────────────────

type AppState struct {
	app       fyne.App
	win       fyne.Window
	editor    *widget.Entry
	logOutput *widget.Label
	status    *widget.Label
	darkMode  bool
}

// ── UI builders ───────────────────────────────────────────────────────────────

func (s *AppState) buildToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			s.editor.SetText("")
			s.log("New document created")
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			dialog.ShowInformation("Open", "File picker goes here", s.win)
			s.log("Open dialog triggered")
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			s.log(fmt.Sprintf("Saved (%d chars)", len(s.editor.Text)))
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ColorPaletteIcon(), func() {
			s.toggleTheme()
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			s.showAbout()
		}),
	)
}

func (s *AppState) buildEditorTab() *container.TabItem {
	s.editor = widget.NewMultiLineEntry()
	s.editor.SetPlaceHolder("Start typing here...")
	s.editor.Wrapping = fyne.TextWrapWord
	return container.NewTabItemWithIcon("Editor", theme.DocumentIcon(), s.editor)
}

func (s *AppState) buildLogTab() *container.TabItem {
	s.logOutput = widget.NewLabel("Application started.\n")
	s.logOutput.Wrapping = fyne.TextWrapWord
	scroll := container.NewScroll(s.logOutput)
	return container.NewTabItemWithIcon("Log", theme.ListIcon(), scroll)
}

func (s *AppState) buildStatusBar() *fyne.Container {
	s.status = widget.NewLabel("Ready")
	clock  := widget.NewLabel("")

	// Update clock every second
	go func() {
		for range time.Tick(time.Second) {
			clock.SetText(time.Now().Format("15:04:05"))
		}
	}()

	return container.NewBorder(nil, nil, s.status, clock)
}

// ── Actions ───────────────────────────────────────────────────────────────────

func (s *AppState) log(msg string) {
	ts  := time.Now().Format("15:04:05")
	cur := s.logOutput.Text
	s.logOutput.SetText(fmt.Sprintf("%s[%s] %s\n", cur, ts, msg))
	s.status.SetText(msg)
}

func (s *AppState) toggleTheme() {
	s.darkMode = !s.darkMode
	if s.darkMode {
		s.app.Settings().SetTheme(theme.DarkTheme())
		s.log("Switched to Dark theme")
	} else {
		s.app.Settings().SetTheme(theme.LightTheme())
		s.log("Switched to Light theme")
	}
}

func (s *AppState) showAbout() {
	dialog.ShowInformation(
		"About ${projectName}",
		"${projectName}\nBuilt with Fyne + SunGo Project Manager\n\nfyne.io",
		s.win,
	)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	a := app.NewWithID("pl.lothar-team.${projectName}")
	w := a.NewWindow("${projectName}")
	w.Resize(fyne.NewSize(900, 600))
	w.SetMaster()

	state := &AppState{app: a, win: w, darkMode: true}

	toolbar   := state.buildToolbar()
	tabs      := container.NewAppTabs(state.buildEditorTab(), state.buildLogTab())
	statusBar := state.buildStatusBar()

	content := container.NewBorder(toolbar, statusBar, nil, nil, tabs)
	w.SetContent(content)

	state.log("${projectName} ready")
	w.ShowAndRun()
}
