//go:build windows

// Package win32 wraps only the Win32 APIs Desop needs, with no third-party
// dependencies. Every call stays in user mode and uses public interfaces.
package win32

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/ntars/Desop/internal/sysmetrics"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	iphlpapi = syscall.NewLazyDLL("iphlpapi.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")

	procRegisterClassExW           = user32.NewProc("RegisterClassExW")
	procCreateWindowExW            = user32.NewProc("CreateWindowExW")
	procDestroyWindow              = user32.NewProc("DestroyWindow")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procUpdateWindow               = user32.NewProc("UpdateWindow")
	procDefWindowProcW             = user32.NewProc("DefWindowProcW")
	procGetMessageW                = user32.NewProc("GetMessageW")
	procPeekMessageW               = user32.NewProc("PeekMessageW")
	procTranslateMessage           = user32.NewProc("TranslateMessage")
	procDispatchMessageW           = user32.NewProc("DispatchMessageW")
	procPostQuitMessage            = user32.NewProc("PostQuitMessage")
	procGetModuleHandleW           = kernel32.NewProc("GetModuleHandleW")
	procLoadCursorW                = user32.NewProc("LoadCursorW")
	procLoadIconW                  = user32.NewProc("LoadIconW")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW          = user32.NewProc("GetWindowLongPtrW")
	procSetWindowTextW             = user32.NewProc("SetWindowTextW")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW       = user32.NewProc("GetWindowTextLengthW")
	procGetClientRect              = user32.NewProc("GetClientRect")
	procGetWindowRect              = user32.NewProc("GetWindowRect")
	procInvalidateRect             = user32.NewProc("InvalidateRect")
	procBeginPaint                 = user32.NewProc("BeginPaint")
	procEndPaint                   = user32.NewProc("EndPaint")
	procSetTimer                   = user32.NewProc("SetTimer")
	procKillTimer                  = user32.NewProc("KillTimer")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procScreenToClient             = user32.NewProc("ScreenToClient")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procSystemParametersInfoW      = user32.NewProc("SystemParametersInfoW")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowRgn               = user32.NewProc("SetWindowRgn")
	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	procGetWindowThreadProcessId   = user32.NewProc("GetWindowThreadProcessId")
	procGetAncestor                = user32.NewProc("GetAncestor")
	procGetClassNameW              = user32.NewProc("GetClassNameW")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procSendMessageW               = user32.NewProc("SendMessageW")
	procPostMessageW               = user32.NewProc("PostMessageW")
	procCreatePopupMenu            = user32.NewProc("CreatePopupMenu")
	procAppendMenuW                = user32.NewProc("AppendMenuW")
	procTrackPopupMenu             = user32.NewProc("TrackPopupMenu")
	procDestroyMenu                = user32.NewProc("DestroyMenu")
	procEnumDisplayMonitors        = user32.NewProc("EnumDisplayMonitors")
	procGetMonitorInfoW            = user32.NewProc("GetMonitorInfoW")
	procGetDpiForWindow            = user32.NewProc("GetDpiForWindow")
	procMessageBoxW                = user32.NewProc("MessageBoxW")

	procCreateSolidBrush   = gdi32.NewProc("CreateSolidBrush")
	procCreatePen          = gdi32.NewProc("CreatePen")
	procGetStockObject     = gdi32.NewProc("GetStockObject")
	procCreateRoundRectRgn = gdi32.NewProc("CreateRoundRectRgn")
	procFrameRgn           = gdi32.NewProc("FrameRgn")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procFillRect           = user32.NewProc("FillRect")
	procDrawTextW          = user32.NewProc("DrawTextW")
	procSetTextColor       = gdi32.NewProc("SetTextColor")
	procSetBkMode          = gdi32.NewProc("SetBkMode")
	procCreateFontW        = gdi32.NewProc("CreateFontW")
	procRectangle          = gdi32.NewProc("Rectangle")
	procEllipse            = gdi32.NewProc("Ellipse")
	procMoveToEx           = gdi32.NewProc("MoveToEx")
	procLineTo             = gdi32.NewProc("LineTo")

	procShellExecuteW     = shell32.NewProc("ShellExecuteW")
	procShell_NotifyIconW = shell32.NewProc("Shell_NotifyIconW")

	procDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")

	procGlobalMemoryStatusEx     = kernel32.NewProc("GlobalMemoryStatusEx")
	procGetDiskFreeSpaceExW      = kernel32.NewProc("GetDiskFreeSpaceExW")
	procGetSystemTimes           = kernel32.NewProc("GetSystemTimes")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procGetCurrentProcess        = kernel32.NewProc("GetCurrentProcess")
	procGetProcessTimes          = kernel32.NewProc("GetProcessTimes")
	procGetProcessMemoryInfo     = psapi.NewProc("GetProcessMemoryInfo")

	procGetIfTable          = iphlpapi.NewProc("GetIfTable")
	procCreateEventW        = kernel32.NewProc("CreateEventW")
	procOpenEventW          = kernel32.NewProc("OpenEventW")
	procResetEvent          = kernel32.NewProc("ResetEvent")
	procSetEvent            = kernel32.NewProc("SetEvent")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
)

