//go:build windows

package win32

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ntars/Desop/internal/app"
	"github.com/ntars/Desop/internal/config"
	"github.com/ntars/Desop/internal/dock"
	"github.com/ntars/Desop/internal/geometry"
	"github.com/ntars/Desop/internal/island"
	"github.com/ntars/Desop/internal/tokens"
	"github.com/ntars/Desop/internal/topbar"
	"github.com/ntars/Desop/internal/widgets"
)

const (
	timerRefresh = 1
	timerIsland  = 2
	refreshEvery = 8

	cmdSettings      = 1001
	cmdEnableDock    = 1010
	cmdEnableTopBar  = 1011
	cmdEnableWidgets = 1012
	cmdEnableIsland  = 1013
	cmdDockItemBase  = 2000
	cmdWidgetBase    = 3000
	cmdRestart       = 1019
	cmdAutostart     = 1020
	cmdExit          = 1021

	className = "DesopDesktopWindow"
)

// DesktopWindow owns the message-only host, tray icon, all component windows
// and the loopback island API. Everything is created in user mode.
type DesktopWindow struct {
	hwnd        HWND
	cfg         config.Config
	cfgPath     string
	instance    HINSTANCE
	reader      *StatusReader
	island      *island.Controller
	api         *island.Server
	hDock       HWND
	islandRect  geometry.Rect
	hTop        HWND
	hWidgets    []HWND
	hIsland     HWND
	hNoteEdit   HWND
	lastCursor  POINT
	monitor     Monitors
	drag        *widgetDrag
	closing     bool
	restarting  bool
	closed      bool
	tick        uint64
	currentPID  uint32
	launchTimes map[string]time.Time
	stopEvent   HANDLE
	stopName    string
	wndProc     uintptr
	dirty       map[HWND]struct{}
	frozen      bool
	paintFrames uint64
	topVisible  bool
	dockVisible bool
}

type widgetDrag struct {
	index  int
	offset geometry.Point
}

// NewDesktopWindow registers the host window class and creates the tray host.
func NewDesktopWindow(cfg config.Config, cfgPath string) (*DesktopWindow, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	instance := GetModuleHandle("")
	d := &DesktopWindow{
		cfg: cfg, cfgPath: cfgPath, instance: instance,
		reader: NewStatusReader(), island: island.New(cfg.Island.Expanded, nil),
		launchTimes: make(map[string]time.Time),
		currentPID:  CurrentProcessID(),
	}
	d.wndProc = syscall.NewCallback(d.wndProcFunc)
	runtime.KeepAlive(d.wndProc)
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   d.wndProc,
		HInstance:     instance,
		HbrBackground: 0,
		LpszClassName: UTF16(className),
	}
	stopEvent, err := CreateStopEvent(app.StopEventName(cfgPath))
	if err != nil {
		return nil, fmt.Errorf("create stop event: %w", err)
	}
	d.stopEvent = stopEvent
	d.stopName = app.StopEventName(cfgPath)
	_, err = RegisterClassEx(&wc)
	if err != nil && !errors.Is(err, syscall.Errno(1410)) { // ERROR_CLASS_ALREADY_EXISTS
		return nil, fmt.Errorf("register window class: %w", err)
	}
	for i := 0; i < 4; i++ {
		for _, name := range []string{componentClass(i), widgetClass(i)} {
			class := WNDCLASSEXW{
				CbSize: uint32(unsafe.Sizeof(WNDCLASSEXW{})), LpfnWndProc: d.wndProc,
				HInstance: instance, HbrBackground: 0, LpszClassName: UTF16(name),
			}
			_, err = RegisterClassEx(&class)
			if err != nil && !errors.Is(err, syscall.Errno(1410)) {
				return nil, fmt.Errorf("register component class %s: %w", name, err)
			}
		}
	}
	for _, name := range []string{dockClass(), islandClass()} {
		class := WNDCLASSEXW{
			CbSize: uint32(unsafe.Sizeof(WNDCLASSEXW{})), LpfnWndProc: d.wndProc,
			HInstance: instance, HbrBackground: 0, LpszClassName: UTF16(name),
		}
		_, err = RegisterClassEx(&class)
		if err != nil && !errors.Is(err, syscall.Errno(1410)) {
			return nil, fmt.Errorf("register component class %s: %w", name, err)
		}
	}
	d.hwnd, err = CreateWindowEx(WS_EX_TOOLWINDOW, className, "Desop", 0, 0, 0, 0, 0, 0, 0, instance, 0)
	if err != nil {
		return nil, fmt.Errorf("create host window: %w", err)
	}
	SetWindowLongPtr(d.hwnd, GWLP_USERDATA, uintptr(unsafe.Pointer(d)))
	return d, nil
}

