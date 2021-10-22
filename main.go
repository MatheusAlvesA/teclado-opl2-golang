package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

var availablePorts []string
var appWindow fyne.Window
var interfaceControl *KeyboardInterfaceControl
var lockPort sync.Mutex
var playerMidiInterface MIDIPlayerInterface

func main() {
	playerMidiInterface = MIDIPlayerInterface{events: make([]MIDIEvent, 0), currentTime: 0, pauseFlag: true}

	appWindow = setupWindow(
		func(nota int) {
			if interfaceControl != nil {
				interfaceControl.playNote(nota)
			}
		},
		func(nota int) {
			if interfaceControl != nil {
				interfaceControl.releaseNote(nota)
			}
		},
	)
	go mainControl()
	appWindow.ShowAndRun()
}

func mainControl() {
	loadAvailablePorts()
	if len(availablePorts) > 1 {
		options := make([]ButtomOption, 0)
		for i := 0; i < len(availablePorts); i++ {
			// Declarando localmente para evitar referenciar a variável 'i' que terá valor inválido ao final do loop
			x := i
			options = append(options, ButtomOption{label: availablePorts[i], cb: func() { connectToDevice(availablePorts[x]) }})
		}
		appWindow.SetContent(getListButtonsScreen("Escolha a porta para conectar", options))
	} else if len(availablePorts) == 1 {
		connectToDevice(availablePorts[0])
	} else {
		appWindow.SetContent(getCenterMessageScreen("Nenhum dispositivo conectado", mainControl))
	}
}

func connectToDevice(portName string) {
	appWindow.SetContent(getInfiniteLoadingScreen())
	interfaceControl = getKeyboardCon(portName)
	if interfaceControl != nil {
		appWindow.SetContent(setupMainControlScreen())
		go checkConnectionEngine()
	} else {
		appWindow.SetContent(getCenterMessageScreen("Não foi possível conectar", mainControl))
	}
}

func checkConnectionEngine() {
	for checkConnection() {
		time.Sleep(time.Second * 2)
	}
}

func setupMainControlScreen() fyne.CanvasObject {
	volumeControl := func(vol float64) {
		interfaceControl.setVolume(int(vol))
	}
	attackControl := func(val float64) {
		interfaceControl.setAttack(int(val))
	}
	decayControl := func(val float64) {
		interfaceControl.setDecay(int(val))
	}
	sustainControl := func(val float64) {
		interfaceControl.setSustain(int(val))
	}
	releaseControl := func(val float64) {
		interfaceControl.setRelease(int(val))
	}

	playNoteControl := func(nota int) {
		interfaceControl.playNote(nota)
		time.Sleep(time.Microsecond * 20)
		interfaceControl.releaseNote(nota)
	}

	playerMidiInterface.playCB = func(m MIDIEvent) {
		if m.eventType == "NoteOn" {
			//fmt.Printf("NoteOn (ch %v, key %v, vel %v)\n", m.channel, m.key, m.vel)
			interfaceControl.playNoteOnChannel(int(m.channel), int(m.key), int(m.vel))
		} else {
			//fmt.Printf("NoteOff (ch %v, key %v, vel %v)\n", m.channel, m.key, m.vel)
			interfaceControl.releaseNoteOnChannel(int(m.channel))
		}
		//fmt.Printf("Track: %v Pos: %v %v (ch %v, key %v, vel %v)\n", m.track, m.pos, m.eventType, m.channel, m.key, m.vel)
	}

	playerMidiInterface.instrumentCB = func(channel, instrument uint8) {
		interfaceControl.setInstrumentMidi(int(channel), int(instrument))
		//fmt.Printf("Channel: %v Instrument: %v \n", channel, instrument)
	}

	return getMainInterfaceControl(
		volumeControl,
		attackControl,
		decayControl,
		sustainControl,
		releaseControl,
		playNoteControl,
		interfaceControl.setInstrument,
		playerMidiInterface,
	)
}

func checkConnection() bool {
	lockPort.Lock()
	if interfaceControl != nil {
		if !interfaceControl.ping() {
			// O dipositivo não está respondendo, resetando estado do App
			defer resetAppState()
			not := fyne.NewNotification("Teclado Lindinho", "O dispositivo foi desconectado")
			fyne.CurrentApp().SendNotification(not)
			lockPort.Unlock()
			return false
		}
		lockPort.Unlock()
		return true
	}

	lockPort.Unlock()
	return false
}

func resetAppState() {
	lockPort.Lock()
	availablePorts = make([]string, 0)
	if interfaceControl != nil {
		interfaceControl.close()
		interfaceControl = nil
		defer mainControl()
	}
	lockPort.Unlock()
}

func loadAvailablePorts() {
	appWindow.SetContent(getInfiniteLoadingScreen())
	availablePorts = listSerialOpeneds()
}