// Handles and primitive window types.
type (
	HWND      uintptr
	HDC       uintptr
	HBRUSH    uintptr
	HPEN      uintptr
	HRGN      uintptr
	HFONT     uintptr
	HANDLE    uintptr
	HMENU     uintptr
	HICON     uintptr
	HCURSOR   uintptr
	HINSTANCE uintptr
	ATOM      uint16
)

// POINT is a Win32 POINT.
type POINT struct {
	X int32
	Y int32
}

// RECT is a Win32 RECT.
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// WNDCLASSEXW is the Unicode extended window class structure.
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

// MSG is a Win32 MSG.
type MSG struct {
	Hwnd     HWND
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

// PAINTSTRUCT is a Win32 PAINTSTRUCT.
type PAINTSTRUCT struct {
	Hdc      HDC
	FErase   bool
	RcPaint  RECT
	FRestore bool
	FUpdate  bool
	RgnHRgn  HRGN
}

// MONITORINFO is the Win32 monitor information structure.
type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

// MEMORYSTATUSEX is the Win32 extended memory status structure.
type MEMORYSTATUSEX struct {
	DwLength                uint32
	DwMemoryLoad            uint32
	UllTotalPhys            uint64
	UllAvailPhys            uint64
	UllTotalPageFile        uint64
	UllAvailPageFile        uint64
	UllTotalVirtual         uint64
	UllAvailVirtual         uint64
	UllAvailExtendedVirtual uint64
}

// FILETIME is a Win32 FILETIME.
type FILETIME struct {
	LowDateTime  uint32
	HighDateTime uint32
}

// PROCESSENTRY32W is the Unicode process snapshot entry.
type PROCESSENTRY32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

// NOTIFYICONDATAW is the Unicode tray icon structure.
type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            HICON
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     HICON
}

// NOTIFYICONDATAW_V1 is the version 1 structure accepted by Windows 10/11
// without the extra Vista+ fields. Its fixed size is essential: passing the
// larger struct with a mismatched cbSize makes Shell_NotifyIcon fail.
type NOTIFYICONDATAW_V1 struct {
	CbSize           uint32
	HWnd             HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            HICON
	SzTip            [128]uint16
}

// PROCESS_MEMORY_COUNTERS is the PSAPI structure used for working-set data.
type PROCESS_MEMORY_COUNTERS struct {
	Cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

// Window styles and common messages.
const (
	WS_POPUP            = 0x80000000
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_TABSTOP          = 0x00010000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_VSCROLL          = 0x00200000
	WS_EX_TOOLWINDOW    = 0x00000080
	WS_EX_NOACTIVATE    = 0x08000000
	WS_EX_LAYERED       = 0x00080000
	WS_EX_TOPMOST       = 0x00000008
	WS_EX_TRANSPARENT   = 0x00000020

	BS_PUSHBUTTON  = 0x00000000
	ES_MULTILINE   = 0x00000004
	ES_AUTOVSCROLL = 0x00000040
	ES_WANTRETURN  = 0x00001000

	WM_CREATE      = 0x0001
	WM_DESTROY     = 0x0002
	WM_QUIT        = 0x0012
	WM_CLOSE       = 0x0010
	WM_COMMAND     = 0x0111
	WM_TIMER       = 0x0113
	WM_PAINT       = 0x000F
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_MOUSEMOVE   = 0x0200
	WM_RBUTTONUP   = 0x0205
	WM_APP         = 0x8000
	WM_APP_TRAY    = 0x8001
	WM_APP_REFRESH = 0x8002
	WM_APP_EXIT    = 0x8003

	SW_HIDE       = 0
	SW_SHOWNORMAL = 1
	SW_SHOWNA     = 8

	IDOK = 1

	GWLP_USERDATA = -21

	SM_XVIRTUALSCREEN  = 76
	SM_YVIRTUALSCREEN  = 77
	SM_CXVIRTUALSCREEN = 78
	SM_CYVIRTUALSCREEN = 79

	SPI_GETWORKAREA      = 0x0030
	LWA_ALPHA            = 0x00000002
	MONITORINFOF_PRIMARY = 0x00000001

	DT_LEFT       = 0x00000000
	DT_CENTER     = 0x00000001
	DT_VCENTER    = 0x00000004
	DT_SINGLELINE = 0x00000020
	DT_WORDBREAK  = 0x00000010

	TRANSPARENT = 1
	NULL_PEN    = 8

	TH32CS_SNAPPROCESS = 0x00000002

	NIF_ICON    = 0x00000002
	NIF_MESSAGE = 0x00000001
	NIF_TIP     = 0x00000004
	NIM_ADD     = 0x00000000
	NIM_DELETE  = 0x00000002
	NIM_MODIFY  = 0x00000001

	MF_STRING     = 0x00000000
	MF_SEPARATOR  = 0x00000800
	TPM_LEFTALIGN = 0x0000
	TPM_RETURNCMD = 0x0100

	DWMWA_SYSTEMBACKDROP_TYPE   = 38
	DWMWA_EXTENDED_FRAME_BOUNDS = 9

	ERROR_INSUFFICIENT_BUFFER = 122
	WAIT_OBJECT_0             = 0
	WM_DPICHANGED             = 0x02E0

	IDI_APPLICATION = 32512
)

// UTF16 converts a Go string to a NUL-terminated UTF-16 pointer.
func UTF16(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		var empty [1]uint16
		p = &empty[0]
	}
	return p
}