// Start creates every enabled component and the optional local API.
func (d *DesktopWindow) Start() error {
	monitors := EnumMonitors()
	if len(monitors) == 0 {
		return errors.New("no display detected")
	}
	for _, m := range monitors {
		if m.Primary {
			d.monitor = m
			break
		}
	}
	if d.monitor.Screen.Right == 0 {
		d.monitor = monitors[0]
	}
	// A tray icon is part of the normal interactive experience, but headless
	// sessions and some automation contexts cannot create one. The desktop
	// components and selftest must still run in those environments.
	if err := NotifyIcon(d.hwnd, NIM_ADD, "Desop - right-click for settings"); err != nil {
		_ = err
	}
	if d.cfg.Dock.Enabled {
		if err := d.createDock(); err != nil {
			return err
		}
	}
	if d.cfg.TopBar.Enabled {
		if err := d.createTopBar(); err != nil {
			return err
		}
	}
	if d.cfg.Widgets.Enabled {
		if err := d.createWidgets(); err != nil {
			return err
		}
	}
	if d.cfg.Island.Enabled {
		if err := d.createIsland(); err != nil {
			return err
		}
	}
	if d.cfg.Island.Enabled && d.cfg.Island.LocalAPI {
		api, err := island.NewServer(d.island, d.cfg.Island.LocalAPIPort)
		if err != nil {
			// Port fallback is internal to NewServer; here only true OS errors
			// remain. Keep the desktop shell alive without the optional API.
			_ = NotifyIcon(d.hwnd, NIM_MODIFY, "Desop - local API unavailable; desktop still active")
		} else {
			d.api = api
			go func() { _ = api.Start() }()
		}
	}
	ShowWindow(d.hwnd, SW_HIDE)
	// Component windows were created while hidden by the host; trigger their
	// first paint so the desktop is visible before the first throttled refresh.
	d.dirty = make(map[HWND]struct{})
	d.markFullRefresh()
	if d.hTop != 0 {
		d.topVisible = true
	}
	if d.hDock != 0 {
		d.dockVisible = true
	}
	d.flushRepaints()
	return nil
}

// PumpMessages drains host messages without blocking, then runs one refresh
// tick. This keeps timers, tray commands and island events responsive even if
// an external settings window owns the thread's message queue.
func (d *DesktopWindow) PumpMessages() bool {
	for {
		msg := MSG{}
		got, quit, err := PeekThreadMessage(&msg)
		if err != nil || !got {
			break
		}
		if quit {
			return false
		}
		TranslateMessage(&msg)
		DispatchMessage(&msg)
		if d.closing {
			return false
		}
	}
	if !d.closing && !d.restarting && !d.closed && d.stopEvent != 0 && WaitForStopEvent(d.stopEvent, 0) {
		d.closing = true
		d.Close()
		return false
	}
	if !d.closing && !d.restarting && !d.closed {
		d.refreshDockMotion()
	}
	return !d.closing
}

// RefreshDue runs the low-frequency top bar/widget/island refresh. Keeping it
// outside PumpMessages lets performance selftests measure an idle desktop.
func (d *DesktopWindow) RefreshDue() error {
	if d.closed || d.frozen {
		return nil
	}
	d.tick++
	if d.tick%refreshEvery == 1 {
		d.refresh()
	}
	return nil
}

// FreezeRefresh stops periodic content updates for the rest of the session.
func (d *DesktopWindow) FreezeRefresh() error {
	if d.closed {
		return fmt.Errorf("desktop already closed")
	}
	d.frozen = true
	return nil
}

// PaintFrames returns the number of WM_PAINT frames processed so far.
func (d *DesktopWindow) PaintFrames() uint64 {
	return d.paintFrames
}

// Close removes all windows, timers and tray state. It is idempotent.
func (d *DesktopWindow) Close() {
	if d.closed {
		return
	}
	d.closed = true
	if d.api != nil {
		_ = d.api.Close()
	}
	KillTimer(d.hwnd, timerIsland)
	if d.stopEvent != 0 {
		// Clear and retain the session event so a late `stop` cannot kill the
		// next instance that reuses the same deterministic name.
		_ = ResetStopEvent(d.stopEvent)
		CloseHandle(d.stopEvent)
		d.stopEvent = 0
	}
	_ = NotifyIcon(d.hwnd, NIM_DELETE, "")
	for _, h := range d.widgetWindows() {
		_ = DestroyWindow(h)
	}
	if d.hIsland != 0 {
		_ = DestroyWindow(d.hIsland)
	}
	if d.hTop != 0 {
		_ = DestroyWindow(d.hTop)
	}
	if d.hDock != 0 {
		_ = DestroyWindow(d.hDock)
	}
	_ = DestroyWindow(d.hwnd)
}

func (d *DesktopWindow) widgetWindows() []HWND {
	return append([]HWND(nil), d.hWidgets...)
}

func (d *DesktopWindow) screen() geometry.Rect {
	r := d.monitor.Screen
	return geometry.NewRect(int(r.Left), int(r.Top), int(r.Right), int(r.Bottom))
}

func (d *DesktopWindow) workArea() geometry.Rect {
	r := d.monitor.Work
	return geometry.NewRect(int(r.Left), int(r.Top), int(r.Right), int(r.Bottom))
}

