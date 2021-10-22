package main

import (
	"sort"
	"time"

	"gitlab.com/gomidi/midi/reader"
)

type MIDIEvent struct {
	eventType string
	channel   uint8
	key       uint8
	vel       uint8
	track     int16
	pos       uint64
}

type MIDIPlayerInterface struct {
	events       []MIDIEvent
	playCB       func(MIDIEvent)
	instrumentCB func(channel uint8, instrument uint8)
	rd           *reader.Reader
	currentTime  time.Duration
	pauseFlag    bool
}

func (p *MIDIPlayerInterface) orderNotes() {
	sort.Slice(p.events, func(i, j int) bool {
		return p.events[i].pos < p.events[j].pos
	})
}

func (p *MIDIPlayerInterface) play() {
	if p.pauseFlag {
		p.pauseFlag = false
		go p.runPlay()
	}
}

func (p *MIDIPlayerInterface) pause() {
	p.pauseFlag = true
}

func (p *MIDIPlayerInterface) stop() {
	p.pauseFlag = true
	p.currentTime = 0
}

func (p *MIDIPlayerInterface) runPlay() {
	for i := 0; i < len(p.events) && !p.pauseFlag; i++ {
		ev := p.events[i]
		time.Sleep(*reader.TimeAt(p.rd, ev.pos) - p.currentTime)
		p.currentTime = *reader.TimeAt(p.rd, ev.pos)
		if p.playCB != nil {
			p.playCB(ev)
		}
	}
}

func (p *MIDIPlayerInterface) getDuration() time.Duration {
	lastEvent := p.events[len(p.events)-1]
	return *reader.TimeAt(p.rd, lastEvent.pos)
}

func (p *MIDIPlayerInterface) addNoteEvent(pos *reader.Position, channel, key, vel uint8, event string) {
	p.events = append(p.events, MIDIEvent{
		channel:   channel,
		track:     pos.Track,
		key:       key,
		vel:       vel,
		pos:       pos.AbsoluteTicks,
		eventType: event,
	})
}

func (p *MIDIPlayerInterface) loadFile(filePath string) error {
	p.rd = reader.New(reader.NoLogger(),
		reader.NoteOn(func(pos *reader.Position, channel, key, vel uint8) {
			p.addNoteEvent(pos, channel, key, vel, "NoteOn")
		}),
		reader.NoteOff(func(pos *reader.Position, channel, key, vel uint8) {
			p.addNoteEvent(pos, channel, key, vel, "NoteOff")
		}),
		reader.ProgramChange(func(pos *reader.Position, channel, instrument uint8) {
			if p.instrumentCB != nil {
				p.instrumentCB(channel, instrument)
			}
		}),
	)
	p.events = make([]MIDIEvent, 0)
	err := reader.ReadSMFFile(p.rd, filePath)
	if err == nil {
		p.orderNotes()
	}
	return err
}