// UTF16String converts a NUL-terminated UTF-16 pointer to a Go string.
func UTF16String(p *uint16) string {
	if p == nil {
		return ""
	}
	return syscall.UTF16ToString(unsafe.Slice(p, 32768))
}

// LastError converts a syscall errno into an error, ignoring success.
func LastError(e1 error) error {
	if errno, ok := e1.(syscall.Errno); ok && errno == 0 {
		return nil
	}
	return e1
}

// RegisterClassEx registers a Unicode window class.
func RegisterClassEx(wc *WNDCLASSEXW) (ATOM, error) {
	r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	if r == 0 {
		return 0, LastError(err)
	}
	return ATOM(r), nil
}

// CreateWindowEx creates a window.
func CreateWindowEx(exStyle uint32, class, title string, style uint32, x, y, w, h int32, parent HWND, menu HMENU, instance HINSTANCE, param uintptr) (HWND, error) {
	r, _, err := procCreateWindowExW.Call(
		uintptr(exStyle), uintptr(unsafe.Pointer(UTF16(class))), uintptr(unsafe.Pointer(UTF16(title))),
		uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(parent), uintptr(menu), uintptr(instance), param)
	if r == 0 {
		return 0, LastError(err)
	}
	return HWND(r), nil
}

// WindowDPI returns the effective DPI of a window. Windows 10 1607+ provides
// GetDpiForWindow; older systems use the safe 96-DPI fallback.
func WindowDPI(hwnd HWND) uint32 {
	if hwnd == 0 {
		return 96
	}
	r, _, _ := procGetDpiForWindow.Call(uintptr(hwnd))
	if r < 96 || r > 1000 {
		return 96
	}
	return uint32(r)
}

// DestroyWindow destroys a window.
func DestroyWindow(hwnd HWND) error {
	r, _, err := procDestroyWindow.Call(uintptr(hwnd))
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// ShowWindow shows or hides a window.
func ShowWindow(hwnd HWND, cmd int32) { procShowWindow.Call(uintptr(hwnd), uintptr(cmd)) }

// UpdateWindow forces an initial paint.
func UpdateWindow(hwnd HWND) { procUpdateWindow.Call(uintptr(hwnd)) }

// DefWindowProc dispatches unhandled messages.
func DefWindowProc(hwnd HWND, msg uint32, wp, lp uintptr) uintptr {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wp, lp)
	return r
}

// GetMessage retrieves the next message. It returns false when WM_QUIT arrives.
func GetMessage(msg *MSG) (bool, error) {
	return GetMessageForWindow(msg, 0, 0, 0)
}

// GetMessageForWindow filters messages by window handle. This prevents a
// modal settings window from stealing the desktop host's messages.
func GetMessageForWindow(msg *MSG, hwnd HWND, minMsg, maxMsg uint32) (bool, error) {
	r, _, err := procGetMessageW.Call(uintptr(unsafe.Pointer(msg)), uintptr(hwnd), uintptr(minMsg), uintptr(maxMsg))
	if r == ^uintptr(0) {
		return false, LastError(err)
	}
	return r != 0, nil
}