func (d *DesktopWindow) dockWindowRect() geometry.Rect {
	screen := d.screen()
	work := d.workArea()
	tk := tokens.Current()
	size := d.dockIconSize()
	if d.cfg.Dock.Magnification {
		size = int(float64(size) * 1.7)
	}
	thickness := size + tk.RadiusSmall
	var rect geometry.Rect
	switch dock.Position(d.cfg.Dock.Position) {
	case dock.PositionLeft:
		rect = geometry.NewRect(work.X, work.Y, work.X+thickness, work.Bottom())
	case dock.PositionRight:
		rect = geometry.NewRect(work.Right()-thickness, work.Y, work.Right(), work.Bottom())
	default:
		rect = geometry.NewRect(work.X, work.Bottom()-thickness, work.Right(), work.Bottom())
	}
	return geometry.FitInside(rect, screen)
}

// dockIconSize uses the effective DPI of the host window, which Windows
// reports for the monitor owning that window. The dock window is created on
// the primary monitor, so mixed-DPI systems no longer use one fixed size for
// every display.
func (d *DesktopWindow) dockIconSize() int {
	return dock.ScaleIconSize(d.cfg.Dock.IconSize, WindowDPI(d.hwnd))
}

func (d *DesktopWindow) createDock() error {
	rect := d.dockWindowRect()
	h, err := d.createLayeredWindow(dockClass(), "Desop Dock", rect, WS_EX_TOOLWINDOW|WS_EX_NOACTIVATE|WS_EX_LAYERED)
	if err != nil {
		return err
	}
	d.hDock = h
	SetTimer(d.hwnd, timerIsland, 33)
	return nil
}

func (d *DesktopWindow) createTopBar() error {
	opts := topbar.LayoutOptions{Screen: d.screen(), WorkArea: d.workArea(), Height: 32, Padding: 12, Gap: 12}
	rect := topbar.Layout(d.cfg.TopBar, opts).Bar
	h, err := d.createLayeredWindow(componentClass(1), "Desop Top Bar", rect, WS_EX_TOOLWINDOW|WS_EX_NOACTIVATE|WS_EX_LAYERED)
	if err != nil {
		return err
	}
	d.hTop = h
	return nil
}

func (d *DesktopWindow) createWidgets() error {
	sizes := map[string]geometry.Rect{
		"clock":    widgetSize("clock"),
		"status":   widgetSize("status"),
		"notes":    widgetSize("notes"),
		"weather":  widgetSize("weather"),
		"calendar": widgetSize("calendar"),
	}
	for _, placed := range widgets.PlaceAll(d.cfg.Widgets, d.workArea(), sizes) {
		index := widgetIndex(d.cfg.Widgets, placed.ID)
		if index < 0 || !d.cfg.Widgets.Widgets[index].Enabled {
			continue
		}
		h, err := d.createLayeredWindow(widgetClass(index), "Desop Widget", placed.Rect, WS_EX_TOOLWINDOW|WS_EX_NOACTIVATE|WS_EX_LAYERED)
		if err != nil {
			return err
		}
		SetWindowLongPtr(h, GWLP_USERDATA, uintptr(index+1))
		d.hWidgets = append(d.hWidgets, h)
		if placed.Kind == "notes" {
			d.hNoteEdit, err = CreateWindowEx(0, "Edit", "", WS_CHILD|WS_VISIBLE|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_WANTRETURN,
				10, 34, int32(placed.Rect.Width-20), int32(placed.Rect.Height-44), h, 0, d.instance, 0)
			if err != nil {
				return err
			}
			SetWindowText(d.hNoteEdit, d.cfg.Widgets.Note)
		}
	}
	return nil
}

func widgetIndex(cfg config.WidgetsConfig, id string) int {
	for i := range cfg.Widgets {
		if cfg.Widgets[i].ID == id {
			return i
		}
	}
	return -1
}

func (d *DesktopWindow) createIsland() error {
	rect := island.Layout(d.island.Snapshot(), d.screen())
	h, err := d.createLayeredWindow(islandClass(), "Desop Island", rect, WS_EX_TOOLWINDOW|WS_EX_NOACTIVATE|WS_EX_LAYERED|WS_EX_TRANSPARENT)
	if err != nil {
		return err
	}
	d.hIsland = h
	return nil
}

func (d *DesktopWindow) createLayeredWindow(class, title string, rect geometry.Rect, exStyle uint32) (HWND, error) {
	tk := tokens.Current()
	h, err := CreateWindowEx(exStyle, class, title, WS_POPUP|WS_VISIBLE,
		int32(rect.X), int32(rect.Y), int32(rect.Width), int32(rect.Height), 0, 0, d.instance, 0)
	if err != nil {
		return 0, err
	}
	SetWindowLongPtr(h, GWLP_USERDATA, uintptr(unsafe.Pointer(d)))
	_ = DwmSetSystemBackdrop(h)
	if err := SetLayeredWindowAttributes(h, tk.WindowAlpha); err != nil {
		return 0, err
	}
	region := CreateRoundRectRgn(0, 0, int32(rect.Width+1), int32(rect.Height+1), int32(tk.RadiusMedium), int32(tk.RadiusMedium))
	_ = SetWindowRgn(h, region)
	return h, nil
}

