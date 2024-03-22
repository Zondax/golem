package zptr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringToPtr(t *testing.T) {
	s := "hello"
	sp := StringToPtr(s)
	assert.Equal(t, *sp, s)
}

func TestStringOrDefault(t *testing.T) {
	var s *string
	result := StringOrDefault(s)
	assert.Equal(t, "", result)

	hello := "hello"
	s = &hello
	result = StringOrDefault(s)
	assert.Equal(t, hello, result)
}

func TestBoolToPtr(t *testing.T) {
	var b = true
	bp := BoolToPtr(b)
	assert.Equal(t, *bp, b)
}

func TestBoolOrDefault(t *testing.T) {
	var b *bool
	result := BoolOrDefault(b)
	assert.Equal(t, false, result)

	trueVal := true
	b = &trueVal
	result = BoolOrDefault(b)
	assert.Equal(t, true, result)
}

func TestIntToPtr(t *testing.T) {
	i := 42
	ip := IntToPtr(i)
	assert.Equal(t, *ip, i)
}

func TestIntOrDefault(t *testing.T) {
	var i *int
	result := IntOrDefault(i)
	assert.Equal(t, 0, result)

	val := 42
	i = &val
	result = IntOrDefault(i)
	assert.Equal(t, 42, result)
}

func TestFloat32ToPtr(t *testing.T) {
	f := float32(3.14)
	fp := Float32ToPtr(f)
	assert.Equal(t, *fp, f)
}

func TestFloat32OrDefault(t *testing.T) {
	var f *float32
	result := Float32OrDefault(f)
	assert.Equal(t, float32(0.0), result)

	val := float32(3.14159)
	f = &val
	result = Float32OrDefault(f)
	assert.Equal(t, float32(3.14159), result)
}

func TestFloat64ToPtr(t *testing.T) {
	f := 3.14159
	fp := Float64ToPtr(f)
	assert.Equal(t, *fp, f)
}

func TestFloat64OrDefault(t *testing.T) {
	var f *float64
	result := Float64OrDefault(f)
	assert.Equal(t, 0.0, result)

	val := 3.14159
	f = &val
	result = Float64OrDefault(f)
	assert.Equal(t, 3.14159, result)
}
