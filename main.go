package main

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
	"golang.org/x/sys/windows"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	enumWindows      = user32.NewProc("EnumWindows")
	getWindowText    = user32.NewProc("GetWindowTextW")
	getWindowTextLen = user32.NewProc("GetWindowTextLengthW")
	getWindowRect    = user32.NewProc("GetWindowRect")
)

const (
	WM_SETTEXT = 0x000C
	WM_KEYDOWN = 0x0100
	WM_KEYUP   = 0x0101
	VK_RETURN  = 0x0D
	VK_3       = 0x33
)

func main() {
	var handles []syscall.Handle

	enumWindowsCallback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		bufLen, _, _ := getWindowTextLen.Call(uintptr(hwnd))
		if bufLen == 0 {
			return 1
		}

		buf := make([]uint16, bufLen+1)
		getWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), bufLen+1)

		handles = append(handles, hwnd)

		return 1
	})

	enumWindows.Call(enumWindowsCallback, 0)

	for _, hwnd := range handles {
		bufLen, _, _ := getWindowTextLen.Call(uintptr(hwnd))
		if bufLen == 0 {
			continue
		}

		buf := make([]uint16, bufLen+1)
		getWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), bufLen+1)

		//fmt.Printf("Window handle: %v, Window title: %v\n", hwnd, syscall.UTF16ToString(buf))
		if strings.Contains(syscall.UTF16ToString(buf), "微信") || strings.Contains(syscall.UTF16ToString(buf), "T00ls安全") {
			fmt.Println("Window handle: ", hwnd)
			hwnd := syscall.Handle(hwnd)

			// 获取窗口标题
			title := getWindowTitle(hwnd)
			fmt.Println("Title:", title)

			// 获取窗口内容
			content := getWindowContent(hwnd)
			fmt.Println("Content:", content)

			// 获取窗口大小
			rect := getWindowRectSize(hwnd)
			fmt.Printf("Window Rect: left=%d, top=%d, right=%d, bottom=%d\n", rect.Left, rect.Top, rect.Right, rect.Bottom)
			width := rect.Right - rect.Left
			height := rect.Bottom - rect.Top
			fmt.Printf("width: %d, height: %d\n", width, height)
			// 微信聊天框最小大小为width: 400, height: 374
			if width == 400 && height == 374 {
				fmt.Println("find it !")
				// 获取不到聊天窗口中的文本框控件句柄（因为微信聊天界面用的DirectUI渲染的获取不到）
				sendmessage(hwnd, rect)
				break
			}
			fmt.Println("------------------------------------")
		}
	}
}

// 获取窗口标题
func getWindowTitle(hwnd syscall.Handle) string {
	const nMaxCount = 256
	var buf [nMaxCount]uint16
	getWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(nMaxCount))
	return syscall.UTF16ToString(buf[:])
}

// 获取窗口内容
func getWindowContent(hwnd syscall.Handle) string {
	length, _, _ := getWindowTextLen.Call(uintptr(hwnd))
	buf := make([]uint16, length+1)
	getWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	return syscall.UTF16ToString(buf)
}

// 获取窗口大小
func getWindowRectSize(hwnd syscall.Handle) *windows.Rect {
	var rect windows.Rect
	getWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))
	return &rect
}

const (
	INPUT_MOUSE          = 0
	INPUT_KEYBOARD       = 1
	INPUT_HARDWARE       = 2
	KEYEVENTF_KEYUP      = 0x0002
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004
)

type MOUSEINPUT struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type KEYBDINPUT struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uint64
}

type HARDWAREINPUT struct {
	uMsg    uint32
	wParamL uint16
	wParamH uint16
}

type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	Mi   MOUSEINPUT
	Hi   HARDWAREINPUT
}

func sendmessage(hwnd syscall.Handle, rect *windows.Rect) {
	// 将焦点设置到目标窗口
	winHandle := win.HWND(hwnd)
	win.SetForegroundWindow(winHandle)

	// 计算单击区域的坐标
	x := rect.Left + 140
	y := rect.Bottom - 59

	// 将鼠标移动到单击区域

	var input INPUT
	input.Type = INPUT_MOUSE
	input.Mi.dx = x //int32(x * 65535 / win.GetSystemMetrics(win.SM_CXSCREEN))
	input.Mi.dy = y //int32(y * 65535 / win.GetSystemMetrics(win.SM_CYSCREEN))
	fmt.Println(input.Mi.dx, input.Mi.dy)

	input.Mi.dwFlags = MOUSEEVENTF_LEFTDOWN
	win.SendInput(1, unsafe.Pointer(&input), int32(unsafe.Sizeof(input)))

	input.Mi.dwFlags = MOUSEEVENTF_LEFTUP
	win.SendInput(1, unsafe.Pointer(&input), int32(unsafe.Sizeof(input)))

	fmt.Println("鼠标单击已模拟")

	time.Sleep(time.Second * 1)

	win.SendMessage(winHandle, win.WM_CHAR, uintptr('3'), 0)

	time.Sleep(time.Second * 1)

	win.SendMessage(winHandle, win.WM_KEYDOWN, uintptr(VK_RETURN), 0)
	win.SendMessage(winHandle, win.WM_KEYUP, uintptr(VK_RETURN), 0)
	fmt.Println("发送消息已模拟")
}