func (d *DesktopWindow) wndProcFunc(hwnd HWND, msg uint32, wp, lp uintptr) uintptr {
	if d.closed {
		return DefWindowProc(hwnd, msg, wp, lp)
	}
	switch msg {
	case WM_TIMER:
		switch wp {
		case timerIsland:
			d.refreshDockMotion()
			return 0
		}
	case WM_PAINT:
		delete(d.dirty, hwnd)
		d.paintFrames++
		d.paint(hwnd)
		return 0
	case WM_LBUTTONDOWN:
		d.handleClick(hwnd, lp)
		return 0
	case WM_MOUSEMOVE:
		d.handleMouseMove(hwnd, lp)
		return 0
	case WM_LBUTTONUP:
		d.endWidgetDrag()
		return 0
	case WM_APP_TRAY:
		switch uint32(lp & 0xffffffff) {
		case WM_LBUTTONUP:
			_ = RunSettings(d.cfgPath)
			d.restartShell()
		case WM_RBUTTONUP:
			d.showTrayMenu()
		}
		return 0
	case WM_DESTROY:
		return 0
	}
	return DefWindowProc(hwnd, msg, wp, lp)
}

func (d *DesktopWindow) refresh() {
	fullscreen := d.confirmedFullscreen()
	visible := !topbar.ShouldAutoHide(d.cfg.TopBar, fullscreen)
	if d.hTop != 0 {
		if d.topVisible != visible {
			d.topVisible = visible
			showWindow(d.hTop, visible)
		}
		if visible {
			d.markDirty(d.hTop)
		}
	}
	if d.hDock != 0 {
		dockVisible := visible
		if d.cfg.Dock.AutoHide && !IsCursorOver(d.hDock) {
			dockVisible = false
		}
		if d.dockVisible != dockVisible {
			d.dockVisible = dockVisible
			showWindow(d.hDock, dockVisible)
		}
	}
	if d.cfg.Widgets.Enabled {
		for _, h := range d.widgetWindows() {
			d.markDirty(h)
		}
	}
	if d.cfg.Island.Enabled && d.hIsland != 0 {
		snap := d.island.Snapshot()
		rect := island.Layout(snap, d.screen())
		if rect != d.islandRect {
			d.islandRect = rect
			moveWindow(d.hIsland, rect)
		}
		d.markDirty(d.hIsland)
	}
	d.flushRepaints()
	if d.hNoteEdit != 0 {
		note := windowText(d.hNoteEdit)
		if note != d.cfg.Widgets.Note {
			d.cfg.Widgets.Note = note
			_ = config.Save(d.cfgPath, d.cfg)
		}
	}
}

func (d *DesktopWindow) markDirty(h HWND) {
	if d.dirty == nil {
		d.dirty = make(map[HWND]struct{})
	}
	d.dirty[h] = struct{}{}
}

// markFullRefresh requests the next paint for every enabled component window.
// The dock is motion-driven and is deliberately excluded from periodic repaint.
func (d *DesktopWindow) markFullRefresh() {
	if d.cfg.TopBar.Enabled && d.hTop != 0 {
		d.markDirty(d.hTop)
	}
	if d.cfg.Widgets.Enabled {
		for _, h := range d.widgetWindows() {
			d.markDirty(h)
		}
	}
	if d.cfg.Island.Enabled && d.hIsland != 0 {
		d.markDirty(d.hIsland)
	}
}

// flushRepaints issues at most one invalidation per dirty window. Unpainted
// requests carry into the next tick, so no frames are dropped or duplicated.
func (d *DesktopWindow) flushRepaints() {
	for h := range d.dirty {
		if IsWindowVisible(h) {
			InvalidateRect(h, false)
		}
	}
}

func showWindow(h HWND, visible bool) {
	cmd := int32(SW_HIDE)
	if visible {
		cmd = SW_SHOWNORMAL
	}
	ShowWindow(h, cmd)
}

func (d *DesktopWindow) confirmedFullscreen() bool {
	foreground := GetForegroundWindow()
	if foreground == 0 {
		return false
	}
	root := RootOwner(foreground)
	if root == 0 || WindowBelongsToProcess(root, d.currentPID) {
		return false
	}
	class := WindowClassName(root)
	if class == "Progman" || class == "WorkerW" || class == "Shell_TrayWnd" {
		return false
	}
	if !IsWindowVisible(root) {
		return false
	}
	return IsFullscreen(root, d.monitor.Screen)
}

func (d *DesktopWindow) refreshDockMotion() {
	if d.hDock == 0 || !d.dockVisible {
		return
	}
	p := GetCursorPos()
	if p == d.lastCursor {
		return
	}
	d.lastCursor = p
	// Magnification only changes when the pointer is over the dock. Repainting
	// for global mouse movement caused steady flicker and needless CPU load.
	if IsCursorOver(d.hDock) {
		d.markDirty(d.hDock)
		d.flushRepaints()
	}
}

