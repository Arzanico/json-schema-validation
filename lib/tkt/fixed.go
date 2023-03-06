package tkt

import (
	"bytes"
	"fmt"
	"math"
	"strings"
)

type Fixed struct {
	value  int64
	scale  int
	factor int
}

func (o *Fixed) Scale() int {
	return o.scale
}

func (o *Fixed) Float64() float64 {
	return float64(o.value) / float64(o.factor)
}

func (o *Fixed) SetFloat64(f float64) *Fixed {
	v := f * float64(o.factor)
	o.value = int64(v)
	return o
}

func (o *Fixed) Add(v Fixed) {
	o.checkScale(v)
	o.value += v.value
}
func (o *Fixed) Sub(v Fixed) {
	o.checkScale(v)
	o.value -= v.value
}

func (o *Fixed) Mult(v Fixed) {
	o.checkScale(v)
	r := o.value * v.value
	o.value = r / int64(o.factor)
}

func (o *Fixed) Div(v Fixed) {
	o.checkScale(v)
	r := float64(o.value) / float64(v.value)
	o.value = int64(r * float64(o.factor))
}

func (o *Fixed) Parse(s string) {
	parts := strings.Split(s, ".")
	o.scale = len(parts[1])
	o.factor = int(math.Pow(10, float64(o.scale)))
	o.value = ParseInt(parts[0] + parts[1])
}

func (o *Fixed) String() string {
	s := fmt.Sprintf("%d", o.value)
	buf := bytes.Buffer{}
	n := 0
	for i := len(s); i < o.scale; i++ {
		buf.WriteByte('0')
		n++
	}
	if len(s)+n == o.scale {
		buf.WriteByte('0')
	}
	buf.WriteString(s)
	s = buf.String()
	buf.Reset()
	buf.WriteString(s[0 : len(s)-o.scale])
	buf.WriteByte('.')
	buf.WriteString(s[len(s)-o.scale:])
	return buf.String()
}

func (o *Fixed) checkScale(v Fixed) {
	if o.scale != v.scale {
		panic("Different scales")
	}
}

func (o *Fixed) MarshalJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	buf.WriteByte('"')
	buf.WriteString(o.String())
	buf.WriteByte('"')
	return buf.Bytes(), nil
}

func (o *Fixed) UnmarshalJSON(b []byte) error {
	s := string(b[1 : len(b)-1])
	o.Parse(s)
	return nil
}

func NewFixed(scale int) *Fixed {
	factor := math.Pow(10, float64(scale))
	return &Fixed{value: 0, scale: scale, factor: int(factor)}
}

func FixedWithValue(scale int, value float64) *Fixed {
	factor := math.Pow(10, float64(scale))
	f := Fixed{value: 0, scale: scale, factor: int(factor)}
	f.SetFloat64(value)
	return &f
}