// PeekThreadMessage retrieves the next non-blocking message for the current
// thread. Thread-scoped peeping is required because WM_QUIT has no hwnd and
// would be invisible to a hwnd-filtered loop.
func PeekThreadMessage(msg *MSG) (bool, bool, error) {
	r, _, err := procPeekMessageW.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0, 1) // PM_REMOVE
	if r == ^uintptr(0) {
		return false, false, LastError(err)
	}
	if r == 0 {
		return false, false, nil
	}
	return true, msg.Message == WM_QUIT, nil
}

// TranslateMessage performs keyboard translation.
func TranslateMessage(msg *MSG) { procTranslateMessage.Call(uintptr(unsafe.Pointer(msg))) }

// DispatchMessage sends a message to its window procedure.
func DispatchMessage(msg *MSG) { procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg))) }

// PostQuitMessage stops the message loop.
func PostQuitMessage(code int32) { procPostQuitMessage.Call(uintptr(code)) }

// GetModuleHandle returns a module handle; nil means the current executable.
func GetModuleHandle(name string) HINSTANCE {
	var p *uint16
	if name != "" {
		p = UTF16(name)
	}
	r, _, _ := procGetModuleHandleW.Call(uintptr(unsafe.Pointer(p)))
	return HINSTANCE(r)
}

// LoadCursor loads a predefined cursor.
func LoadCursor(id uintptr) HCURSOR {
	r, _, _ := procLoadCursorW.Call(0, id)
	return HCURSOR(r)
}

// SetWindowLongPtr stores a pointer-sized value in window user data.
func SetWindowLongPtr(hwnd HWND, index int32, value uintptr) uintptr {
	r, _, _ := procSetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index), value)
	return r
}

// GetWindowLongPtr reads a pointer-sized value from window user data.
func GetWindowLongPtr(hwnd HWND, index int32) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index))
	return r
}

// SetWindowText sets a window's title text.
func SetWindowText(hwnd HWND, text string) {
	procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(UTF16(text))))
}

// GetWindowText gets a window's text.
func GetWindowText(hwnd HWND) string {
	n, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	buf := make([]uint16, int(n)+2)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

// GetClientRect returns the client rectangle.
func GetClientRect(hwnd HWND) RECT {
	var r RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	return r
}

// GetWindowRect returns the screen rectangle.
func GetWindowRect(hwnd HWND) RECT {
	var r RECT
	procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
	return r
}

// IsCursorOver reports whether the cursor is over a window.
func IsCursorOver(hwnd HWND) bool {
	p := GetCursorPos()
	r := GetWindowRect(hwnd)
	return p.X >= r.Left && p.X < r.Right && p.Y >= r.Top && p.Y < r.Bottom
}

// InvalidateRect requests a repaint.
func InvalidateRect(hwnd HWND, erase bool) {
	var p *RECT
	eraseValue := uintptr(0)
	if erase {
		eraseValue = 1
	}
	procInvalidateRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(p)), eraseValue)
}

// BeginPaint starts a paint operation.
func BeginPaint(hwnd HWND) (HDC, PAINTSTRUCT) {
	var ps PAINTSTRUCT
	r, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	return HDC(r), ps
}

// EndPaint ends a paint operation.
func EndPaint(hwnd HWND, ps *PAINTSTRUCT) {
	procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
}

// SetTimer installs a timer on a window.
func SetTimer(hwnd HWND, id uintptr, ms uint32) uintptr {
	r, _, _ := procSetTimer.Call(uintptr(hwnd), id, uintptr(ms), 0)
	return r
}

// KillTimer removes a window timer.
func KillTimer(hwnd HWND, id uintptr) { procKillTimer.Call(uintptr(hwnd), id) }

// GetCursorPos returns the cursor position in screen coordinates.
func GetCursorPos() POINT {
	var p POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	return p
}

// StockIcon returns a shared system icon; stock icons must not be destroyed.
func StockIcon(id uintptr) HICON {
	r, _, _ := procLoadIconW.Call(0, id)
	return HICON(r)
}

// ScreenToClient converts a screen point to client coordinates.
func ScreenToClient(hwnd HWND, p *POINT) {
	procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(p)))
}

// GetSystemMetrics returns a system metric.
func GetSystemMetrics(index int32) int {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(r)
}

