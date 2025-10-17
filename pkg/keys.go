package pkg

// USB HID Keyboard Scan Codes
// Based on USB HID Usage Tables specification

type Key byte

func (k Key) IsModifier() bool {
	return k == ModLeftCtrl || k == ModLeftShift || k == ModLeftAlt || k == ModLeftGUI || k == ModRightCtrl || k == ModRightShift || k == ModRightAlt || k == ModRightGUI
}

// Letter keys
const (
	KeyA Key = 0x04
	KeyB Key = 0x05
	KeyC Key = 0x06
	KeyD Key = 0x07
	KeyE Key = 0x08
	KeyF Key = 0x09
	KeyG Key = 0x0A
	KeyH Key = 0x0B
	KeyI Key = 0x0C
	KeyJ Key = 0x0D
	KeyK Key = 0x0E
	KeyL Key = 0x0F
	KeyM Key = 0x10
	KeyN Key = 0x11
	KeyO Key = 0x12
	KeyP Key = 0x13
	KeyQ Key = 0x14
	KeyR Key = 0x15
	KeyS Key = 0x16
	KeyT Key = 0x17
	KeyU Key = 0x18
	KeyV Key = 0x19
	KeyW Key = 0x1A
	KeyX Key = 0x1B
	KeyY Key = 0x1C
	KeyZ Key = 0x1D
)

// Number keys
const (
	Key1 Key = 0x1E
	Key2 Key = 0x1F
	Key3 Key = 0x20
	Key4 Key = 0x21
	Key5 Key = 0x22
	Key6 Key = 0x23
	Key7 Key = 0x24
	Key8 Key = 0x25
	Key9 Key = 0x26
	Key0 Key = 0x27
)

// Special keys
const (
	KeyEnter      Key = 0x28
	KeyEscape     Key = 0x29
	KeyBackspace  Key = 0x2A
	KeyTab        Key = 0x2B
	KeySpace      Key = 0x2C
	KeyMinus      Key = 0x2D // - and _
	KeyEqual      Key = 0x2E // = and +
	KeyLeftBrace  Key = 0x2F // [ and {
	KeyRightBrace Key = 0x30 // ] and }
	KeyBackslash  Key = 0x31 // \ and |
	KeyNonUSHash  Key = 0x32 // Non-US # and ~
	KeySemicolon  Key = 0x33 // ; and :
	KeyApostrophe Key = 0x34 // ' and "
	KeyGrave      Key = 0x35 // ` and ~
	KeyComma      Key = 0x36 // , and
	KeyDot        Key = 0x37 // . and >
	KeySlash      Key = 0x38 // / and ?
	KeyCapsLock   Key = 0x39
)

// Function keys
const (
	KeyF1  Key = 0x3A
	KeyF2  Key = 0x3B
	KeyF3  Key = 0x3C
	KeyF4  Key = 0x3D
	KeyF5  Key = 0x3E
	KeyF6  Key = 0x3F
	KeyF7  Key = 0x40
	KeyF8  Key = 0x41
	KeyF9  Key = 0x42
	KeyF10 Key = 0x43
	KeyF11 Key = 0x44
	KeyF12 Key = 0x45
)

// Control keys
const (
	KeyPrintScreen Key = 0x46
	KeyScrollLock  Key = 0x47
	KeyPause       Key = 0x48
	KeyInsert      Key = 0x49
	KeyHome        Key = 0x4A
	KeyPageUp      Key = 0x4B
	KeyDelete      Key = 0x4C
	KeyEnd         Key = 0x4D
	KeyPageDown    Key = 0x4E
	KeyRight       Key = 0x4F
	KeyLeft        Key = 0x50
	KeyDown        Key = 0x51
	KeyUp          Key = 0x52
)

