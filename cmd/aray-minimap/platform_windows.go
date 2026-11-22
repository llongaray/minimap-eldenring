//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/llongaray/minimap-eldenring/internal/core"
	"github.com/llongaray/minimap-eldenring/internal/game"
)

var (
	u32              = syscall.NewLazyDLL("user32.dll")
	g32              = syscall.NewLazyDLL("gdi32.dll")
	k32              = syscall.NewLazyDLL("kernel32.dll")
	registerClass    = u32.NewProc("RegisterClassExW")
	createWindow     = u32.NewProc("CreateWindowExW")
	defWindow        = u32.NewProc("DefWindowProcW")
	showWindow       = u32.NewProc("ShowWindow")
	isWindowVisible  = u32.NewProc("IsWindowVisible")
	setWindowPos     = u32.NewProc("SetWindowPos")
	getForeground    = u32.NewProc("GetForegroundWindow")
	getWindowPID     = u32.NewProc("GetWindowThreadProcessId")
	getKey           = u32.NewProc("GetAsyncKeyState")
	peekMessage      = u32.NewProc("PeekMessageW")
	translateMessage = u32.NewProc("TranslateMessage")
	dispatchMessage  = u32.NewProc("DispatchMessageW")
	setForeground    = u32.NewProc("SetForegroundWindow")
	getDC            = u32.NewProc("GetDC")
	releaseDC        = u32.NewProc("ReleaseDC")
	sysMetric        = u32.NewProc("GetSystemMetrics")
	getCursor        = u32.NewProc("GetCursorPos")
	createBrush      = g32.NewProc("CreateSolidBrush")
	fillRect         = u32.NewProc("FillRect")
	deleteObject     = g32.NewProc("DeleteObject")
	createFont       = g32.NewProc("CreateFontW")
	selectObject     = g32.NewProc("SelectObject")
	setTextColor     = g32.NewProc("SetTextColor")
	setBkMode        = g32.NewProc("SetBkMode")
	textOut          = g32.NewProc("TextOutW")
	ellipse          = g32.NewProc("Ellipse")
	stretchDIB       = g32.NewProc("StretchDIBits")
	setStretchMode   = g32.NewProc("SetStretchBltMode")
	getModule        = k32.NewProc("GetModuleHandleW")
	getTick          = k32.NewProc("GetTickCount64")
	validateRect     = u32.NewProc("ValidateRect")
	setLayered       = u32.NewProc("SetLayeredWindowAttributes")
	screenToClient   = u32.NewProc("ScreenToClient")
	createCompatibleDC = g32.NewProc("CreateCompatibleDC")
	createCompatibleBitmap = g32.NewProc("CreateCompatibleBitmap")
	deleteDC = g32.NewProc("DeleteDC")
	bitBlt = g32.NewProc("BitBlt")
	saveDC = g32.NewProc("SaveDC")
	restoreDC = g32.NewProc("RestoreDC")
	intersectClipRect = g32.NewProc("IntersectClipRect")
)

const (
	wsPopup          = 0x80000000
	wsVisible        = 0x10000000
	wsExTopmost      = 0x8
	wsExTool         = 0x80
	wsExLayered      = 0x80000
	wsExTransparent  = 0x20
	wsExNoActivate   = 0x08000000
	swHide           = 0
	swShow           = 5
	swShowNoActivate = 8
	pmRemove         = 1
	wmPaint          = 0x000F
	wmClose          = 0x10
	wmDestroy        = 2
	wmErase          = 0x0014
	wmChar           = 0x102
	wmKeyDown        = 0x100
	wmLButtonUp      = 0x202
	wmRButtonUp      = 0x205
	wmMouseWheel     = 0x20A
	swpNoActivate    = 0x10
	srccopy          = 0x00CC0020
	dibRGB           = 0
	halftone         = 4
)

type point struct{ X, Y int32 }
type rect struct{ Left, Top, Right, Bottom int32 }
type msg struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             point
	Private        uint32
}
type wndClass struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	Menu, Name                         *uint16
	IconSmall                          uintptr
}
type bitmapHeader struct {
	Size             uint32
	Width, Height    int32
	Planes, BitCount uint16
	Compression      uint32
	ImageSize        uint32
	XPels, YPels     int32
	Used, Important  uint32
}
type bitmapInfo struct {
	Header bitmapHeader
	Colors [4]byte
}
type tile struct {
	Pixels []byte
	W, H   int
}
type bounds struct{ X, Y, W, H int }

