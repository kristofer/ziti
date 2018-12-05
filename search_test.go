package ziti

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchForward(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
	//    0123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 12345678
	//    012345678911234567892123456789312345678941234567895123456789612345678971234567898123456789912345678901234567891123456789212345678
	e := editor{}
	e.initEditor()
	//e.editorOpen("foo")
	e.editorInsertRow(e.numrows, s)
	r, c := e.searchForward(0, 0, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 17, c)
	r, c = e.searchForward(0, 5, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 17, c)
	r, c = e.searchForward(0, 5, "elit")
	assert.Equal(t, 0, r)
	assert.Equal(t, 55, c)
	r, c = e.searchForward(0, 17, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 108, c)

	r, c = e.searchForward(0, 31, "sed do")
	assert.Equal(t, 0, r)
	assert.Equal(t, 63, c)

}

func TestSearchBackwards(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
	//    0123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 12345678
	//    0123456789112345678921234567893123456789412345678951234567896123456789712345678981234567899         0         1         2
	e := editor{}
	e.initEditor()
	//e.editorOpen("foo")
	e.editorInsertRow(e.numrows, s)
	r, c := 0, 0
	r, c = e.searchBackwards(0, 10, "Lorem")
	assert.Equal(t, 0, r)
	assert.Equal(t, 0, c)
	r, c = e.searchBackwards(0, 30, "ipsum")
	assert.Equal(t, 0, r)
	assert.Equal(t, 6, c)
	r, c = e.searchBackwards(0, 123, "ali")
	assert.Equal(t, 0, r)
	assert.Equal(t, 116, c)
	r, c = e.searchBackwards(0, 21, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 12, c)
	r, c = e.searchBackwards(0, 99, "elit")
	assert.Equal(t, 0, r)
	assert.Equal(t, 51, c)
	r, c = e.searchBackwards(0, 31, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 12, c)

	r, c = e.searchBackwards(0, 99, "sed do")
	assert.Equal(t, 0, r)
	assert.Equal(t, 57, c)

}
func Test2SearchBackwards(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod "
	u := "tempor incididunt ut labore et dolore magna aliqua. "
	//    0123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 12345678
	//    0123456789112345678921234567893123456789412345678951234567896123456789712345678981234567899         0         1         2
	e := editor{}
	e.initEditor()
	//e.editorOpen("foo")
	e.editorInsertRow(e.numrows, s)
	e.editorInsertRow(e.numrows, u)
	r, c := 0, 0
	r, c = e.searchBackwards(0, 10, "Lorem")
	assert.Equal(t, 0, r)
	assert.Equal(t, 0, c)
	r, c = e.searchBackwards(0, 30, "ipsum")
	assert.Equal(t, 0, r)
	assert.Equal(t, 6, c)
	r, c = e.searchBackwards(1, 50, "ali")
	assert.Equal(t, 1, r)
	assert.Equal(t, 44, c)
	r, c = e.searchBackwards(0, 21, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 12, c)
	r, c = e.searchBackwards(1, 40, "tempor")
	assert.Equal(t, 1, r)
	assert.Equal(t, 0, c)
	r, c = e.searchBackwards(1, 31, "dolor")
	assert.Equal(t, 0, r)
	assert.Equal(t, 12, c)

	r, c = e.searchBackwards(1, 40, "sed do")
	assert.Equal(t, 0, r)
	assert.Equal(t, 57, c)

}