// GetWorkArea returns the primary monitor work area.
func GetWorkArea() RECT {
	var r RECT
	procSystemParametersInfoW.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&r)), 0)
	return r
}

// SetLayeredWindowAttributes applies whole-window alpha transparency.
func SetLayeredWindowAttributes(hwnd HWND, alpha uint8) error {
	r, _, err := procSetLayeredWindowAttributes.Call(uintptr(hwnd), 0, uintptr(alpha), LWA_ALPHA)
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// SetWindowRgn applies a clipping region, producing rounded corners.
func SetWindowRgn(hwnd HWND, rgn HRGN) error {
	r, _, err := procSetWindowRgn.Call(uintptr(hwnd), uintptr(rgn), 1)
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// GetForegroundWindow returns the foreground window handle.
func GetForegroundWindow() HWND {
	r, _, _ := procGetForegroundWindow.Call()
	return HWND(r)
}

// SetForegroundWindow brings a window to the foreground for menu tracking.
func SetForegroundWindow(hwnd HWND) {
	procSetForegroundWindow.Call(uintptr(hwnd))
}

// WindowClassName returns the class name of a window.
func WindowClassName(hwnd HWND) string {
	buf := make([]uint16, 256)
	r, _, _ := procGetClassNameW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

// RootOwner resolves a window to its top-level ancestor.
func RootOwner(hwnd HWND) HWND {
	r, _, _ := procGetAncestor.Call(uintptr(hwnd), 2) // GA_ROOT
	return HWND(r)
}

// IsWindowVisible reports whether the window is visible.
func IsWindowVisible(hwnd HWND) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

// WindowBelongsToProcess checks which process owns a window.
func WindowBelongsToProcess(hwnd HWND, pid uint32) bool {
	var actual uint32
	procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&actual)))
	return actual == pid
}

// CurrentProcessID returns the current process identifier.
func CurrentProcessID() uint32 {
	r, _, _ := kernel32.NewProc("GetCurrentProcessId").Call()
	return uint32(r)
}

// ProcessTimes returns raw kernel/user CPU times for the current process.
func ProcessTimes() (kernel, user FILETIME, ok bool) {
	h, _, _ := procGetCurrentProcess.Call()
	var create, exit FILETIME
	r, _, _ := procGetProcessTimes.Call(h, uintptr(unsafe.Pointer(&create)), uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)), uintptr(unsafe.Pointer(&user)))
	return kernel, user, r != 0
}

// WorkingSetBytes returns the current process working set size.
func WorkingSetBytes() uint64 {
	h, _, _ := procGetCurrentProcess.Call()
	var counters PROCESS_MEMORY_COUNTERS
	counters.Cb = uint32(unsafe.Sizeof(counters))
	r, _, _ := procGetProcessMemoryInfo.Call(h, uintptr(unsafe.Pointer(&counters)), uintptr(counters.Cb))
	if r == 0 {
		return 0
	}
	return uint64(counters.WorkingSetSize)
}

// CreateStopEvent creates a manual-reset event used by `desop.exe stop`.
func CreateStopEvent(name string) (HANDLE, error) {
	r, _, err := procCreateEventW.Call(0, 1, 0, uintptr(unsafe.Pointer(UTF16(name))))
	if r == 0 {
		return 0, LastError(err)
	}
	h := HANDLE(r)
	_ = ResetStopEvent(h)
	return h, nil
}

// OpenStopEvent opens an existing session event with EVENT_MODIFY_STATE.
func OpenStopEvent(name string) (HANDLE, error) {
	r, _, err := procOpenEventW.Call(0x0002, 0, uintptr(unsafe.Pointer(UTF16(name))))
	if r == 0 {
		return 0, LastError(err)
	}
	return HANDLE(r), nil
}

