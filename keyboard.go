package main

import (
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"golang.org/x/exp/errors/fmt"
)

const (
	// NotaDO Nota dó
	NotaDO = iota
	// NotaDOS Nota dó sustenido
	NotaDOS = iota
	// NotaRE Nota ré
	NotaRE = iota
	// NotaRES Nota ré sustenido
	NotaRES = iota
	// NotaMI Nota mi
	NotaMI = iota
	// NotaFA Nota fá
	NotaFA = iota
	// NotaFAS Nota fá sustenido
	NotaFAS = iota
	// NotaSOL Nota sol
	NotaSOL = iota
	// NotaSOLS Nota sol sustenido
	NotaSOLS = iota
	// NotaLA Nota lá
	NotaLA = iota
	// NotaLAS Nota lá sustenido
	NotaLAS = iota
	// NotaSI Nota si
	NotaSI = iota
)

// ChannelNoteAssociation representa uma associação entre uma nota sendo tocada em em qual canal
type ChannelNoteAssociation struct {
	free bool
	note int
}

// KeyboardInterfaceControl controla o dispositivo sintetizador
type KeyboardInterfaceControl struct {
	serialCon   io.ReadWriteCloser
	channelOn   [9]ChannelNoteAssociation
	serialLock  sync.Mutex
	channelLock sync.Mutex
}

func (c *KeyboardInterfaceControl) ping() bool {
	if c.serialCon != nil {
		c.serialLock.Lock()
		c.serialCon.Write([]byte("ping:ping;"))
		time.Sleep(time.Millisecond * 1)
		buf := make([]byte, 128)
		n, err := c.serialCon.Read(buf)
		c.serialLock.Unlock()
		if err != nil || string(buf[:n]) != "pong" {
			fmt.Println(err)
			fmt.Println("Resposta ping inválida: [" + string(buf) + "]")
			return false
		}
		return true
	}
	return false
}

func (c *KeyboardInterfaceControl) sendCommand(command string) {
	if c.serialCon != nil {
		c.serialLock.Lock()
		c.serialCon.Write([]byte(command))
		// Aguardando 1ms para garantir que deu tempo para o arduino processar
		//time.Sleep(time.Millisecond * 1)
		c.serialLock.Unlock()
	}
}

func (c *KeyboardInterfaceControl) setUtilizedChannel(note int) int {
	c.channelLock.Lock()
	i := 0
	for i < len(c.channelOn) && !c.channelOn[i].free {
		i++
	}
	if i >= len(c.channelOn) {
		c.channelLock.Lock()
		return -1
	}
	c.channelOn[i].free = false
	c.channelOn[i].note = note
	c.channelLock.Unlock()
	return i
}

func (c *KeyboardInterfaceControl) setAvailableChannel(note int) int {
	c.channelLock.Lock()
	i := 0
	for i < len(c.channelOn) &&
		(c.channelOn[i].free ||
			c.channelOn[i].note != note) {
		i++
	}
	if i >= len(c.channelOn) {
		c.channelLock.Unlock()
		return -1
	}
	c.channelOn[i].free = true
	c.channelLock.Unlock()
	return i
}

func (c *KeyboardInterfaceControl) playNote(note int) {
	if note < 0 || note > 11 {
		return
	}
	channel := c.setUtilizedChannel(note)
	if channel >= 0 {
		c.sendCommand("playNote:" + strconv.Itoa(channel) + "_" + strconv.Itoa(note+48 /* 4º oitavo */) + ";")
	}
}

func (c *KeyboardInterfaceControl) playNoteOnChannel(channel, note, vel int) {
	if note >= 0 && channel >= 0 {
		c.sendCommand("playNoteVel:" + strconv.Itoa(channel) + "_" + strconv.Itoa(note) + "_" + strconv.Itoa(vel/2) + ";")
	}
}

func (c *KeyboardInterfaceControl) releaseNote(note int) {
	if note < 0 || note > 11 {
		return
	}
	channel := c.setAvailableChannel(note)
	c.sendCommand("releaseNote:" + strconv.Itoa(channel) + ";")
}

func (c *KeyboardInterfaceControl) releaseNoteOnChannel(channel int) {
	c.sendCommand("releaseNote:" + strconv.Itoa(channel) + ";")
}

func (c *KeyboardInterfaceControl) setVolume(volume int) {
	if volume < 0 || volume > 100 {
		return
	}
	c.sendCommand("volume:" + strconv.Itoa(volume) + ";")
}

func (c *KeyboardInterfaceControl) setAttack(value int) {
	if value < 0 || value > 15 {
		return
	}
	c.sendCommand("attack:" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) setDecay(value int) {
	if value < 0 || value > 15 {
		return
	}
	c.sendCommand("decay:" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) setSustain(value int) {
	if value < 0 || value > 15 {
		return
	}
	c.sendCommand("sustain:" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) setRelease(value int) {
	if value < 0 || value > 15 {
		return
	}
	c.sendCommand("release:" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) setInstrument(value int) {
	c.sendCommand("instrument:" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) setInstrumentMidi(channel, value int) {
	c.sendCommand("instrument_midi:" + strconv.Itoa(channel) + "_" + strconv.Itoa(value) + ";")
}

func (c *KeyboardInterfaceControl) close() {
	c.serialLock.Lock()
	c.serialCon.Close()
	c.serialCon = nil
	c.serialLock.Unlock()
}



func openSerialPort(port string) io.ReadWriteCloser {
	config := serial.OpenOptions{
		PortName:        port,
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	c, err := serial.Open(config)

	if err != nil {
		return nil
	}
	return c
}

func listSerialOpeneds() []string {
	r := make([]string, 0)
	com := "COM"

	for i := 0; i < 100; i++ {
		currentCOM := strings.Join([]string{com, strconv.Itoa(i)}, "")
		config := serial.OpenOptions{
			PortName:        currentCOM,
			BaudRate:        9600,
			DataBits:        8,
			StopBits:        1,
			MinimumReadSize: 4,
		}
		c, err := serial.Open(config)
		if err == nil {
			r = append(r, currentCOM)
			c.Close()
		}
	}
	// Aguarde o Arduino reiniciar antes de retornar dessa função
	time.Sleep(time.Second * 3)
	return r
}

func getKeyboardCon(portName string) *KeyboardInterfaceControl {
	port := openSerialPort(portName)
	if port == nil {
		return nil
	}
	time.Sleep(time.Second * 3)

	buf := make([]byte, 128)
	port.Write([]byte("ver:x;"))
	n, err := port.Read(buf)
	if err != nil {
		port.Close()
		return nil
	}

	if string(buf[:n]) == "mte:1.0" {
		return &KeyboardInterfaceControl{
			serialCon: port,
			channelOn: [9]ChannelNoteAssociation{
				{free: true},
				{free: true},
				{free: true},
				{free: true},
				{free: true},
				{free: true},
				{free: true},
				{free: true},
				{free: true},
			},
		}
	}

	fmt.Printf("Resposta inválida do controlador: [%s]\n", string(buf[:n]))
	port.Close()
	return nil
}
