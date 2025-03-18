// Code generated by ./cmd/ch-gen-col, DO NOT EDIT.

package proto

// ColFixedStr256 represents FixedStr256 column.
type ColFixedStr256 [][256]byte

// Compile-time assertions for ColFixedStr256.
var (
	_ ColInput  = ColFixedStr256{}
	_ ColResult = (*ColFixedStr256)(nil)
	_ Column    = (*ColFixedStr256)(nil)
)

// Rows returns count of rows in column.
func (c ColFixedStr256) Rows() int {
	return len(c)
}

// Reset resets data in row, preserving capacity for efficiency.
func (c *ColFixedStr256) Reset() {
	*c = (*c)[:0]
}

// Type returns ColumnType of FixedStr256.
func (ColFixedStr256) Type() ColumnType {
	return ColumnTypeFixedString.With("256")
}

// Row returns i-th row of column.
func (c ColFixedStr256) Row(i int) [256]byte {
	return c[i]
}

// Append [256]byte to column.
func (c *ColFixedStr256) Append(v [256]byte) {
	*c = append(*c, v)
}

// Append [256]byte slice to column.
func (c *ColFixedStr256) AppendArr(vs [][256]byte) {
	*c = append(*c, vs...)
}

// LowCardinality returns LowCardinality for FixedStr256.
func (c *ColFixedStr256) LowCardinality() *ColLowCardinality[[256]byte] {
	return &ColLowCardinality[[256]byte]{
		index: c,
	}
}

// Array is helper that creates Array of [256]byte.
func (c *ColFixedStr256) Array() *ColArr[[256]byte] {
	return &ColArr[[256]byte]{
		Data: c,
	}
}

// Nullable is helper that creates Nullable([256]byte).
func (c *ColFixedStr256) Nullable() *ColNullable[[256]byte] {
	return &ColNullable[[256]byte]{
		Values: c,
	}
}

// NewArrFixedStr256 returns new Array(FixedStr256).
func NewArrFixedStr256() *ColArr[[256]byte] {
	return &ColArr[[256]byte]{
		Data: new(ColFixedStr256),
	}
}