// ResetStopEvent clears a stop event left signalled by a previous instance.
func ResetStopEvent(h HANDLE) error {
	r, _, err := procResetEvent.Call(uintptr(h))
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// SetStopEvent signals a stop event.
func SetStopEvent(h HANDLE) error {
	r, _, err := procSetEvent.Call(uintptr(h))
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// WaitForStopEvent waits up to milliseconds for a stop signal.
func WaitForStopEvent(h HANDLE, milliseconds uint32) bool {
	r, _, _ := procWaitForSingleObject.Call(uintptr(h), uintptr(milliseconds))
	return r == WAIT_OBJECT_0
}

// CloseHandle closes a kernel handle.
func CloseHandle(h HANDLE) { procCloseHandle.Call(uintptr(h)) }

// SendMessage sends a message synchronously.
func SendMessage(hwnd HWND, msg uint32, wp, lp uintptr) uintptr {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wp, lp)
	return r
}

// PostMessage posts a message asynchronously.
func PostMessage(hwnd HWND, msg uint32, wp, lp uintptr) error {
	r, _, err := procPostMessageW.Call(uintptr(hwnd), uintptr(msg), wp, lp)
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// CreateSolidBrush creates a solid GDI brush.
func CreateSolidBrush(rgb uint32) HBRUSH {
	r, _, _ := procCreateSolidBrush.Call(uintptr(rgb))
	return HBRUSH(r)
}

// CreatePen creates a solid cosmetic pen.
func CreatePen(style, width int32, rgb uint32) HPEN {
	r, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(rgb))
	return HPEN(r)
}

// NullPen returns the shared NULL_PEN stock object.
func NullPen() HPEN {
	r, _, _ := procGetStockObject.Call(NULL_PEN)
	return HPEN(r)
}

// CreateRoundRectRgn creates a rounded rectangle GDI region.
func CreateRoundRectRgn(x1, y1, x2, y2, w, h int32) HRGN {
	r, _, _ := procCreateRoundRectRgn.Call(uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2), uintptr(w), uintptr(h))
	return HRGN(r)
}

// FrameRgn draws a border around a GDI region.
func FrameRgn(hdc HDC, rgn HRGN, brush HBRUSH, w, h int32) {
	procFrameRgn.Call(uintptr(hdc), uintptr(rgn), uintptr(brush), uintptr(w), uintptr(h))
}

// DeleteObject releases a GDI object.
func DeleteObject(obj uintptr) { procDeleteObject.Call(obj) }

// SelectObject selects a GDI object into a device context.
func SelectObject(hdc HDC, obj uintptr) uintptr {
	r, _, _ := procSelectObject.Call(uintptr(hdc), obj)
	return r
}

// FillRect fills a rectangle with a brush.
func FillRect(hdc HDC, rect RECT, brush HBRUSH) {
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&rect)), uintptr(brush))
}

// DrawText draws formatted text into a rectangle.
func DrawText(hdc HDC, text string, rect *RECT, flags uint32) int {
	r, _, _ := procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(UTF16(text))), ^uintptr(0), uintptr(unsafe.Pointer(rect)), uintptr(flags))
	return int(r)
}

// SetTextColor sets the GDI text colour.
func SetTextColor(hdc HDC, rgb uint32) uint32 {
	r, _, _ := procSetTextColor.Call(uintptr(hdc), uintptr(rgb))
	return uint32(r)
}

// SetBkMode sets the GDI background mode.
func SetBkMode(hdc HDC, mode int32) int32 {
	r, _, _ := procSetBkMode.Call(uintptr(hdc), uintptr(mode))
	return int32(r)
}

// CreateFont creates a GDI font. Negative height means character height in pixels.
func CreateFont(height, weight int32, face string) HFONT {
	r, _, _ := procCreateFontW.Call(
		uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(UTF16(face))))
	return HFONT(r)
}

// Rectangle draws a rectangle outline with the current pen.
func Rectangle(hdc HDC, x1, y1, x2, y2 int32) {
	procRectangle.Call(uintptr(hdc), uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
}

// Ellipse draws an ellipse with the current brush and pen.
func Ellipse(hdc HDC, x1, y1, x2, y2 int32) {
	procEllipse.Call(uintptr(hdc), uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2))
}

// CreatePopupMenu creates an in-memory popup menu.
func CreatePopupMenu() HMENU {
	r, _, _ := procCreatePopupMenu.Call()
	return HMENU(r)
}

