package ziti

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

/* =================== FILE OPS ==============================*/
/* Load the specified text file into the editor return any error*/
func (e *editor) editorOpen(filename string) error {

	// open the file filename

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	e.filename = filename
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// does the line contain the newline?
		line := scanner.Text()
		e.editorInsertRow(e.numrows, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	e.dirty = false
	return nil
}

/* Save the currentLine file on disk. Return 0 on success, 1 on error. */
func (e *editor) editorSave() error {
	file, err := os.OpenFile(e.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	totalbytes := 0
	for _, line := range e.row {
		totalbytes += len(line.runes) + 1
		fmt.Fprintln(w, string(line.runes))
	}
	err = w.Flush()
	e.checkErr(err)
	if err == nil {
		e.editorSetStatusMessage("Saved %d bytes.", totalbytes)
		e.dirty = false
	}
	return err
}