var singleton *app

// O scanner de assinaturas pode consumir tempo no executavel real. Seus
// resultados chegam por canal; somente a thread Win32 manipula as janelas.
type readerOutcome struct {
	reader *game.Reader
	err error
}


// Ativada pela build c-shared: a interface vive dentro do eldenring.exe.
var insideGameDLL bool

type app struct {
	dir               string
	store             *core.Store
	settings          core.Settings
	markers           []core.Marker
	cats              []string
	catCount          map[string]int
	tiles             map[string]*tile
	missedTiles       map[string]bool
	mini, full        uintptr
	screenW, screenH  int
	open              bool
	searchOn          bool
	search            string
	focusID           string
	centerX, centerY  float64
	zoom              float64
	player            *core.Point
	raw               game.RawSample
	gamePID           uint32
	reader            *game.Reader
	lastReaderAttempt time.Time
	lastSample        time.Time
	lastF6            bool
	lastActive        bool
	miniShown         bool
	font, fontSmall   uintptr
	message           string
}

func ptr(s string) *uint16       { return syscall.StringToUTF16Ptr(s) }
func lo(v uintptr) int           { return int(uint16(v)) }
func hi(v uintptr) int           { return int(uint16(v >> 16)) }
func rgb(r, g, b uint32) uintptr { return uintptr(r | g<<8 | b<<16) }
func originInt(v uintptr) int    { return int(int16(uint16(v))) }
func (a *app) save() {
	if err := a.store.Put(a.settings); err != nil {
		a.message = "Falha ao gravar: " + err.Error()
	}
}
func (a *app) visibleMarkers() []core.Marker {
	out := make([]core.Marker, 0, 128)
	needle := strings.ToLower(strings.TrimSpace(a.search))
	for _, m := range a.markers {
		if !core.Visible(m, a.settings) {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(m.Name+" "+m.Description+" "+m.Category), needle) {
			continue
		}
		out = append(out, m)
	}
	return out
}
func winproc(hwnd uintptr, message uint32, wparam, lparam uintptr) uintptr {
	a := singleton
	if a != nil && hwnd == a.full {
		switch message {
		case wmChar:
			if a.searchOn {
				r := rune(wparam)
				if r == 8 {
					if len(a.search) > 0 {
						rs := []rune(a.search)
						a.search = string(rs[:len(rs)-1])
					}
				} else if r == 27 {
					a.searchOn = false
				} else if r >= 32 && len([]rune(a.search)) < 90 {
					a.search += string(r)
				}
				return 0
			}
		case wmKeyDown:
			if wparam == 0x1B {
				a.openMap(false)
				return 0
			}
			if wparam == 0xBF || wparam == 0x46 && !a.searchOn {
				a.searchOn = true
				a.search = ""
				return 0
			} // / or F
			if wparam == 0x0D {
				a.searchOn = false
				return 0
			}
			if !a.searchOn {
				switch wparam {
				case 0x31:
					a.settings.Layer = core.Surface
					a.save()
				case 0x32:
					a.settings.Layer = core.Underground
					a.save()
				case 0x33:
					a.settings.Layer = core.Shadow
					a.save()
				case 0xBB, 0x6B:
					b:=a.fullBounds(); a.zoomAt(b.X+b.W/2,b.Y+b.H/2,1.25)
				case 0xBD, 0x6D:
					b:=a.fullBounds(); a.zoomAt(b.X+b.W/2,b.Y+b.H/2,1/1.25)
				case 0x52:
					a.centerX = 5248
					a.centerY = 5248
					a.zoom = 1
				case 0x4C:
					if a.player != nil {
						a.centerX = a.player.X
						a.centerY = a.player.Y
						a.settings.Layer = a.player.Layer
						a.save()
					}
				case 0x4D:
					if a.focusID != "" {
						core.TogglePin(&a.settings, a.focusID)
						a.save()
					}
				case 0x43:
					if a.focusID != "" {
						core.ToggleFound(&a.settings, a.focusID)
						a.save()
					}
				}
			}
			return 0
		case wmMouseWheel:
			// HIWORD(wParam) contem um delta ASSINADO (int16).
			// Antes, 0xFF88 era lido como positivo e ambas direcoes ampliavam.
			delta := int16(uint16(wparam >> 16))
			pt := point{X:int32(originInt(lparam)),Y:int32(originInt(lparam>>16))}
			screenToClient.Call(hwnd,uintptr(unsafe.Pointer(&pt)))
			if delta > 0 { a.zoomAt(int(pt.X),int(pt.Y),1.25) }
			if delta < 0 { a.zoomAt(int(pt.X),int(pt.Y),1/1.25) }
			return 0
		case wmLButtonUp, wmRButtonUp:
			x, y := originInt(lparam), originInt(lparam>>16)
			a.click(x, y, message == wmRButtonUp)
			return 0
		}
	}
	if message == wmErase {
		return 1
	}
	if message == wmPaint {
		validateRect.Call(hwnd, 0)
		return 0
	}
	if message == wmClose {
		if a != nil && hwnd == a.full {
			a.openMap(false)
		}
		return 0
	}
	r, _, _ := defWindow.Call(hwnd, uintptr(message), wparam, lparam)
	return r
}
func (a *app) click(x, y int, right bool) {
	if y >= 38 && y < 75 {
		if x > 270 && x < 480 {
			a.settings.Layer = core.Surface
		}
		if x >= 480 && x < 710 {
			a.settings.Layer = core.Underground
		}
		if x >= 710 && x < 965 {
			a.settings.Layer = core.Shadow
		}
		a.save()
		return
	}
	if x < 258 && y > 124 {
		row := (y - 137) / 27
		if row >= 0 && row < len(a.cats) {
			cat := a.cats[row]
			a.settings.HiddenCategories[cat] = !a.settings.HiddenCategories[cat]
			a.save()
		}
		return
	}
	b := a.fullBounds()
	if x < b.X || y < b.Y || x > b.X+b.W || y > b.Y+b.H {
		return
	}
	scale := float64(b.W) / 10496 * a.zoom
	if scale <= 0 {
		return
	}
	hitX := a.centerX + float64(x-(b.X+b.W/2))/scale
	hitY := a.centerY + float64(y-(b.Y+b.H/2))/scale
	closest := math.Inf(1)
	id := ""
	for _, m := range a.visibleMarkers() {
		d := math.Hypot(m.X-hitX, m.Y-hitY)
		if d < closest {
			closest = d
			id = m.ID
		}
	}
	if closest*scale <= 19 && id != "" {
		a.focusID = id
		if right {
			core.ToggleFound(&a.settings, id)
		} else {
			core.TogglePin(&a.settings, id)
		}
		a.save()
	} else {
		a.centerX = hitX
		a.centerY = hitY
	}
}
func (a *app) showMini(visible bool) {
	if visible == a.miniShown { return }
	if visible { showWindow.Call(a.mini, swShowNoActivate) } else { showWindow.Call(a.mini, swHide) }
	a.miniShown = visible
	log.Printf("HUD visibilidade alterada: %v", visible)
}

