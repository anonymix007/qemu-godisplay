package qemu

const (
    qemuIntf = "org.qemu"

    displayIntf = "org.qemu.Display1"
    displayPath = "/org/qemu/Display1"

    vmPath = displayPath + "/VM"
    vmIntf = displayIntf + ".VM"

    vmName       = vmIntf + ".Name"
    vmUUID       = vmIntf + ".UUID"
    vmInterfaces = vmIntf + ".Interfaces"
    vmConsoleIDs = vmIntf + ".ConsoleIDs"

    consolePath = displayPath + "/Console_%d"
    consoleIntf = displayIntf + ".Console"

    consoleType       = consoleIntf + ".Type"
    consoleLabel      = consoleIntf + ".Label"
    consoleWidth      = consoleIntf + ".Width"
    consoleHeight     = consoleIntf + ".Height"
    consoleInterfaces = consoleIntf + ".Interfaces"

    consoleRegisterListener = consoleIntf + ".RegisterListener"

    mouseIntf = displayIntf + ".Mouse"

    mouseIsAbs = mouseIntf + ".IsAbsolute"

    mousePress          = mouseIntf + ".Press"
    mouseRelease        = mouseIntf + ".Release"
    mouseRelMotion      = mouseIntf + ".RelMotion"
    mouseSetAbsPosition = mouseIntf + ".SetAbsPosition"

    keyboardIntf = displayIntf + ".Keyboard"

    keyboardPress     = keyboardIntf + ".Press"
    keyboardRelease   = keyboardIntf + ".Release"
    keyboardModifiers = keyboardIntf + ".Modifiers"

    listenerPath = displayPath + "/Listener"
    listenerIntf = displayIntf + ".Listener"

    listenerUnixMapIntf = listenerIntf + ".Unix.Map"

    listenerUnixScanoutDMABUF2Intf = listenerIntf + ".Unix.ScanoutDMABUF2"
)
