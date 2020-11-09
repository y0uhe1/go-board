package main

import (
	"errors"
	"flag"
	"syscall"
	"unsafe"

	"github.com/cwchiu/go-winapi"
)

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	procSetWindowRgn               = user32.NewProc("SetWindowRgn")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procDrawText                   = user32.NewProc("DrawTextW")
)

var (
	board *Board

	hBitmap winapi.HBITMAP
	hRgn    winapi.HRGN
	hFont   winapi.HFONT
	rc      winapi.RECT

	white = winapi.RGB(0xFF, 0xFF, 0xFF)
	black = winapi.RGB(0x00, 0x00, 0x00)
)

var text string

func runBoard() int {
	flag.StringVar(&text, "m", "message", "type message to show.")
	flag.Parse()

	hInstance := winapi.GetModuleHandle(nil)

	if registerWindowClass(hInstance) == 0 {
		showErrorMessage(0, "registerWindowClass failed")
		return 1
	}

	if err := initializeInstance(hInstance, winapi.SW_SHOW); err != nil {
		showErrorMessage(0, err.Error())
		return 1
	}

	var msg winapi.MSG
	for winapi.GetMessage(&msg, 0, 0, 0) != 0 {
		winapi.TranslateMessage(&msg)
		winapi.DispatchMessage(&msg)
	}

	finalizeInstance(hInstance)

	return int(msg.WParam)
}

func showErrorMessage(hWnd winapi.HWND, msg string) {
	s, _ := syscall.UTF16PtrFromString(msg)
	t, _ := syscall.UTF16PtrFromString("board")
	winapi.MessageBox(hWnd, s, t, winapi.MB_ICONWARNING|winapi.MB_OK)
}

func registerWindowClass(hInstance winapi.HINSTANCE) winapi.ATOM {
	var wc winapi.WNDCLASSEX

	wc.CbSize = uint32(unsafe.Sizeof(winapi.WNDCLASSEX{}))
	wc.Style = 0
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.CbClsExtra = 0
	wc.CbWndExtra = 0
	wc.HInstance = hInstance
	wc.HIcon = winapi.LoadIcon(hInstance, winapi.MAKEINTRESOURCE(132))
	wc.HCursor = winapi.LoadCursor(0, winapi.MAKEINTRESOURCE(winapi.IDC_HAND))
	wc.HbrBackground = winapi.HBRUSH(winapi.GetStockObject(winapi.WHITE_BRUSH))
	wc.LpszMenuName = nil
	wc.LpszClassName, _ = syscall.UTF16PtrFromString("board")

	return winapi.RegisterClassEx(&wc)
}

func initializeInstance(hInstance winapi.HINSTANCE, nCmdShow int) error {
	var err error
	board, err = makeBoard()
	if err != nil {
		return err
	}

	hFont = winapi.CreateFont(
		15, 0, 0, 0, winapi.FW_NORMAL, 0, 0, 0,
		winapi.ANSI_CHARSET, winapi.OUT_DEVICE_PRECIS,
		winapi.CLIP_DEFAULT_PRECIS, winapi.DEFAULT_QUALITY,
		winapi.VARIABLE_PITCH|winapi.FF_ROMAN, nil)

	pc, _ := syscall.UTF16PtrFromString("board")
	pt, _ := syscall.UTF16PtrFromString("board")
	hWnd := winapi.CreateWindowEx(
		winapi.WS_EX_TOOLWINDOW|winapi.WS_EX_TOPMOST|winapi.WS_EX_NOACTIVATE|winapi.WS_EX_LAYERED,
		pc, pt, winapi.WS_POPUP,
		int32(board.X()),
		int32(board.Y()),
		int32(board.W()),
		int32(board.H()),
		0, 0, hInstance, nil)
	if hWnd == 0 {
		return errors.New("CreateWindowEx failed")
	}

	updateWindowRegion(hWnd)

	procSetLayeredWindowAttributes.Call(uintptr(hWnd), uintptr(white), 255, 0x00001)
	winapi.ShowWindow(hWnd, int32(nCmdShow))
	winapi.SetTimer(hWnd, 1, 50, 0)
	return nil
}

func finalizeInstance(hInstance winapi.HINSTANCE) error {
	winapi.DeleteObject(winapi.HGDIOBJ(hFont))
	winapi.DeleteObject(winapi.HGDIOBJ(hBitmap))
	winapi.DeleteObject(winapi.HGDIOBJ(hRgn))
	return nil
}

func updateWindowRegion(hWnd winapi.HWND) {
	tmp := winapi.CreateRectRgn(0, 0, 0, 0)
	winapi.CombineRgn(tmp, hRgn, 0, winapi.RGN_COPY)
	winapi.SetWindowPos(hWnd, 0, int32(board.X()), int32(board.Y()), 0, 0,
		winapi.SWP_NOSIZE|winapi.SWP_NOZORDER|winapi.SWP_NOOWNERZORDER)
	procSetWindowRgn.Call(uintptr(hWnd), uintptr(tmp), uintptr(1))
	winapi.InvalidateRect(hWnd, nil, false)
}

func paintBoard(hWnd winapi.HWND) {
	var ps winapi.PAINTSTRUCT

	hdc := winapi.BeginPaint(hWnd, &ps)
	hCompatDC := winapi.CreateCompatibleDC(hdc)
	winapi.SelectObject(hCompatDC, winapi.HGDIOBJ(hBitmap))
	winapi.BitBlt(hdc, 0, 0, int32(board.W()), int32(board.H()), hCompatDC, 0, 0, winapi.SRCCOPY)
	winapi.DeleteDC(hCompatDC)
	winapi.EndPaint(hWnd, &ps)
}

func animateBoard(hWnd winapi.HWND) {
	board.Motion()
	updateWindowRegion(hWnd)
}

func clickBoard(hWnd winapi.HWND) {
	if winapi.GetKeyState(winapi.VK_SHIFT) < 0 {
		return
	}
}

func wndProc(hWnd winapi.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case winapi.WM_PAINT:
		paintBoard(hWnd)
	case winapi.WM_TIMER:
		animateBoard(hWnd)
	case winapi.WM_LBUTTONDOWN:
		winapi.PostQuitMessage(0)
	case winapi.WM_DESTROY:
		winapi.PostQuitMessage(0)
	default:
		return winapi.DefWindowProc(hWnd, msg, wParam, lParam)
	}
	return 0
}
