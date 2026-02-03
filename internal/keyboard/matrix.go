package keyboard

import "github.com/bendahl/uinput"

// GPIO pin definitions
const (
	// Column pins (inputs with pull-down)
	C0 = 4
	C1 = 17
	C2 = 18
	C3 = 27
	C4 = 22
	C5 = 23
	C6 = 24
	C7 = 25
	C8 = 5

	// Row pins (outputs)
	R0 = 6
	R1 = 12
	R2 = 13
	R3 = 19
	R4 = 16
	R5 = 26
	R6 = 20
	R7 = 21
)

// PinConfig holds the pin configuration for the keyboard matrix
type PinConfig struct {
	ColumnPins []int
	RowPins    []int
}

// DefaultPinConfig returns the default pin configuration
func DefaultPinConfig() PinConfig {
	return PinConfig{
		ColumnPins: []int{C0, C1, C2, C3, C4, C5, C6, C7, C8},
		RowPins:    []int{R0, R1, R2, R3, R4, R5, R6, R7},
	}
}

// KeyMatrix maps (row, col) to Linux key codes
// Layout matches the TRS-80 Model 100 keyboard matrix
var KeyMatrix = [8][9]int{
	// Row 0
	{
		uinput.KeyZ, uinput.KeyA, uinput.KeyQ, uinput.KeyO,
		uinput.Key1, uinput.Key9, uinput.KeyBackspace, uinput.KeyF1,
		uinput.KeyLeftshift,
	},
	// Row 1
	{
		uinput.KeyX, uinput.KeyS, uinput.KeyW, uinput.KeyP,
		uinput.Key2, uinput.Key0, uinput.KeyUp, uinput.KeyF2,
		uinput.KeyLeftctrl,
	},
	// Row 2
	{
		uinput.KeyC, uinput.KeyD, uinput.KeyE, uinput.KeyEqual,
		uinput.Key3, uinput.KeySemicolon, uinput.KeyDown, uinput.KeyF3,
		uinput.KeyLeftalt,
	},
	// Row 3
	{
		uinput.KeyV, uinput.KeyF, uinput.KeyR, uinput.KeyBackslash,
		uinput.Key4, uinput.KeyApostrophe, uinput.KeyLeft, uinput.KeyF4,
		uinput.KeyLeftshift, // CODE key position
	},
	// Row 4
	{
		uinput.KeyB, uinput.KeyG, uinput.KeyT, uinput.KeyComma,
		uinput.Key5, uinput.KeyMinus, uinput.KeyRight, uinput.KeyF5,
		uinput.KeyLeftshift,
	},
	// Row 5
	{
		uinput.KeyN, uinput.KeyH, uinput.KeyY, uinput.KeyDot,
		uinput.Key6, uinput.KeyLeftbrace, uinput.KeyTab, uinput.KeyF6,
		uinput.KeyLeftshift,
	},
	// Row 6
	{
		uinput.KeyM, uinput.KeyJ, uinput.KeyU, uinput.KeySlash,
		uinput.Key7, uinput.KeySpace, uinput.KeyEsc, uinput.KeyF7,
		uinput.KeyLeftshift,
	},
	// Row 7
	{
		uinput.KeyL, uinput.KeyK, uinput.KeyI, uinput.KeyRightbrace,
		uinput.Key8, uinput.KeyDelete, uinput.KeyEnter, uinput.KeyF11,
		uinput.KeyF12,
	},
}

// Special key positions for modifier detection
const (
	ShiftRow1 = 0
	ShiftCol  = 8
	CtrlRow   = 1
	CtrlCol   = 8
	AltRow    = 2
	AltCol    = 8
	CodeRow   = 3
	CodeCol   = 8
)

// Special CODE key combinations
type CodeKey struct {
	Row int
	Col int
	Key int
}

var CodeKeys = []CodeKey{
	{Row: 6, Col: 5, Key: uinput.KeyPageup},   // CODE + Space = PageUp
	{Row: 7, Col: 5, Key: uinput.KeyPagedown}, // CODE + Delete = PageDown
}
