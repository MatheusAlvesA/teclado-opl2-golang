package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// PointerButton is a button with pointer mouse coursor when enabled.
type PointerButton struct {
	widget.Button
}

// NewPointerButton creates a new button widget with the set label and tap
// handler.
func NewPointerButton(text string, onTapped func()) *PointerButton {
	btn := &PointerButton{}
	btn.ExtendBaseWidget(btn)
	btn.Text = text
	btn.OnTapped = onTapped
	return btn
}

// NewPointerButton creates a new button widget with the set icon and tap
// handler.
func NewPointerIconButton(icon fyne.Resource, onTapped func()) *PointerButton {
	btn := &PointerButton{}
	btn.ExtendBaseWidget(btn)
	btn.Icon = icon
	btn.OnTapped = onTapped
	return btn
}

// Cursor returns the cursor type of this widget.
func (b *PointerButton) Cursor() desktop.Cursor {
	if !b.Disabled() {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

// --- PointerButton END ---

// ButtomOption é uma definição de botão incluindo label e callback
type ButtomOption struct {
	label string
	cb    func()
}

func setupWindow(
	playNoteByKeyboard func(int),
	releaseNoteByKeyboard func(int),
) fyne.Window {
	a := app.NewWithID("br.com.matheusalves.tecladolindinho")
	w := a.NewWindow("Teclado Lindinho")

	deskCanvas, _ := w.Canvas().(desktop.Canvas)
	deskCanvas.SetOnKeyDown(func(k *fyne.KeyEvent) {
		switch k.Name {
		case "A":
			playNoteByKeyboard(NotaDO)
		case "W":
			playNoteByKeyboard(NotaDOS)
		case "S":
			playNoteByKeyboard(NotaRE)
		case "E":
			playNoteByKeyboard(NotaRES)
		case "D":
			playNoteByKeyboard(NotaMI)
		case "F":
			playNoteByKeyboard(NotaFA)
		case "T":
			playNoteByKeyboard(NotaFAS)
		case "G":
			playNoteByKeyboard(NotaSOL)
		case "Y":
			playNoteByKeyboard(NotaSOLS)
		case "H":
			playNoteByKeyboard(NotaLA)
		case "U":
			playNoteByKeyboard(NotaLAS)
		case "J":
			playNoteByKeyboard(NotaSI)
		}
	})
	deskCanvas.SetOnKeyUp(func(k *fyne.KeyEvent) {
		switch k.Name {
		case "A":
			releaseNoteByKeyboard(NotaDO)
		case "W":
			releaseNoteByKeyboard(NotaDOS)
		case "S":
			releaseNoteByKeyboard(NotaRE)
		case "E":
			releaseNoteByKeyboard(NotaRES)
		case "D":
			releaseNoteByKeyboard(NotaMI)
		case "F":
			releaseNoteByKeyboard(NotaFA)
		case "T":
			releaseNoteByKeyboard(NotaFAS)
		case "G":
			releaseNoteByKeyboard(NotaSOL)
		case "Y":
			releaseNoteByKeyboard(NotaSOLS)
		case "H":
			releaseNoteByKeyboard(NotaLA)
		case "U":
			releaseNoteByKeyboard(NotaLAS)
		case "J":
			releaseNoteByKeyboard(NotaSI)
		}
	})
	w.Resize(fyne.NewSize(500, 600))
	return w
}

func getInfiniteLoadingScreen() fyne.CanvasObject {
	vBox := container.NewVBox(
		container.NewCenter(widget.NewLabel("Carregando...")),
		widget.NewProgressBarInfinite(),
	)

	paddingRetangle := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	paddingRetangle.SetMinSize(fyne.NewSize(1, 100))

	return container.NewBorder(paddingRetangle, nil, nil, nil, vBox)
}

func getCenterMessageScreen(message string, callback func()) fyne.CanvasObject {
	vBox := container.NewVBox(widget.NewLabel(message))
	if callback != nil {
		vBox.Add(NewPointerButton("Tentar Novamente", callback))
	}
	return container.NewCenter(vBox)
}

func getListButtonsScreen(title string, options []ButtomOption) fyne.CanvasObject {
	paddingRetangle := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	paddingRetangle.SetMinSize(fyne.NewSize(1, 20))
	topBorder := container.NewBorder(paddingRetangle, paddingRetangle, nil, nil, container.NewCenter(widget.NewLabel(title)))

	vBox := container.NewVBox()
	for i := 0; i < len(options); i++ {
		vBox.Add(NewPointerButton(options[i].label, options[i].cb))
	}
	return container.NewBorder(topBorder, nil, nil, nil, vBox)
}

func getMainInterfaceControl(
	onVolumeChange func(float64),
	onAttackChange func(float64),
	onDecayChange func(float64),
	onSustainChange func(float64),
	onReleaseChange func(float64),
	playNote func(int),
	chooseInstrument func(int),
	midiInterface MIDIPlayerInterface,
) fyne.CanvasObject {
	volumeLabel := widget.NewLabel("Volume:")
	volumeControl := widget.NewSlider(0, 100)
	volumeControl.Step = 1
	volumeControl.Value = 100
	volumeControl.OnChanged = onVolumeChange
	vBoxControl := container.NewVBox(volumeLabel, volumeControl)

	attackControl := widget.NewSlider(0, 15)
	attackControl.Step = 1
	attackControl.Value = 15
	attackControl.OnChanged = onAttackChange

	decayControl := widget.NewSlider(0, 15)
	decayControl.Step = 1
	decayControl.Value = 15
	decayControl.OnChanged = onDecayChange

	sustainControl := widget.NewSlider(0, 15)
	sustainControl.Step = 1
	sustainControl.Value = 15
	sustainControl.OnChanged = onSustainChange

	releaseControl := widget.NewSlider(0, 15)
	releaseControl.Step = 1
	releaseControl.Value = 15
	releaseControl.OnChanged = onReleaseChange

	hBoxEnvelope := container.NewGridWithColumns(
		4,
		container.NewVBox(widget.NewLabel("Attack:"), attackControl),
		container.NewVBox(widget.NewLabel("Decay:"), decayControl),
		container.NewVBox(widget.NewLabel("Sustain:"), sustainControl),
		container.NewVBox(widget.NewLabel("Release:"), releaseControl),
	)

	hBoxInstruments := container.NewGridWithColumns(
		4,
		NewPointerButton("Piano", func() { chooseInstrument(0) }),
		NewPointerButton("Piano 1", func() { chooseInstrument(1) }),
		NewPointerButton("Triângulo", func() { chooseInstrument(2) }),
		NewPointerButton("Wave", func() { chooseInstrument(3) }),
	)

	hBoxPiano := container.NewGridWithColumns(
		12,
		NewPointerButton("Dó(A)", func() { playNote(NotaDO) }),
		NewPointerButton("Dó S(W)", func() { playNote(NotaDOS) }),
		NewPointerButton("Ré(S)", func() { playNote(NotaRE) }),
		NewPointerButton("Ré S(E)", func() { playNote(NotaRES) }),
		NewPointerButton("Mi(D)", func() { playNote(NotaMI) }),
		NewPointerButton("Fá(F)", func() { playNote(NotaFA) }),
		NewPointerButton("Fá S(T)", func() { playNote(NotaFAS) }),
		NewPointerButton("Sol(G)", func() { playNote(NotaSOL) }),
		NewPointerButton("Sol S(Y)", func() { playNote(NotaSOLS) }),
		NewPointerButton("Lá(H)", func() { playNote(NotaLA) }),
		NewPointerButton("Lá S(U)", func() { playNote(NotaLAS) }),
		NewPointerButton("Si(J)", func() { playNote(NotaSI) }),
	)
	return container.NewGridWithRows(
		5,
		generatePlayMidiControl(midiInterface),
		hBoxInstruments,
		vBoxControl,
		hBoxEnvelope,
		hBoxPiano,
	)
}

func generatePlayMidiControl(midiInterface MIDIPlayerInterface) fyne.CanvasObject {
	opener := dialog.NewFileOpen(func(uc fyne.URIReadCloser, e error) {
		if uc != nil {
			midiInterface.loadFile(uc.URI().Path())
		}
	}, appWindow)
	hBoxEnvelope := container.NewGridWithColumns(
		3,
		NewPointerIconButton(theme.DocumentSaveIcon(), func() { opener.Show() }),
		NewPointerIconButton(theme.MediaPlayIcon(), func() { midiInterface.play() }),
		NewPointerIconButton(theme.MediaStopIcon(), func() {
			midiInterface.stop()
		}),
	)

	return hBoxEnvelope
}
