package hata

type argumentIterator struct {
	list    []string
	current int
}

func (it *argumentIterator) next() (value string, ok bool) {
	if it.current < len(it.list) {
		value := it.list[it.current]
		it.current++
		return value, true
	}

	return "", false
}

func (it *argumentIterator) nextValue() (value string, ok bool) {
	value, ok = it.next()
	if !ok {
		return "", false
	}
	if isFlag(value) {
		it.back()
		return "", false
	}
	return value, true
}

func (it *argumentIterator) back() (ok bool) {
	it.current--
	if it.current >= 0 {
		return true
	}

	it.current++
	return false
}

func (it *argumentIterator) remaining() []string {
	return it.list[it.current:]
}
