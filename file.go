package ziti

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
)

/* =================== FILE OPS ==============================*/
/* Load the specified program in the editor memory and returns 0 on success
 * or 1 on error. */
func (e *editor) editorOpen(filename string) error {

	// open the file filename
	file, err := os.Open(filename)
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
	buf := e.editorRowsToString()
	err := ioutil.WriteFile(e.filename, []byte(buf), 0644)
	e.checkErr(err)
	if err == nil {
		e.editorSetStatusMessage("Saved %d bytes (%d runes).", len(buf), len([]rune(buf)))
		e.dirty = false
	}
	return err
}
