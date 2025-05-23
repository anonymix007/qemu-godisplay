package main

var keymap = map[string]uint32{
    "space":             0x39,
    "exclam":            0x02,
    "quotedbl":          0x28,
    "numbersign":        0x04,
    "dollar":            0x05,
    "percent":           0x06,
    "ampersand":         0x08,
    "apostrophe":        0x28,
    "parenleft":         0x0a,
    "parenright":        0x0b,
    "asterisk":          0x09,
    "plus":              0x0d,
    "comma":             0x33,
    "minus":             0x0c,
    "period":            0x34,
    "slash":             0x35,
    "0":                 0x0b,
    "1":                 0x02,
    "2":                 0x03,
    "3":                 0x04,
    "4":                 0x05,
    "5":                 0x06,
    "6":                 0x07,
    "7":                 0x08,
    "8":                 0x09,
    "9":                 0x0a,
    "colon":             0x27,
    "semicolon":         0x27,
    "less":              0x33,
    "equal":             0x0d,
    "greater":           0x34,
    "question":          0x35,
    "at":                0x03,
    "A":                 0x1e,
    "B":                 0x30,
    "C":                 0x2e,
    "D":                 0x20,
    "E":                 0x12,
    "F":                 0x21,
    "G":                 0x22,
    "H":                 0x23,
    "I":                 0x17,
    "J":                 0x24,
    "K":                 0x25,
    "L":                 0x26,
    "M":                 0x32,
    "N":                 0x31,
    "O":                 0x18,
    "P":                 0x19,
    "Q":                 0x10,
    "R":                 0x13,
    "S":                 0x1f,
    "T":                 0x14,
    "U":                 0x16,
    "V":                 0x2f,
    "W":                 0x11,
    "X":                 0x2d,
    "Y":                 0x15,
    "Z":                 0x2c,
    "bracketleft":       0x1a,
    "backslash":         0x56,
    "bracketright":      0x1b,
    "asciicircum":       0x07,
    "underscore":        0x73,
    "grave":             0x29,
    "a":                 0x1e,
    "b":                 0x30,
    "c":                 0x2e,
    "d":                 0x20,
    "e":                 0x12,
    "f":                 0x21,
    "g":                 0x22,
    "h":                 0x23,
    "i":                 0x17,
    "j":                 0x24,
    "k":                 0x25,
    "l":                 0x26,
    "m":                 0x32,
    "n":                 0x31,
    "o":                 0x18,
    "p":                 0x19,
    "q":                 0x10,
    "r":                 0x13,
    "s":                 0x1f,
    "t":                 0x14,
    "u":                 0x16,
    "v":                 0x2f,
    "w":                 0x11,
    "x":                 0x2d,
    "y":                 0x15,
    "z":                 0x2c,
    "braceleft":         0x1a,
    "bar":               0x2b,
    "braceright":        0x1b,
    "asciitilde":        0x29,
    "multiply":          0x37,
    "BackSpace":         0x0e,
    "Tab":               0x0f,
    "Return":            0x1c,
    "Pause":             0xc6,
    "Scroll_Lock":       0x46,
    "Sys_Req":           0x54,
    "Escape":            0x01,
    "Muhenkan":          0x7b,
    "Henkan_Mode":       0x79,
    "Hiragana":          0x77,
    "Katakana":          0x78,
    "Hiragana_Katakana": 0x70,
    "Zenkaku_Hankaku":   0x76,
    "Home":              0xc7,
    "Left":              0xcb,
    "Up":                0xc8,
    "Right":             0xcd,
    "Down":              0xd0,
    "Prior":             0xc9,
    "Next":              0xd1,
    "End":               0xcf,
    "Insert":            0xd2,
    "Help":              0xf5,
    "Num_Lock":          0x45,
    "KP_Enter":          0x9c,
    "KP_Add":            0x4e,
    "KP_Separator":      0x5c,
    "KP_Subtract":       0x4a,
    "KP_Decimal":        0x53,
    "KP_Divide":         0xb5,
    "KP_0":              0x52,
    "KP_1":              0x4f,
    "KP_2":              0x50,
    "KP_3":              0x51,
    "KP_4":              0x4b,
    "KP_5":              0x4c,
    "KP_6":              0x4d,
    "KP_7":              0x47,
    "KP_8":              0x48,
    "KP_9":              0x49,
    "KP_Equal":          0x59,
    "F1":                0x3b,
    "F2":                0x3c,
    "F3":                0x3d,
    "F4":                0x3e,
    "F5":                0x3f,
    "F6":                0x40,
    "F7":                0x41,
    "F8":                0x42,
    "F9":                0x43,
    "F10":               0x44,
    "F11":               0x57,
    "F12":               0x58,
    "Shift_L":           0x2a,
    "Shift_R":           0x36,
    "Control_L":         0x1d,
    "Control_R":         0x9d,
    "Caps_Lock":         0x3a,
    "Meta_L":            0xdb,
    "Meta_R":            0xdc,
    "Super_L":           0xdb,
    "Super_R":           0xdc,
    "Alt_L":             0x38,
    "Alt_R":             0xb8,
    "Delete":            0xd3,
}

