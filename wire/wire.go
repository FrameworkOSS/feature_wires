package wire

import (
	"sync"

	crunch "github.com/superwhiskers/crunch/v3"
)

type Wire struct {
	sync.Mutex
	fields []*wireField
}

type wireField struct {
	field  int64
	values [][]byte
}

func NewWire(in ...[]byte) (w *Wire) {
	w = new(Wire)
	w.fields = make([]*wireField, 0)
	if len(in) == 0 {
		return
	}

	buf := crunch.NewBuffer()
	for i := 0; i < len(in); i++ {
		buf.Grow(int64(len(in[i])))
		buf.WriteBytesNext(in[i])
		buf.SeekByte(0, false)
		for {
			pos := buf.ByteOffset()
			if pos+1 >= int64(buf.ByteCapacity()) {
				break
			}

			field := ReadIV64Next(buf)
			count := ReadIV64Next(buf)

			values := make([][]byte, count)
			for j := int64(0); j < count; j++ {
				size := ReadIV64Next(buf)
				value := make([]byte, 0)
				if size > 0 {
					value = buf.ReadBytesNext(size)
				}
				values[j] = value
			}

			w.addField(field, values...)
		}
		buf.Reset()
	}
	return
}

func (w *Wire) Close() {
	for i := range w.fields {
		for j := range w.fields[i].values {
			w.fields[i].values[j] = nil
		}
		w.fields[i].field = 0
	}
	w.fields = nil
}

func (w *Wire) Bytes() (out []byte) {
	w.Lock()
	defer w.Unlock()

	buf := crunch.NewBuffer()
	for i := 0; i < len(w.fields); i++ {
		f := w.fields[i]
		count := int64(len(f.values))

		//grows automatically
		WriteIV64Next(buf, f.field)
		WriteIV64Next(buf, count)

		for j := int64(0); j < count; j++ {
			value := f.values[j]
			size := int64(len(value))

			WriteIV64Next(buf, size)

			buf.Grow(size) //data
			buf.WriteBytesNext(value)
		}
	}
	out = buf.Bytes()
	buf.Reset()
	return
}

func (w *Wire) AddField(field int64, values ...[]byte) {
	w.Lock()
	defer w.Unlock()
	w.addField(field, values...)
}
func (w *Wire) addField(field int64, values ...[]byte) {
	if w.getField(field) != nil {
		w.addValues(field, values...)
		return
	}

	f := new(wireField)
	f.field = field
	f.values = values
	w.fields = append(w.fields, f)
}

func (w *Wire) GetField(field int64) *wireField {
	w.Lock()
	defer w.Unlock()
	return w.getField(field)
}
func (w *Wire) getField(field int64) *wireField {
	for i := 0; i < len(w.fields); i++ {
		if w.fields[i].field == field {
			return w.fields[i]
		}
	}
	return nil
}

func (w *Wire) GetFields() (fields []int64) {
	w.Lock()
	defer w.Unlock()
	return w.getFields()
}
func (w *Wire) getFields() (fields []int64) {
	fields = make([]int64, len(w.fields))
	for i := 0; i < len(w.fields); i++ {
		fields[i] = w.fields[i].field
	}
	return
}

func (w *Wire) RemoveField(field int64) {
	w.Lock()
	defer w.Unlock()
	w.removeField(field)
}
func (w *Wire) removeField(field int64) {
	nf := make([]*wireField, 0)
	for i := 0; i < len(w.fields); i++ {
		if w.fields[i].field == field {
			continue
		}
		nf = append(nf, w.fields[i])
	}
	w.fields = nf
}

func (w *Wire) AddValues(field int64, values ...[]byte) {
	w.Lock()
	defer w.Unlock()
	w.addValues(field, values...)
}
func (w *Wire) addValues(field int64, values ...[]byte) {
	f := w.getField(field)
	if f == nil {
		w.addField(field, values...)
		return
	}
	f.values = append(f.values, values...)
}

func (w *Wire) GetValue(field int64, value int64) []byte {
	w.Lock()
	defer w.Unlock()
	return w.getValue(field, value)
}
func (w *Wire) getValue(field int64, value int64) []byte {
	f := w.getField(field)
	if f == nil || len(f.values) == 0 || int(value) >= len(f.values) {
		return nil
	}
	return f.values[value]
}

func (w *Wire) GetValues(field int64) [][]byte {
	w.Lock()
	defer w.Unlock()
	return w.getValues(field)
}
func (w *Wire) getValues(field int64) [][]byte {
	f := w.getField(field)
	if f == nil {
		return nil
	}
	return f.values
}

func (w *Wire) RemoveValue(field int64, value int64) {
	w.Lock()
	defer w.Unlock()
	w.removeValue(field, value)
}
func (w *Wire) removeValue(field int64, value int64) {
	for i := 0; i < len(w.fields); i++ {
		if f := w.fields[i]; f.field == field {
			nv := make([][]byte, 0)
			for j := 0; j < len(f.values); j++ {
				if j == int(value) {
					continue
				}
				nv = append(nv, f.values[j])
			}
			f.values = nv
			break
		}
	}
}
