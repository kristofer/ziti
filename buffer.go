package ziti

func (e *editor) indexOfBuffer(element *buffer) int {
	for k, v := range e.buffers {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func (e *editor) nextBuffer() {
	idx := e.indexOfBuffer(e.cb)
	if (idx + 1) >= len(e.buffers) {
		e.cb = e.buffers[0]
		return
	}
	e.cb = e.buffers[idx+1]
}
