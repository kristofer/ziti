package ziti

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
)

func (e *editor) indexOfBuffer(element *buffer) int {
	for k, v := range e.buffers {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func (e *editor) anyBufferDirty() bool {
	for _, v := range e.buffers {
		if v.dirty {
			return true
		}
	}
	return false
}

func (e *editor) indexOfBufferNamed(name string) (int, error) {
	for k, v := range e.buffers {
		if strings.Compare(v.filename, name) == 0 {
			return k, nil
		}
	}
	return -1, errors.New("not found") //not found.
}

func (e *editor) removeBuffer(element *buffer) {
	i := e.indexOfBuffer(element)
	e.buffers = append(e.buffers[:i], e.buffers[i+1:]...)
}

func (e *editor) nextBuffer() {
	idx := e.indexOfBuffer(e.cb)
	if (idx + 1) >= len(e.buffers) {
		e.cb = e.buffers[0]
		return
	}
	e.cb = e.buffers[idx+1]
}

func (e *editor) addNewBuffer() {
	nb := &buffer{}
	e.buffers = append(e.buffers, nb)
	e.cb = nb
	e.cb.filename = "*scratch*"
}

func (e *editor) listBuffers() {
	found, err := e.indexOfBufferNamed(zitiListBuffers)
	if err == nil {
		e.cb = e.buffers[found]
	} else {
		e.addNewBuffer()
		e.cb.filename = zitiListBuffers
	}
	bufferlist := " ** Current Buffers\n\n"
	for k, v := range e.buffers {
		bufferlist = bufferlist + fmt.Sprintf("%d - %s - %d lines \n", k, v.filename, v.numrows)
	}
	e.cb.rows = nil
	e.cb.numrows = 0
	e.cb.readonly = false
	scanner := bufio.NewScanner(strings.NewReader(bufferlist))
	for scanner.Scan() {
		line := scanner.Text()
		e.editorInsertRow(e.cb.numrows, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	e.cb.dirty = false
	e.cb.readonly = true
}

func (e *editor) killBuffer() {
	if e.cb.dirty {
		e.editorSetStatusMessage("Buffer %s is modified.", e.cb.filename)
		return
	}
	e.removeBuffer(e.cb)
	if len(e.buffers) > 0 {
		e.cb = e.buffers[0]
	}
	if len(e.buffers) == 0 {
		e.done = true
	}
}