func (a *app) zoomAt(x, y int, factor float64) {
	if factor <= 0 { return }
	next := math.Max(.35, math.Min(8, a.zoom*factor))
	if next == a.zoom { return }
	b := a.fullBounds()
	a.centerX,a.centerY = core.ZoomAt(a.centerX,a.centerY,b.X,b.Y,b.W,b.H,a.zoom,next,float64(x),float64(y))
	a.zoom = next
}

func (a *app) openMap(open bool) {
	a.open = open
	if open {
		showWindow.Call(a.full, swShow)
		setForeground.Call(a.full)
	} else {
		showWindow.Call(a.full, swHide)
	}
}
func run(dir string, store *core.Store, markers []core.Marker) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	a := &app{dir: dir, store: store, settings: store.Get(), markers: markers, tiles: map[string]*tile{}, missedTiles: map[string]bool{}, catCount: map[string]int{}, centerX: 5248, centerY: 5248, zoom: 1}
	if insideGameDLL { a.gamePID = uint32(os.Getpid()) }
	log.Printf("Inicializacao Win32: jogoPID=%d emDLL=%v marcadores=%d", a.gamePID, insideGameDLL, len(markers))
	singleton = a
	for _, m := range markers {
		a.catCount[m.Category]++
	}
	for k := range a.catCount {
		a.cats = append(a.cats, k)
	}
	sort.Strings(a.cats)
	a.screenW = int(val(sysMetric.Call(0)))
	a.screenH = int(val(sysMetric.Call(1)))
	if a.screenW < 800 || a.screenH < 600 {
		return fmt.Errorf("resolução de tela inesperada: %dx%d", a.screenW, a.screenH)
	}
	instance, _, _ := getModule.Call(0)
	cls := ptr("ArayMinimapWindows_2026")
	wc := wndClass{Size: uint32(unsafe.Sizeof(wndClass{})), WndProc: syscall.NewCallback(winproc), Instance: instance, Name: cls}
	ret, _, err := registerClass.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		return fmt.Errorf("RegisterClassExW: %v", err)
	}
	miniSize := a.settings.Size
	miniX := a.screenW - miniSize - 25
	miniY := 25
	if strings.HasSuffix(a.settings.Position, "left") {
		miniX = 25
	}
	if strings.HasPrefix(a.settings.Position, "bottom") {
		miniY = a.screenH - miniSize - 25
	}
	a.mini, _, err = createWindow.Call(wsExTopmost|wsExTool|wsExLayered|wsExTransparent|wsExNoActivate, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(ptr("Aray Minimap HUD"))), wsPopup, uintptr(miniX), uintptr(miniY), uintptr(miniSize), uintptr(miniSize), 0, 0, instance, 0)
	if a.mini == 0 {
		return fmt.Errorf("HUD Win32: %v", err)
	}
	setLayered.Call(a.mini, 0, uintptr(a.settings.Opacity*255/100), 2)
	a.full, _, err = createWindow.Call(wsExTopmost|wsExTool, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(ptr("Aray Minimap Mapa F6"))), wsPopup, 0, 0, uintptr(a.screenW), uintptr(a.screenH), 0, 0, instance, 0)
	if a.full == 0 {
		return fmt.Errorf("mapa Win32: %v", err)
	}
	a.font, _, _ = createFont.Call(22, 0, 0, 0, 600, 0, 0, 0, 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	a.fontSmall, _, _ = createFont.Call(17, 0, 0, 0, 500, 0, 0, 0, 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(ptr("Segoe UI"))))
	showWindow.Call(a.full, swHide)
	showWindow.Call(a.mini, swHide)
	log.Printf("Janelas criadas: HUD=0x%x MAPA=0x%x tela=%dx%d opacidade=%d", a.mini, a.full, a.screenW, a.screenH, a.settings.Opacity)
	defer func() {
		if a.reader != nil {
			a.reader.Close()
		}
		if a.font != 0 {
			deleteObject.Call(a.font)
		}
		if a.fontSmall != 0 {
			deleteObject.Call(a.fontSmall)
		}
	}()
	lastDraw := time.Time{}
	lastPIDCheck := time.Time{}
	lastDiagnostic := time.Time{}
	lastStatus := ""
	gameGoneAt := time.Now()
	readerResults := make(chan readerOutcome, 1)
	readerPending := false
	lastHeartbeat := time.Time{}
	firstDraw := false
	log.Printf("Loop Win32 iniciado: HUD pronto para exibicao antes da leitura de memoria")
	firstPump := true
	for {
		// Limitar a drenagem por ciclo: a fila Win32 pode receber mensagens
		// continuamente; nunca deixar HUD e F6 sem CPU por causa disso.
		if firstPump { log.Printf("Primeira passagem: antes de PeekMessageW") }
		var m msg
		dispatched := 0
		for ; dispatched < 32; dispatched++ {
			found, _, _ := peekMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0, pmRemove)
			if found == 0 { break }
			translateMessage.Call(uintptr(unsafe.Pointer(&m)))
			dispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
		}
		if firstPump { log.Printf("Primeira passagem: PeekMessageW concluiu, mensagens=%d", dispatched); firstPump = false }
		now := time.Now()
		// Nunca bloquear o loop de mensagens Win32 para procurar assinaturas.
		// O jogo pode ter dezenas de MB de codigo para verificar.
		select {
		case result := <-readerResults:
			readerPending = false
			if result.err == nil {
				a.reader = result.reader
				a.message = result.reader.Build()
				log.Printf("Leitura de memoria inicializada: %s", a.message)
			} else {
				a.raw = game.RawSample{Status: result.err.Error()}
				log.Printf("Leitura de memoria indisponivel: %v", result.err)
			}
		default:
		}
		if a.reader == nil && !readerPending && now.Sub(a.lastReaderAttempt) > 5*time.Second {
			a.lastReaderAttempt = now
			readerPending = true
			log.Printf("Iniciando busca de assinaturas de memoria em goroutine")
			go func() {
				r, e := game.NewReader()
				readerResults <- readerOutcome{reader: r, err: e}
			}()
		}
		if !insideGameDLL && now.Sub(lastPIDCheck) > 800*time.Millisecond {
			lastPIDCheck = now
			pid, e := game.PID()
			if e == nil && a.gamePID != 0 && a.gamePID != pid {
				if a.reader != nil {
					a.reader.Close()
					a.reader = nil
				}
				a.player = nil
			}
			if e == nil {
				a.gamePID = pid
				gameGoneAt = now
			} else if now.Sub(gameGoneAt) > 8*time.Second {
				log.Printf("O processo Elden Ring encerrou; saindo do overlay")
				return nil
			}
		}
		if a.reader != nil && now.Sub(a.lastSample) > 250*time.Millisecond {
			a.lastSample = now
			a.raw = a.reader.Sample()
			if a.raw.Ready {
				p, e := a.settings.Calibration.Apply(a.raw.X, a.raw.Y, core.Layer(a.raw.Layer), a.raw.Heading)
				if e == nil {
					a.player = &p
				} else {
					a.player = nil
					a.raw.Status = e.Error()
				}
			} else {
				a.player = nil
			}
		}
		if now.Sub(lastDiagnostic) > 2*time.Second {
			lastDiagnostic = now
			if a.raw.Ready {
				if e := saveDiagnostics(a.dir, a.raw); e != nil {
					log.Printf("Falha no diagnóstico: %v", e)
				}
			}
			if a.raw.Status != lastStatus {
				log.Printf("Estado da posição: %s", a.raw.Status)
				lastStatus = a.raw.Status
			}
		}
		fg, _, _ := getForeground.Call()
		var focusPID uint32
		getWindowPID.Call(fg, uintptr(unsafe.Pointer(&focusPID)))
		active := a.gamePID != 0 && (focusPID == a.gamePID || fg == a.full || fg == a.mini)
		if now.Sub(lastHeartbeat) > 5*time.Second {
			lastHeartbeat = now
			visible, _, _ := isWindowVisible.Call(a.mini)
			log.Printf("Heartbeat Win32: ativo=%v foregroundPID=%d jogoPID=%d mapa=%v readerPending=%v readerReady=%v HUD=0x%x visivel=%v", active, focusPID, a.gamePID, a.open, readerPending, a.reader != nil, a.mini, visible != 0)
		}
		if active != a.lastActive {
			log.Printf("Estado de foco: ativo=%v HWND=0x%x PIDForeground=%d PIDJogo=%d", active, fg, focusPID, a.gamePID)
			a.lastActive = active
		}
		if !active && a.open {
			a.openMap(false)
		}
		// Somente mudar a visibilidade quando o estado variar.
		// ShowWindow a cada 18ms causava recomposicao e piscadas.
		a.showMini(active && !a.open)
		f6 := keyDown(a.settings.F6Key)
		if active && f6 && !a.lastF6 {
			log.Printf("Atalho F6 acionado; mapaAberto=%v", !a.open)
			a.openMap(!a.open)
		}
		a.lastF6 = f6
		if now.Sub(lastDraw) > 110*time.Millisecond && active {
			started := time.Now()
			if !firstDraw { log.Printf("Primeiro desenho iniciado; camada=%s, marcadores=%d", a.settings.Layer, len(a.markers)) }
			if a.open {
				a.drawFull()
			} else {
				a.drawMini()
			}
			lastDraw = now
			if !firstDraw {
				log.Printf("Primeiro desenho concluido em %s", time.Since(started))
				firstDraw = true
			}
		}
		time.Sleep(18 * time.Millisecond)
	}
}
func val(x uintptr, _ uintptr, _ error) uintptr { return x }
func keyDown(v int) bool                        { r, _, _ := getKey.Call(uintptr(v)); return r&0x8000 != 0 }
func (a *app) fullBounds() bounds {
	return bounds{X: 270, Y: 82, W: a.screenW - 290, H: a.screenH - 125}
}
func (a *app) miniBounds() bounds {
	return bounds{X: 8, Y: 32, W: a.settings.Size - 16, H: a.settings.Size - 48}
}
func (a *app) fill(hdc uintptr, region rect, r, g, blue uint32) {
	brush, _, _ := createBrush.Call(rgb(r, g, blue))
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&region)), brush)
	deleteObject.Call(brush)
}
func (a *app) rect(hdc uintptr, x, y, w, h int, r, g, b uint32) {
	a.fill(hdc, rect{int32(x), int32(y), int32(x + w), int32(y + h)}, r, g, b)
}
func (a *app) text(hdc uintptr, x, y int, text string, r, g, b uint32, small bool) {
	if text == "" {
		return
	}
	font := a.font
	if small {
		font = a.fontSmall
	}
	old, _, _ := selectObject.Call(hdc, font)
	setBkMode.Call(hdc, 1)
	setTextColor.Call(hdc, rgb(r, g, b))
	arr := utf16.Encode([]rune(text))
	textOut.Call(hdc, uintptr(x), uintptr(y), uintptr(unsafe.Pointer(&arr[0])), uintptr(len(arr)))
	selectObject.Call(hdc, old)
}
func (a *app) circle(hdc uintptr, x, y, radius int, color uintptr) {
	brush, _, _ := createBrush.Call(color)
	old, _, _ := selectObject.Call(hdc, brush)
	ellipse.Call(hdc, uintptr(x-radius), uintptr(y-radius), uintptr(x+radius), uintptr(y+radius))
	selectObject.Call(hdc, old)
	deleteObject.Call(brush)
}
func (a *app) tileData(layer core.Layer, z, tx, ty int) *tile {
	// The original website assets are NOT distributed in this project.
	// Users may import their own local tiles into data/ctiles/.
	base := fmt.Sprintf("%s/base/%d/%d/%d", layer, z, ty, tx)
	if t, ok := a.tiles[base]; ok {
		return t
	}
	if a.missedTiles[base] {
		return nil
	}
	if len(a.tiles) > 256 {
		a.tiles = map[string]*tile{}
	}
	for _, ext := range []string{".png", ".jpg", ".jpeg"} {
		path := filepath.Join(a.dir, "ctiles", filepath.FromSlash(base+ext))
		f, e := os.Open(path)
		if e != nil {
			continue
		}
		img, _, e := image.Decode(f)
		f.Close()
		if e != nil {
			continue
		}
		bd := img.Bounds()
		rgba := image.NewRGBA(image.Rect(0, 0, bd.Dx(), bd.Dy()))
		draw.Draw(rgba, rgba.Bounds(), img, bd.Min, draw.Src)
		pixels := make([]byte, bd.Dx()*bd.Dy()*4)
		for y := 0; y < bd.Dy(); y++ {
			for x := 0; x < bd.Dx(); x++ {
				i := 4 * (y*bd.Dx() + x)
				px := rgba.RGBAAt(x, y)
				pixels[i] = px.B
				pixels[i+1] = px.G
				pixels[i+2] = px.R
				pixels[i+3] = 0
			}
		}
		t := &tile{Pixels: pixels, W: bd.Dx(), H: bd.Dy()}
		a.tiles[base] = t
		return t
	}
	a.missedTiles[base] = true
	return nil
}
func (a *app) drawTile(hdc uintptr, t *tile, x, y, w, h int) {
	if t == nil || w <= 0 || h <= 0 {
		return
	}
	bi := bitmapInfo{Header: bitmapHeader{Size: uint32(unsafe.Sizeof(bitmapHeader{})), Width: int32(t.W), Height: -int32(t.H), Planes: 1, BitCount: 32}}
	setStretchMode.Call(hdc, halftone)
	stretchDIB.Call(hdc, uintptr(x), uintptr(y), uintptr(w), uintptr(h), 0, 0, uintptr(t.W), uintptr(t.H), uintptr(unsafe.Pointer(&t.Pixels[0])), uintptr(unsafe.Pointer(&bi)), dibRGB, srccopy)
}
func (a *app) renderMap(hdc uintptr, b bounds, layer core.Layer, centerX, centerY, zoom float64, mini bool) {
	if b.W<=0 || b.H<=0 {return}
	// Clipar TODOS os tiles e marcadores ao painel do mapa: impedimos que
	// o zoom invada o cabeçalho, lista de filtros e area de status.
	clip,_,_ := saveDC.Call(hdc)
	if clip != 0 {
		intersectClipRect.Call(hdc,uintptr(b.X),uintptr(b.Y),uintptr(b.X+b.W),uintptr(b.Y+b.H))
		defer restoreDC.Call(hdc,clip)
	}
	a.rect(hdc, b.X, b.Y, b.W, b.H, 30, 38, 36)
	if len(a.markers) == 0 {
		a.text(hdc, b.X+8, b.Y+12, "Sem marcadores; importe markers.js", 218, 195, 141, true)
	}
	scale := float64(b.W) / 10496 * zoom
	if mini {
		scale = zoom * 0.45
	}
	if scale < .0001 {
		return
	}
	left := centerX - float64(b.W/2)/scale
	top := centerY - float64(b.H/2)/scale
	// Highest resolution local master tile pyramid available: z6 (41x41 tiles).
	tileZ := 6
	for z := 6; z > 0; z-- {
		if scale*math.Pow(2, float64(6-z)) > 0.5 {
			tileZ = z
			break
		}
	}
	tileSpan := math.Pow(2, float64(6-tileZ)) * 256
	startX := int(math.Max(0, math.Floor(left/tileSpan)))
	startY := int(math.Max(0, math.Floor(top/tileSpan)))
	endX := int(math.Min(math.Ceil(10496/tileSpan), math.Ceil((left+float64(b.W)/scale)/tileSpan)))
	endY := int(math.Min(math.Ceil(10496/tileSpan), math.Ceil((top+float64(b.H)/scale)/tileSpan)))
	// Avoid pathological redraw work on massive regions.
	count := 0
	for ty := startY; ty <= endY; ty++ {
		for tx := startX; tx <= endX; tx++ {
			if count >= 150 {
				break
			}
			count++
			xx := b.X + int((float64(tx)*tileSpan-left)*scale)
			yy := b.Y + int((float64(ty)*tileSpan-top)*scale)
			s := int(math.Ceil(tileSpan*scale)) + 1
			if t := a.tileData(layer, tileZ, tx, ty); t != nil {
				a.drawTile(hdc, t, xx, yy, s, s)
			}
		}
	}
	// Map coordinates are only meaningful if calibrated against this tile set.
	for _, m := range a.visibleMarkers() {
		if m.Master != layer {
			continue
		}
		if mini && (!a.settings.Markers[m.ID].Pinned && math.Hypot(m.X-centerX, m.Y-centerY) > 950) {
			continue
		}
		x := b.X + int((m.X-left)*scale)
		y := b.Y + int((m.Y-top)*scale)
		if x < b.X+8 || x >= b.X+b.W-8 || y < b.Y+8 || y >= b.Y+b.H-8 {
			continue
		}
		f := a.settings.Markers[m.ID]
		rad := 4
		if f.Pinned {
			rad = 7
		}
		color := categoryColor(m.Category)
		if f.Found {
			color = rgb(94, 135, 95)
		}
		if f.Pinned {
			color = rgb(255, 218, 80)
		}
		a.circle(hdc, x, y, rad, color)
		if !mini && f.Pinned {
			a.text(hdc, x+8, y-7, m.Name, 240, 230, 200, true)
		}
	}
	if a.player != nil && a.player.Layer == layer {
		x := b.X + int((a.player.X-left)*scale)
		y := b.Y + int((a.player.Y-top)*scale)
		if x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			a.playerArrow(hdc, x, y, a.player.Heading)
		}
	}
}
func categoryColor(cat string) uintptr {
	switch cat {
	case "grace":
		return rgb(245, 186, 70)
	case "boss":
		return rgb(232, 89, 95)
	case "npc":
		return rgb(195, 142, 248)
	case "weapon":
		return rgb(77, 169, 245)
	case "armor":
		return rgb(178, 189, 209)
	case "talisman":
		return rgb(255, 146, 67)
	case "spell":
		return rgb(142, 117, 247)
	case "golden", "tear", "crystal":
		return rgb(69, 216, 167)
	default:
		return rgb(211, 182, 134)
	}
}
func (a *app) playerArrow(hdc uintptr, x, y int, heading float64) {
	theta := heading * math.Pi / 180
	points := [3]point{
		{X: int32(x + int(15*math.Sin(theta))), Y: int32(y - int(15*math.Cos(theta)))},
		{X: int32(x + int(7*math.Sin(theta+2.5))), Y: int32(y - int(7*math.Cos(theta+2.5)))},
		{X: int32(x + int(7*math.Sin(theta-2.5))), Y: int32(y - int(7*math.Cos(theta-2.5)))},
	}
	brush, _, _ := createBrush.Call(rgb(54, 198, 245))
	old, _, _ := selectObject.Call(hdc, brush)
	g32.NewProc("Polygon").Call(hdc, uintptr(unsafe.Pointer(&points[0])), 3)
	selectObject.Call(hdc, old)
	deleteObject.Call(brush)
}
// drawBuffered desenha um frame inteiro em um DC de memoria e publica via BitBlt.
// Evita o flash entre apagar o fundo e renderizar dezenas de tiles GDI.
func (a *app) drawBuffered(hwnd uintptr,w,h int,paint func(uintptr)) {
	if w<=0 || h<=0 { return }
	dc,_,_:=getDC.Call(hwnd)
	if dc==0 { return }
	defer releaseDC.Call(hwnd,dc)
	mem,_,_:=createCompatibleDC.Call(dc)
	if mem==0 { paint(dc); return }
	defer deleteDC.Call(mem)
	bmp,_,_:=createCompatibleBitmap.Call(dc,uintptr(w),uintptr(h))
	if bmp==0 { paint(dc); return }
	defer deleteObject.Call(bmp)
	old,_,_:=selectObject.Call(mem,bmp)
	if old==0 { paint(dc); return }
	defer selectObject.Call(mem,old)
	paint(mem)
	bitBlt.Call(dc,0,0,uintptr(w),uintptr(h),mem,0,0,srccopy)
}
func (a *app) drawMini() {
	s := a.settings.Size
	a.drawBuffered(a.mini,s,s,func(hdc uintptr){
	a.rect(hdc, 0, 0, s, s, 12, 18, 20)
	a.rect(hdc, 3, 3, s-6, s-6, 49, 56, 49)
	a.text(hdc, 14, 6, "ARAY MINIMAP", 234, 200, 124, true)
	b := a.miniBounds()
	cx, cy := a.centerX, a.centerY
	if a.player != nil && a.player.Layer == a.settings.Layer {
		cx = a.player.X
		cy = a.player.Y
	}
	a.renderMap(hdc, b, a.settings.Layer, cx, cy, a.settings.Zoom, true)
	msg := "Sem posição validada"
	if a.player != nil {
		msg = "Posição sincronizada"
	}
	a.rect(hdc, 3, s-20, s-6, 18, 17, 24, 24)
	a.text(hdc, 10, s-20, msg, 233, 226, 204, true)
	})
}
func (a *app) drawFull() {
	w, h := a.screenW, a.screenH
	a.drawBuffered(a.full,w,h,func(hdc uintptr){
	a.rect(hdc, 0, 0, w, h, 16, 21, 22)
	a.rect(hdc, 0, 0, w, 78, 33, 40, 36)
	a.text(hdc, 24, 15, "ARAY MINIMAP  |  Mapa interativo", 235, 203, 145, false)
	a.text(hdc, 25, 48, "F6: fechar   |   1/2/3: camadas   |   +/-: zoom   |   F: pesquisar", 206, 203, 184, true)
	a.text(hdc, 284, 46, "1 Superfície     2 Subterrâneo     3 Terra das Sombras", 248, 220, 166, true)
	a.rect(hdc, 0, 78, 258, h-78, 24, 31, 29)
	a.text(hdc, 15, 95, fmt.Sprintf("Categorias (%d)", len(a.cats)), 236, 207, 145, false)
	for i, c := range a.cats {
		y := 137 + i*27
		if y > h-140 {
			break
		}
		r, g, b := uint32(217), uint32(216), uint32(194)
		if a.settings.HiddenCategories[c] {
			r, g, b = 107, 109, 103
		}
		a.text(hdc, 15, y, fmt.Sprintf("%s (%d)", core.CategoryPTBR(c), a.catCount[c]), r, g, b, true)
	}
	b := a.fullBounds()
	a.renderMap(hdc, b, a.settings.Layer, a.centerX, a.centerY, a.zoom, false)
	a.rect(hdc, 260, h-42, w-260, 42, 30, 39, 37)
	status := a.raw.Status
	if status == "" {
		status = "Aguardando jogo"
	}
	a.text(hdc, 275, h-36, fmt.Sprintf("%s | %d marcadores | %s", a.settings.Layer.Label(), len(a.markers), status), 235, 219, 178, true)
	a.text(hdc, 275, h-16, "Clique: fixar | botão direito: concluído | R: centralizar | L: jogador | M/C: selecionado", 189, 195, 183, true)
	if a.searchOn {
		a.rect(hdc, 10, h-102, w-30, 45, 42, 45, 40)
		a.text(hdc, 20, h-97, "PESQUISA: "+a.search+"_", 245, 218, 158, false)
	} else if a.search != "" {
		a.text(hdc, 285, 88, "Filtro: "+a.search, 255, 215, 141, true)
	}
	})
}

func saveDiagnostics(dir string, status game.RawSample) error {
	raw, e := json.MarshalIndent(status, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(dir, "diagnostico-posicao.json"), raw, 0600)
}