// AppendMenu adds a menu item.
func AppendMenu(menu HMENU, flags uint32, id uintptr, text string) error {
	r, _, err := procAppendMenuW.Call(uintptr(menu), uintptr(flags), id, uintptr(unsafe.Pointer(UTF16(text))))
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// TrackPopupMenu shows a popup menu and returns the selected command.
func TrackPopupMenu(menu HMENU, x, y int32, hwnd HWND) uintptr {
	r, _, _ := procTrackPopupMenu.Call(uintptr(menu), TPM_LEFTALIGN|TPM_RETURNCMD, uintptr(x), uintptr(y), 0, uintptr(hwnd), 0)
	return r
}

// DestroyMenu releases a popup menu.
func DestroyMenu(menu HMENU) { procDestroyMenu.Call(uintptr(menu)) }

// DwmSetSystemBackdrop asks Windows for system acrylic where available.
func DwmSetSystemBackdrop(hwnd HWND) error {
	value := uint32(3) // DWMSBT_TRANSIENTWINDOW (acrylic)
	r, _, _ := procDwmSetWindowAttribute.Call(uintptr(hwnd), DWMWA_SYSTEMBACKDROP_TYPE, uintptr(unsafe.Pointer(&value)), uintptr(4))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

// IsFullscreen checks the conservative DWM frame bound against the monitor.
func IsFullscreen(hwnd HWND, monitor RECT) bool {
	if hwnd == 0 {
		return false
	}
	var frame RECT
	r, _, _ := procDwmCall("DwmGetWindowAttribute", uintptr(hwnd), DWMWA_EXTENDED_FRAME_BOUNDS, uintptr(unsafe.Pointer(&frame)), uintptr(4*4))
	if r != 0 {
		return false
	}
	return abs(frame.Left-monitor.Left) <= 2 && abs(frame.Top-monitor.Top) <= 2 &&
		abs(frame.Right-monitor.Right) <= 2 && abs(frame.Bottom-monitor.Bottom) <= 2
}

// procDwmCall resolves rare DWM calls lazily without adding package-level noise.
func procDwmCall(name string, args ...uintptr) (uintptr, uintptr, error) {
	r, second, err := dwmapi.NewProc(name).Call(args...)
	return r, second, err
}

// NotifyIcon adds, modifies or removes the tray icon with the given tooltip.
func NotifyIcon(hwnd HWND, message uint32, tooltip string) error {
	data := NOTIFYICONDATAW_V1{CbSize: uint32(unsafe.Sizeof(NOTIFYICONDATAW_V1{}))}
	data.HWnd = hwnd
	data.UID = 1
	data.UFlags = NIF_MESSAGE | NIF_TIP | NIF_ICON
	data.HIcon = StockIcon(IDI_APPLICATION)
	data.UCallbackMessage = WM_APP_TRAY
	tip, _ := syscall.UTF16FromString(tooltip)
	if len(tip) > len(data.SzTip)-1 {
		tip = tip[:len(data.SzTip)-1]
	}
	copy(data.SzTip[:], tip)
	return ShellNotifyIcon(message, &data)
}

// PopupMenu builds and shows a simple command menu.
func PopupMenu(hwnd HWND, x, y int32, items []MenuItem) uintptr {
	menu := CreatePopupMenu()
	if menu == 0 {
		return 0
	}
	defer DestroyMenu(menu)
	for _, item := range items {
		if item.Separator {
			_ = AppendMenu(menu, MF_SEPARATOR, 0, "")
			continue
		}
		_ = AppendMenu(menu, MF_STRING, uintptr(item.ID), item.Text)
	}
	return TrackPopupMenu(menu, x, y, hwnd)
}

// MenuItem is one tray menu entry.
type MenuItem struct {
	ID        uintptr
	Text      string
	Separator bool
}

// ShellExecute opens a file or executable through the shell.
func ShellExecute(verb, file, params, dir string, show int32) error {
	r, _, err := procShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(UTF16(verb))), uintptr(unsafe.Pointer(UTF16(file))),
		uintptr(unsafe.Pointer(UTF16(params))), uintptr(unsafe.Pointer(UTF16(dir))), uintptr(show))
	if r <= 32 {
		if errno, ok := err.(syscall.Errno); ok && errno != 0 {
			return LastError(errno)
		}
		return errors.New("shell execute failed")
	}
	return nil
}

// ShellNotifyIcon adds, modifies or removes the tray icon.
func ShellNotifyIcon(message uint32, data *NOTIFYICONDATAW_V1) error {
	r, _, err := procShell_NotifyIconW.Call(uintptr(message), uintptr(unsafe.Pointer(data)))
	if r == 0 {
		return LastError(err)
	}
	return nil
}

// MemoryStatus returns current memory utilisation.
func MemoryStatus() MEMORYSTATUSEX {
	var m MEMORYSTATUSEX
	m.DwLength = uint32(unsafe.Sizeof(m))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m)))
	return m
}

// DiskUsage returns the used percentage of the drive containing path.
func DiskUsage(path string) float64 {
	var free, total, totalFree uint64
	r, _, _ := procGetDiskFreeSpaceExW.Call(uintptr(unsafe.Pointer(UTF16(path))),
		uintptr(unsafe.Pointer(&free)), uintptr(unsafe.Pointer(&total)), uintptr(unsafe.Pointer(&totalFree)))
	if r == 0 || total == 0 {
		return 0
	}
	return float64(total-free) / float64(total) * 100
}