// Numpad keys
const (
	KeyNumLock    Key = 0x53
	KeyKPSlash    Key = 0x54 // Keypad /
	KeyKPAsterisk Key = 0x55 // Keypad *
	KeyKPMinus    Key = 0x56 // Keypad -
	KeyKPPlus     Key = 0x57 // Keypad +
	KeyKPEnter    Key = 0x58 // Keypad Enter
	KeyKP1        Key = 0x59 // Keypad 1 and End
	KeyKP2        Key = 0x5A // Keypad 2 and Down
	KeyKP3        Key = 0x5B // Keypad 3 and PageDown
	KeyKP4        Key = 0x5C // Keypad 4 and Left
	KeyKP5        Key = 0x5D // Keypad 5
	KeyKP6        Key = 0x5E // Keypad 6 and Right
	KeyKP7        Key = 0x5F // Keypad 7 and Home
	KeyKP8        Key = 0x60 // Keypad 8 and Up
	KeyKP9        Key = 0x61 // Keypad 9 and PageUp
	KeyKP0        Key = 0x62 // Keypad 0 and Insert
	KeyKPDot      Key = 0x63 // Keypad . and Delete
)

// Additional keys
const (
	KeyNonUSBackslashKey     = 0x64 // Non-US \ and |
	KeyApplication       Key = 0x65 // Application (Windows Menu key)
	KeyPower             Key = 0x66
	KeyKPEqual           Key = 0x67 // Keypad =
	KeyF13               Key = 0x68
	KeyF14               Key = 0x69
	KeyF15               Key = 0x6A
	KeyF16               Key = 0x6B
	KeyF17               Key = 0x6C
	KeyF18               Key = 0x6D
	KeyF19               Key = 0x6E
	KeyF20               Key = 0x6F
	KeyF21               Key = 0x70
	KeyF22               Key = 0x71
	KeyF23               Key = 0x72
	KeyF24               Key = 0x73
	KeyExecute           Key = 0x74
	KeyHelp              Key = 0x75
	KeyMenu              Key = 0x76
	KeySelect            Key = 0x77
	KeyStop              Key = 0x78
	KeyAgain             Key = 0x79
	KeyUndo              Key = 0x7A
	KeyCut               Key = 0x7B
	KeyCopy              Key = 0x7C
	KeyPaste             Key = 0x7D
	KeyFind              Key = 0x7E
	KeyMute              Key = 0x7F
	KeyVolumeUp          Key = 0x80
	KeyVolumeDown        Key = 0x81
)

// Modifier keys - these are NOT keycodes, they're bit flags for the modifier byte
const (
	ModNone       Key = 0x00
	ModLeftCtrl   Key = 0x01 // Left Control
	ModLeftShift  Key = 0x02 // Left Shift
	ModLeftAlt    Key = 0x04 // Left Alt
	ModLeftGUI    Key = 0x08 // Left GUI (Windows/Command/Super key)
	ModRightCtrl  Key = 0x10 // Right Control
	ModRightShift Key = 0x20 // Right Shift
	ModRightAlt   Key = 0x40 // Right Alt (AltGr)
	ModRightGUI   Key = 0x80 // Right GUI (Windows/Command/Super key)
)

// Convenience aliases
const (
	KeyCtrl    Key = ModLeftCtrl
	KeyShift   Key = ModLeftShift
	KeyAlt     Key = ModLeftAlt
	KeySuper   Key = ModLeftGUI
	KeyMeta    Key = ModLeftGUI
	KeyCommand Key = ModLeftGUI
	KeyOption  Key = ModLeftAlt
)

type JSKeyCode string

func (j JSKeyCode) ToKey() Key {
	return JSCodeToHID[j]
}

