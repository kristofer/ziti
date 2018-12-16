package ziti

import (
	"bufio"
	"fmt"
	"log"
	"os"

	termbox "github.com/nsf/termbox-go"
)

/* =================== FILE OPS ==============================*/
/* Load the specified text file into the current buffer return any error*/
func (e *editor) editorOpen(filename string) error {

	found, err := e.indexOfBufferNamed(filename)
	if err == nil {
		e.cb = e.buffers[found]
		return nil
	}
	e.addNewBuffer()
	// open the file filename
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	e.cb.filename = filename
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// does the line contain the newline?
		line := scanner.Text()
		e.editorInsertRow(e.cb.numrows, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	e.cb.dirty = false
	return nil
}

/* Save the currentLine file on disk. Return 0 on success, 1 on error. */
func (e *editor) editorSave() error {
	if e.cb.filename[:1] == "*" {
		mesg := "Save File: %s"
		input, err := e.getInput(mesg)
		if err != nil {
			e.editorSetStatusMessage(mesg, err)
		}
		if input == "" {
			return nil
		}
		e.cb.filename = input
	}
	file, err := os.OpenFile(e.cb.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	totalbytes := 0
	for _, line := range e.cb.rows {
		totalbytes += len(line.runes) + 1
		fmt.Fprintln(w, string(line.runes))
	}
	err = w.Flush()
	e.checkErr(err)
	if err == nil {
		e.editorSetStatusMessage("Saved %d bytes.", totalbytes)
		e.cb.dirty = false
	}
	return err
}

func (e *editor) loadFile() error {
	mesg := "Load File: %s"
	input, err := e.getInput(mesg)
	if err != nil {
		e.editorSetStatusMessage("Load File: %s", err)
	}
	if input == "" {
		return nil
	}
	err = e.editorOpen(input)
	if err != nil {
		e.editorSetStatusMessage("Load File: Error: %s", err)
	}
	return nil
}

func (e *editor) getInput(mesg string) (string, error) {
	input := ""
	for {
		e.editorSetStatusMessage(mesg, input)
		e.editorRefreshScreen(false)
		termbox.SetCursor(len(mesg)+len(input)-1, e.screenrows+1)
		ev := <-e.events
		if ev.Ch != 0 {
			ch := ev.Ch
			input = input + string(ch)
		}
		if ev.Ch == 0 {
			switch ev.Key {
			case termbox.KeyEnter:
				return input, nil
			case termbox.KeyCtrlC:
				e.editorSetStatusMessage("killed.")
				return "", nil
			case termbox.KeyBackspace2, termbox.KeyBackspace:
				if len(input) > 0 {
					input = input[:len(input)-1]
				} else {
					input = ""
				}
			case termbox.KeyCtrlG:
				e.editorSetStatusMessage("")
				return "", nil

			case termbox.KeyEsc:
				e.editorSetStatusMessage("Escape not yet implemented")
				return "", nil

			default:
				e.editorSetStatusMessage(mesg, input)
				e.editorRefreshScreen(false)
				termbox.SetCursor(len(mesg)+len(input)-1, e.screenrows+1)
			}
		}
	}
}