func (d *DesktopWindow) paint(hwnd HWND) {
	hdc, ps := BeginPaint(hwnd)
	defer EndPaint(hwnd, &ps)
	client := GetClientRect(hwnd)
	tk := tokens.Current()
	surface := rgb(tk.SurfaceR, tk.SurfaceG, tk.SurfaceB)
	brush := CreateSolidBrush(surface)
	defer DeleteObject(uintptr(brush))
	FillRect(hdc, client, brush)
	SetBkMode(hdc, TRANSPARENT)
	SetTextColor(hdc, rgb(tk.TextR, tk.TextG, tk.TextB))
	font := CreateFont(-int32(tk.FontBodySize), 400, tk.FontFace)
	oldFont := SelectObject(hdc, uintptr(font))
	defer func() {
		SelectObject(hdc, oldFont)
		DeleteObject(uintptr(font))
	}()

	switch hwnd {
	case d.hDock:
		d.paintDock(hdc, client)
	case d.hTop:
		d.paintTopBar(hdc, client)
	case d.hIsland:
		d.paintIsland(hdc, client)
	default:
		for _, h := range d.hWidgets {
			if h == hwnd {
				index := int(GetWindowLongPtr(h, GWLP_USERDATA)) - 1
				d.paintWidget(hdc, client, index)
				break
			}
		}
	}
}