// SystemTimes returns raw idle, kernel and user CPU times (100ns units).
func SystemTimes() (idle, kernel, user uint64, ok bool) {
	var i, k, u FILETIME
	r, _, _ := procGetSystemTimes.Call(uintptr(unsafe.Pointer(&i)), uintptr(unsafe.Pointer(&k)), uintptr(unsafe.Pointer(&u)))
	if r == 0 {
		return 0, 0, 0, false
	}
	return filetimeToUint(i), filetimeToUint(k), filetimeToUint(u), true
}

func filetimeToUint(ft FILETIME) uint64 {
	return uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
}

// NetworkCounters returns summed receive and transmit octets of active
// non-loopback interfaces by parsing MIB_IFTABLE's documented row layout.
func NetworkCounters() (rx, tx uint64, ok bool) {
	size := uint32(16 * 1024)
	for attempt := 0; attempt < 3; attempt++ {
		buf := make([]byte, size)
		r, _, _ := procGetIfTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)
		if r == ERROR_INSUFFICIENT_BUFFER || r == 232 {
			continue
		}
		if r != 0 || int(size) > len(buf) || size < 8 {
			return 0, 0, false
		}
		return sysmetrics.InterfaceCounters(buf[:size])
	}
	return 0, 0, false
}

// RunningExecutables returns the lower-case file names of currently running processes.
func RunningExecutables() map[string]bool {
	out := make(map[string]bool)
	snap, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snap == ^uintptr(0) || snap == 0 {
		return out
	}
	defer procCloseHandle.Call(uintptr(snap))
	entry := PROCESSENTRY32W{}
	entry.DwSize = uint32(unsafe.Sizeof(entry))
	if r, _, _ := procProcess32FirstW.Call(uintptr(snap), uintptr(unsafe.Pointer(&entry))); r == 0 {
		return out
	}
	for {
		out[syscall.UTF16ToString(entry.SzExeFile[:])] = true
		entry.SzExeFile[0] = 0
		if r, _, _ := procProcess32NextW.Call(uintptr(snap), uintptr(unsafe.Pointer(&entry))); r == 0 {
			break
		}
	}
	return out
}

// MessageBox shows a simple user-facing message.
func MessageBox(hwnd HWND, text, caption string, flags uint32) {
	procMessageBoxW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(UTF16(text))), uintptr(unsafe.Pointer(UTF16(caption))), uintptr(flags))
}

// Monitors describes one display's screen and work rectangles.
type Monitors struct {
	Screen  RECT
	Work    RECT
	Primary bool
}

// monitorList is the callback destination for the duration of EnumMonitors.
// Keeping it package-scoped avoids converting a callback uintptr to a Go
// pointer, which is unsafe and rejected by go vet.
var monitorList *[]Monitors

var monitorCallback = syscall.NewCallback(func(hwnd HWND, hdc HDC, rect *RECT, data uintptr) uintptr {
	if monitorList == nil || rect == nil {
		return 1
	}
	var info MONITORINFO
	info.CbSize = uint32(unsafe.Sizeof(info))
	ok, _, _ := procGetMonitorInfoW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return 1
	}
	*monitorList = append(*monitorList, Monitors{
		Screen:  info.RcMonitor,
		Work:    info.RcWork,
		Primary: info.DwFlags&MONITORINFOF_PRIMARY != 0,
	})
	return 1
})

// EnumMonitors returns all active displays, falling back to the virtual screen.
func EnumMonitors() []Monitors {
	var list []Monitors
	monitorList = &list
	procEnumDisplayMonitors.Call(0, 0, monitorCallback, 0)
	monitorList = nil
	if len(list) > 0 {
		return list
	}
	screen := RECT{
		Left:   int32(GetSystemMetrics(SM_XVIRTUALSCREEN)),
		Top:    int32(GetSystemMetrics(SM_YVIRTUALSCREEN)),
		Right:  int32(GetSystemMetrics(SM_XVIRTUALSCREEN) + GetSystemMetrics(SM_CXVIRTUALSCREEN)),
		Bottom: int32(GetSystemMetrics(SM_YVIRTUALSCREEN) + GetSystemMetrics(SM_CYVIRTUALSCREEN)),
	}
	if screen.Right <= screen.Left || screen.Bottom <= screen.Top {
		screen = RECT{Right: 1920, Bottom: 1080}
	}
	return []Monitors{{Screen: screen, Work: GetWorkArea(), Primary: true}}
}

func abs(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}
