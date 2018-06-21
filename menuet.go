package menuet

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

#ifndef __MENUET_H_H__
#import "menuet.h"
#endif

void setState(const char *jsonString);
void createAndRunApplication();

*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

// ItemType represents what type of menu item this is
type ItemType string

const (
	// Regular is a normal item with text and optional callback
	Regular ItemType = ""
	// Separator is a horizontal line
	Separator = "separator"
	// TODO: StartAtLogin, Quit
)

// FontWeight represents the weight of the font
type FontWeight float64

const (
	// WeightUltraLight is equivalent to NSFontWeightUltraLight
	WeightUltraLight FontWeight = -0.8
	// WeightThin is equivalent to NSFontWeightThin
	WeightThin = -0.6
	// WeightLight is equivalent to NSFontWeightLight
	WeightLight = -0.4
	// WeightRegular is equivalent to NSFontWeightRegular, and is the default
	WeightRegular = 0
	// WeightMedium is equivalent to NSFontWeightMedium
	WeightMedium = 0.23
	// WeightSemibold is equivalent to NSFontWeightSemibold
	WeightSemibold = 0.3
	// WeightBold is equivalent to NSFontWeightBold
	WeightBold = 0.4
	// WeightHeavy is equivalent to NSFontWeightHeavy
	WeightHeavy = 0.56
	// WeightBlack is equivalent to NSFontWeightBlack
	WeightBlack = 0.62
)

// MenuItem represents one item in the dropdown
type MenuItem struct {
	Type ItemType
	// These fields only used for Regular item type:
	Text       string
	FontSize   int // Default: 14
	FontWeight FontWeight
	Callback   string
	State      bool // checkmark if true
	Children   []MenuItem
}

// MenuState represents the title and drop down,
type MenuState struct {
	Title string
	// This is the name of an image in the Resources directory
	Image string
	Items []MenuItem
}

// Application represents the OSX application
type Application struct {
	Name  string
	Label string

	// Clicked receives callbacks of menu items selected
	// It discards messages if the channel is not ready for them
	Clicked    chan<- string
	MenuOpened func() []MenuItem

	// If Version and Repo are set, checks for updates every day
	AutoUpdate struct {
		Version string
		Repo    string // For example "caseymrm/menuet"
	}

	alertChannel       chan AlertClicked
	currentState       *MenuState
	nextState          *MenuState
	pendingStateChange bool
	debounceMutex      sync.Mutex
}

var appInstance *Application
var appOnce sync.Once

// App returns the application singleton
func App() *Application {
	appOnce.Do(func() {
		appInstance = &Application{}
	})
	return appInstance
}

// RunApplication does not return
func (a *Application) RunApplication() {
	if a.AutoUpdate.Version != "" && a.AutoUpdate.Repo != "" {
		go a.checkForUpdates()
	}
	C.createAndRunApplication()
}

// SetMenuState changes what is shown in the dropdown
func (a *Application) SetMenuState(state *MenuState) {
	if reflect.DeepEqual(a.currentState, state) {
		return
	}
	go a.sendState(state)
}

func (a *Application) sendState(state *MenuState) {
	a.debounceMutex.Lock()
	a.nextState = state
	if a.pendingStateChange {
		a.debounceMutex.Unlock()
		return
	}
	a.pendingStateChange = true
	a.debounceMutex.Unlock()
	time.Sleep(100 * time.Millisecond)
	a.debounceMutex.Lock()
	a.pendingStateChange = false
	if reflect.DeepEqual(a.currentState, a.nextState) {
		a.debounceMutex.Unlock()
		return
	}
	a.currentState = a.nextState
	a.debounceMutex.Unlock()
	b, err := json.Marshal(a.currentState)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return
	}
	cstr := C.CString(string(b))
	C.setState(cstr)
	C.free(unsafe.Pointer(cstr))
}

func (a *Application) clicked(callback string) {
	if a.Clicked == nil {
		return
	}
	select {
	case a.Clicked <- callback:
	default:
		fmt.Printf("dropped %s click", callback)
	}
}

func (a *Application) menuOpened() []MenuItem {
	if a.MenuOpened == nil {
		return nil
	}
	return a.MenuOpened()
}

//export itemClicked
func itemClicked(callbackCString *C.char) {
	callback := C.GoString(callbackCString)
	App().clicked(callback)
}

//export menuOpened
func menuOpened() *C.char {
	items := App().menuOpened()
	if items == nil {
		return nil
	}
	b, err := json.Marshal(items)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return nil
	}
	App().currentState.Items = items
	return C.CString(string(b))
}

//export runningAtStartup
func runningAtStartup() bool {
	return App().runningAtStartup()
}

//export toggleStartup
func toggleStartup() {
	a := App()
	if a.runningAtStartup() {
		a.removeStartupItem()
	} else {
		a.addStartupItem()
	}
	go a.sendState(a.currentState)
}