func (d *DesktopWindow) paintDock(hdc HDC, client RECT) {
	tk := tokens.Current()
	items := dockItemsFromConfig(d.cfg.Dock.Items)
	running := win32RunningSet()
	for i := range items {
		items[i].Running = running[strings.ToLower(items[i].ID)]
	}
	pointer := GetCursorPos()
	screen := d.screen()
	work := d.workArea()
	rect := geometry.NewRect(int(client.Left), int(client.Top), int(client.Right), int(client.Bottom))
	opts := dock.LayoutOptions{
		Position:      dock.Position(d.cfg.Dock.Position),
		Screen:        screen,
		WorkArea:      work,
		Window:        rect,
		IconSize:      d.dockIconSize(),
		Gap:           8,
		MaxScale:      1.7,
		Spread:        1.2,
		Magnification: d.cfg.Dock.Magnification,
		Radius:        rect.Height,
		Bounce:        d.cfg.Dock.Bounce,
		LaunchTimes:   d.launchTimes,
		Now:           time.Now(),
	}
	layout := dock.Layout(items, geometry.Point{X: int(pointer.X), Y: int(pointer.Y)}, opts)
	oldPen := SelectObject(hdc, uintptr(NullPen()))
	defer SelectObject(hdc, oldPen)
	for i, icon := range layout.Icons {
		local := geometry.NewRect(icon.Rect.X-rect.X, icon.Rect.Y-rect.Y, icon.Rect.Right()-rect.X, icon.Rect.Bottom()-rect.Y)
		brush := CreateSolidBrush(rgb(tk.AccentR, tk.AccentG, tk.AccentB))
		oldBrush := SelectObject(hdc, uintptr(brush))
		Ellipse(hdc, int32(local.X+4), int32(local.Y+4), int32(local.Right()-4), int32(local.Bottom()-4))
		SelectObject(hdc, oldBrush)
		DeleteObject(uintptr(brush))
		textRect := RECT{Left: int32(local.X), Top: int32(local.Y), Right: int32(local.Right()), Bottom: int32(local.Bottom())}
		DrawText(hdc, icon.Label, &textRect, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		if icon.Running && d.cfg.Dock.RunningDots {
			dotBrush := CreateSolidBrush(rgb(tk.TextR, tk.TextG, tk.TextB))
			oldDotBrush := SelectObject(hdc, uintptr(dotBrush))
			Ellipse(hdc, int32(local.Center().X-3), int32(local.Bottom()-10), int32(local.Center().X+3), int32(local.Bottom()-4))
			SelectObject(hdc, oldDotBrush)
			DeleteObject(uintptr(dotBrush))
		}
		_ = i
	}
}

func (d *DesktopWindow) paintTopBar(hdc HDC, client RECT) {
	status := d.reader.TopBarStatus()
	if d.cfg.TopBar.GlobalMenu {
		status.ActiveApp = d.activeWindowTitle()
	}
	text := topbar.RenderStatus(d.cfg.TopBar, status)
	r := RECT{Left: client.Right - 520, Top: client.Top, Right: client.Right - 8, Bottom: client.Bottom}
	DrawText(hdc, text, &r, DT_RIGHT_ALIGN|DT_VCENTER|DT_SINGLELINE)
	menuRect := RECT{Left: client.Left + 12, Top: client.Top, Right: client.Left + 420, Bottom: client.Bottom}
	DrawText(hdc, topbar.RenderMenu(d.cfg.TopBar, status.ActiveApp), &menuRect, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
}

func (d *DesktopWindow) activeWindowTitle() string {
	foreground := GetForegroundWindow()
	if foreground == 0 || WindowBelongsToProcess(foreground, d.currentPID) {
		return ""
	}
	return GetWindowText(RootOwner(foreground))
}

func (d *DesktopWindow) paintIsland(hdc HDC, client RECT) {
	snap := d.island.Snapshot()
	text := "Desop"
	if snap.Active {
		text = snap.Title
		if snap.Detail != "" {
			text += "  " + snap.Detail
		}
		if snap.Progress > 0 {
			text += fmt.Sprintf("  %d%%", snap.Progress)
		}
	}
	r := client
	DrawText(hdc, text, &r, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
}

func (d *DesktopWindow) paintWidget(hdc HDC, client RECT, index int) {
	if index < 0 || index >= len(d.cfg.Widgets.Widgets) {
		return
	}
	w := d.cfg.Widgets.Widgets[index]
	title, lines := widgets.Render(w.Kind, d.cfg.Widgets.Note, time.Now(), d.reader.WidgetStats(), d.cfg.Widgets.WeatherURL)
	titleRect := RECT{Left: client.Left + 12, Top: client.Top + 6, Right: client.Right - 8, Bottom: client.Top + 30}
	DrawText(hdc, title, &titleRect, DT_LEFT|DT_SINGLELINE)
	if w.Kind == "notes" {
		return // the Edit child control renders note content
	}
	body := RECT{Left: client.Left + 12, Top: client.Top + 34, Right: client.Right - 12, Bottom: client.Bottom - 8}
	DrawText(hdc, strings.Join(lines, "\n"), &body, DT_LEFT|DT_WORDBREAK)
}

func (d *DesktopWindow) handleClick(hwnd HWND, lp uintptr) {
	if d.hDock != 0 && hwnd == d.hDock {
		x, y := int(lp&0xffff), int(lp>>16&0xffff)
		if x < 0 {
			x += 65536
		}
		if y < 0 {
			y += 65536
		}
		rect := GetWindowRect(hwnd)
		screenPoint := geometry.Point{X: int(rect.Left) + x, Y: int(rect.Top) + y}
		work := d.workArea()
		dockRect := d.dockWindowRect()
		items := dockItemsFromConfig(d.cfg.Dock.Items)
		layout := dock.Layout(items, geometry.Point{X: -10000, Y: -10000}, dock.LayoutOptions{
			Position: dock.Position(d.cfg.Dock.Position), Screen: d.screen(), WorkArea: work,
			Window: dockRect, IconSize: d.dockIconSize(), Gap: 8, Radius: dockRect.Height,
		})
		for i, icon := range layout.Icons {
			if icon.Rect.Contains(screenPoint) && i < len(d.cfg.Dock.Items) {
				if d.cfg.Dock.Bounce {
					d.launchTimes[icon.ID] = time.Now()
				}
				if err := ShellExecute("open", d.cfg.Dock.Items[i], "", "", SW_SHOWNORMAL); err != nil {
					MessageBox(hwnd, err.Error(), "Desop", 0)
				}
				return
			}
		}
		return
	}
	for _, h := range d.hWidgets {
		index := int(GetWindowLongPtr(h, GWLP_USERDATA)) - 1
		if hwnd != h || index < 0 || index >= len(d.cfg.Widgets.Widgets) {
			continue
		}
		if !d.cfg.Widgets.Widgets[index].Enabled {
			return
		}
		x, y := lpClientPoint(lp)
		rect := GetWindowRect(hwnd)
		d.drag = &widgetDrag{
			index:  index,
			offset: geometry.Point{X: int(rect.Left) + x - d.cfg.Widgets.Widgets[index].X, Y: int(rect.Top) + y - d.cfg.Widgets.Widgets[index].Y},
		}
		return
	}
	if hwnd == d.hIsland {
		snap := d.island.Snapshot()
		d.island.SetExpanded(!snap.Expanded)
	}
}

func (d *DesktopWindow) handleMouseMove(hwnd HWND, lp uintptr) {
	if d.drag == nil {
		return
	}
	if d.drag.index < 0 || d.drag.index >= len(d.cfg.Widgets.Widgets) {
		return
	}
	x, y := lpClientPoint(lp)
	screen := d.screen()
	work := d.workArea()
	target := geometry.Point{X: int(GetCursorPos().X) - d.drag.offset.X, Y: int(GetCursorPos().Y) - d.drag.offset.Y}
	size := widgetSize(d.cfg.Widgets.Widgets[d.drag.index].Kind)
	moved := widgets.MoveBy(d.cfg.Widgets.Widgets[d.drag.index], target.X-d.cfg.Widgets.Widgets[d.drag.index].X,
		target.Y-d.cfg.Widgets.Widgets[d.drag.index].Y, work, size, d.cfg.Widgets.GridSnap, d.cfg.Widgets.GridSize)
	d.cfg.Widgets.Widgets[d.drag.index].X = moved.X
	d.cfg.Widgets.Widgets[d.drag.index].Y = moved.Y
	moveWindow(hwnd, geometry.NewRect(moved.X, moved.Y, moved.X+size.Width, moved.Y+size.Height))
	InvalidateRect(hwnd, false)
	_ = screen
	_ = x
	_ = y
}

func (d *DesktopWindow) endWidgetDrag() {
	if d.drag == nil {
		return
	}
	d.drag = nil
	_ = config.Save(d.cfgPath, d.cfg)
}

func widgetSize(kind string) geometry.Rect {
	switch kind {
	case "clock":
		return geometry.Rect{Width: 280, Height: 150}
	case "status":
		return geometry.Rect{Width: 280, Height: 190}
	case "notes":
		return geometry.Rect{Width: 300, Height: 220}
	case "weather", "calendar":
		return geometry.Rect{Width: 280, Height: 190}
	default:
		return geometry.Rect{Width: 240, Height: 160}
	}
}

func lpClientPoint(lp uintptr) (int, int) {
	x, y := int(lp&0xffff), int(lp>>16&0xffff)
	if x >= 1<<15 {
		x -= 1 << 16
	}
	if y >= 1<<15 {
		y -= 1 << 16
	}
	return x, y
}

func (d *DesktopWindow) handleCommand(id int) {
	if id >= cmdDockItemBase && id < cmdDockItemBase+len(d.cfg.Dock.Items) {
		item := d.cfg.Dock.Items[id-cmdDockItemBase]
		if err := ShellExecute("open", item, "", "", SW_SHOWNORMAL); err != nil {
			MessageBox(d.hwnd, err.Error(), "Desop", 0)
		}
		return
	}
	if id >= cmdWidgetBase && id < cmdWidgetBase+len(d.cfg.Widgets.Widgets) {
		index := id - cmdWidgetBase
		d.cfg.Widgets.Widgets[index].Enabled = !d.cfg.Widgets.Widgets[index].Enabled
		_ = config.Save(d.cfgPath, d.cfg)
		d.restartShell()
		return
	}
	switch id {
	case cmdExit:
		d.restartShellWithoutRelaunch()
		PostQuitMessage(0)
	case cmdSettings:
		_ = RunSettings(d.cfgPath)
		// Settings may change enabled components; restart to apply cleanly.
		d.restartShell()
	case cmdRestart:
		d.restartShell()
	case cmdAutostart:
		d.cfg.General.Autostart = !d.cfg.General.Autostart
		_ = SetAutostart(d.cfg.General.Autostart)
		_ = config.Save(d.cfgPath, d.cfg)
	case cmdEnableDock, cmdEnableTopBar, cmdEnableWidgets, cmdEnableIsland:
		name := map[int]string{
			cmdEnableDock: config.ComponentDock, cmdEnableTopBar: config.ComponentTopBar,
			cmdEnableWidgets: config.ComponentWidgets, cmdEnableIsland: config.ComponentIsland,
		}[id]
		enabled := false
		switch name {
		case config.ComponentDock:
			enabled = !d.cfg.Dock.Enabled
		case config.ComponentTopBar:
			enabled = !d.cfg.TopBar.Enabled
		case config.ComponentWidgets:
			enabled = !d.cfg.Widgets.Enabled
		case config.ComponentIsland:
			enabled = !d.cfg.Island.Enabled
		}
		_ = d.cfg.SetComponent(name, enabled)
		_ = config.Save(d.cfgPath, d.cfg)
		d.restartShell()
	}
}

func (d *DesktopWindow) restartShell() {
	if d.closing || d.restarting {
		return
	}
	d.restarting = true
	d.Close()
	if err := ShellExecute("open", os.Args[0], "", "", SW_SHOWNORMAL); err != nil {
		MessageBox(0, "Restart failed: "+err.Error(), "Desop", 0)
	}
	PostQuitMessage(0)
}

func (d *DesktopWindow) restartShellWithoutRelaunch() {
	if d.closing || d.restarting {
		return
	}
	d.closing = true
	d.Close()
}

func (d *DesktopWindow) showTrayMenu() {
	SetForegroundWindow(d.hwnd)
	cmd := PopupMenu(d.hwnd, GetCursorPos().X, GetCursorPos().Y, []MenuItem{
		{ID: cmdSettings, Text: "Settings"},
		{ID: cmdEnableDock, Text: toggleLabel("Dock", d.cfg.Dock.Enabled)},
		{ID: cmdEnableTopBar, Text: toggleLabel("Top bar", d.cfg.TopBar.Enabled)},
		{ID: cmdEnableWidgets, Text: toggleLabel("Widgets", d.cfg.Widgets.Enabled)},
		{ID: cmdEnableIsland, Text: toggleLabel("Island", d.cfg.Island.Enabled)},
		{ID: cmdAutostart, Text: toggleLabel("Autostart", d.cfg.General.Autostart)},
		{ID: cmdRestart, Text: "Restart Desop"},
		{Separator: true},
		{ID: cmdExit, Text: "Exit Desop"},
	})
	if cmd != 0 {
		d.handleCommand(int(cmd))
	}
}

func toggleLabel(name string, on bool) string {
	if on {
		return "Disable " + name
	}
	return "Enable " + name
}

// RunSettings shows a native dialog with four component checkboxes, autostart
// and an API help button. It saves only when OK is pressed.
func RunSettings(path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}
	instance := GetModuleHandle("")
	proc := syscall.NewCallback(func(hwnd HWND, msg uint32, wp, lp uintptr) uintptr {
		return DefWindowProc(hwnd, msg, wp, lp)
	})
	settingsClass := fmt.Sprintf("DesopSettings_%d", CurrentProcessID())
	wc := WNDCLASSEXW{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEXW{})), LpfnWndProc: proc,
		HInstance: instance, HbrBackground: HBRUSH(6), LpszClassName: UTF16(settingsClass),
	}
	_, err = RegisterClassEx(&wc)
	if err != nil && !errors.Is(err, syscall.Errno(1410)) {
		return err
	}
	hwnd, err := CreateWindowEx(WS_EX_TOOLWINDOW, settingsClass, "Desop Settings", WS_OVERLAPPEDWINDOW,
		200, 140, 520, 420, 0, 0, instance, 0)
	if err != nil {
		return err
	}
	defer DestroyWindow(hwnd)
	checkbox := func(x, y int32, text string, checked bool) HWND {
		h, e := CreateWindowEx(0, "Button", text, 0x00000003|WS_CHILD|WS_VISIBLE|WS_TABSTOP, x, y, 460, 26, hwnd, 0, instance, 0)
		if e != nil {
			return 0
		}
		if checked {
			SendMessage(h, 0x00F1, 1, 0) // BM_SETCHECK
		}
		return h
	}
	hDock := checkbox(24, 24, "Dock", cfg.Dock.Enabled)
	hTop := checkbox(24, 56, "Top bar", cfg.TopBar.Enabled)
	hWidgets := checkbox(24, 88, "Widgets", cfg.Widgets.Enabled)
	hIsland := checkbox(24, 120, "Dynamic island", cfg.Island.Enabled)
	hAuto := checkbox(24, 152, "Start Desop when I sign in", cfg.General.Autostart)
	hAPI := checkbox(24, 184, "Local island event API (127.0.0.1 only)", cfg.Island.LocalAPI)
	hOK, _ := CreateWindowEx(0, "Button", "Save", BS_PUSHBUTTON|WS_CHILD|WS_VISIBLE|WS_TABSTOP, 24, 232, 100, 30, hwnd, HMENU(IDOK), instance, 0)
	hCancel, _ := CreateWindowEx(0, "Button", "Cancel", BS_PUSHBUTTON|WS_CHILD|WS_VISIBLE|WS_TABSTOP, 140, 232, 100, 30, hwnd, HMENU(2), instance, 0)
	hHelp, _ := CreateWindowEx(0, "Button", "Local API examples", BS_PUSHBUTTON|WS_CHILD|WS_VISIBLE|WS_TABSTOP, 256, 232, 180, 30, hwnd, HMENU(3), instance, 0)
	done := false
	quit := false
	ShowWindow(hwnd, SW_SHOWNORMAL)
	UpdateWindow(hwnd)
	// Stay on the settings window only: a thread-wide GetMessage would swallow
	// the desktop host's tray and timer messages while this dialog is open.
	for !done && !quit {
		msg := MSG{}
		got, err := GetMessageForWindow(&msg, hwnd, 0, 0)
		if err != nil || !got {
			quit = err == nil
			break
		}
		if msg.Message == WM_COMMAND {
			id := int(msg.WParam & 0xffff)
			_ = hOK
			_ = hCancel
			_ = hHelp
			if id == IDOK {
				_ = cfg.SetComponent(config.ComponentDock, isChecked(hDock))
				_ = cfg.SetComponent(config.ComponentTopBar, isChecked(hTop))
				_ = cfg.SetComponent(config.ComponentWidgets, isChecked(hWidgets))
				_ = cfg.SetComponent(config.ComponentIsland, isChecked(hIsland))
				cfg.General.Autostart = isChecked(hAuto)
				cfg.Island.LocalAPI = isChecked(hAPI)
				if err := config.Save(path, cfg); err != nil {
					return err
				}
				if cfg.General.Autostart {
					_ = SetAutostart(true)
				} else {
					_ = SetAutostart(false)
				}
				done = true
				continue
			}
			if id == 2 || id == 3 {
				if id == 3 {
					MessageBox(hwnd, island.APIExample(), "Desop local API", 0)
				}
				done = true
				DestroyWindow(hwnd)
				continue
			}
		}
		if msg.Message == WM_CLOSE {
			done = true
			DestroyWindow(hwnd)
			continue
		}
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
	return nil
}

func isChecked(h HWND) bool {
	if h == 0 {
		return false
	}
	return SendMessage(h, 0x00F0, 0, 0) == 1 // BM_GETCHECK
}

// SetAutostart writes a Run entry under HKCU only. Disable removes the entry.
func SetAutostart(enabled bool) error {
	if !enabled {
		return removeAutostartValue()
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return writeAutostartValue(`"` + exe + `"`)
}