var JSCodeToHID = map[JSKeyCode]Key{
	// Letters
	"KeyA": KeyA, "KeyB": KeyB, "KeyC": KeyC, "KeyD": KeyD, "KeyE": KeyE,
	"KeyF": KeyF, "KeyG": KeyG, "KeyH": KeyH, "KeyI": KeyI, "KeyJ": KeyJ,
	"KeyK": KeyK, "KeyL": KeyL, "KeyM": KeyM, "KeyN": KeyN, "KeyO": KeyO,
	"KeyP": KeyP, "KeyQ": KeyQ, "KeyR": KeyR, "KeyS": KeyS, "KeyT": KeyT,
	"KeyU": KeyU, "KeyV": KeyV, "KeyW": KeyW, "KeyX": KeyX, "KeyY": KeyY,
	"KeyZ": KeyZ,

	// Digits (top row numbers)
	"Digit0": Key0, "Digit1": Key1, "Digit2": Key2, "Digit3": Key3, "Digit4": Key4,
	"Digit5": Key5, "Digit6": Key6, "Digit7": Key7, "Digit8": Key8, "Digit9": Key9,

	// Function keys
	"F1": KeyF1, "F2": KeyF2, "F3": KeyF3, "F4": KeyF4,
	"F5": KeyF5, "F6": KeyF6, "F7": KeyF7, "F8": KeyF8,
	"F9": KeyF9, "F10": KeyF10, "F11": KeyF11, "F12": KeyF12,
	"F13": KeyF13, "F14": KeyF14, "F15": KeyF15, "F16": KeyF16,
	"F17": KeyF17, "F18": KeyF18, "F19": KeyF19, "F20": KeyF20,
	"F21": KeyF21, "F22": KeyF22, "F23": KeyF23, "F24": KeyF24,

	// Special keys
	"Enter":        KeyEnter,
	"Escape":       KeyEscape,
	"Backspace":    KeyBackspace,
	"Tab":          KeyTab,
	"Space":        KeySpace,
	"Minus":        KeyMinus,      // -
	"Equal":        KeyEqual,      // =
	"BracketLeft":  KeyLeftBrace,  // [
	"BracketRight": KeyRightBrace, // ]
	"Backslash":    KeyBackslash,  // \
	"Semicolon":    KeySemicolon,  // ;
	"Quote":        KeyApostrophe, // '
	"Backquote":    KeyGrave,      // `
	"Comma":        KeyComma,      // ,
	"Period":       KeyDot,        // .
	"Slash":        KeySlash,      // /
	"CapsLock":     KeyCapsLock,

	// Control keys
	"PrintScreen": KeyPrintScreen,
	"ScrollLock":  KeyScrollLock,
	"Pause":       KeyPause,
	"Insert":      KeyInsert,
	"Home":        KeyHome,
	"PageUp":      KeyPageUp,
	"Delete":      KeyDelete,
	"End":         KeyEnd,
	"PageDown":    KeyPageDown,
	"ArrowRight":  KeyRight,
	"ArrowLeft":   KeyLeft,
	"ArrowDown":   KeyDown,
	"ArrowUp":     KeyUp,

	// Numpad
	"NumLock":        KeyNumLock,
	"NumpadDivide":   KeyKPSlash,    // /
	"NumpadMultiply": KeyKPAsterisk, // *
	"NumpadSubtract": KeyKPMinus,    // -
	"NumpadAdd":      KeyKPPlus,     // +
	"NumpadEnter":    KeyKPEnter,
	"Numpad1":        KeyKP1,
	"Numpad2":        KeyKP2,
	"Numpad3":        KeyKP3,
	"Numpad4":        KeyKP4,
	"Numpad5":        KeyKP5,
	"Numpad6":        KeyKP6,
	"Numpad7":        KeyKP7,
	"Numpad8":        KeyKP8,
	"Numpad9":        KeyKP9,
	"Numpad0":        KeyKP0,
	"NumpadDecimal":  KeyKPDot,   // .
	"NumpadEqual":    KeyKPEqual, // =

	// Modifier keys (these return 0 since they're handled as modifiers)
	"ControlLeft":  KeyCtrl,
	"ShiftLeft":    KeyShift,
	"AltLeft":      KeyAlt,
	"MetaLeft":     KeyMeta, // Windows/Command key
	"ControlRight": KeyCtrl,
	"ShiftRight":   KeyShift,
	"AltRight":     KeyAlt,
	"MetaRight":    KeyMeta,

	// Additional keys
	"ContextMenu":     KeyApplication, // Menu key
	"Power":           KeyPower,
	"Help":            KeyHelp,
	"Undo":            KeyUndo,
	"Cut":             KeyCut,
	"Copy":            KeyCopy,
	"Paste":           KeyPaste,
	"Find":            KeyFind,
	"AudioVolumeMute": KeyMute,
	"AudioVolumeUp":   KeyVolumeUp,
	"AudioVolumeDown": KeyVolumeDown,
}
