package geecache

type ByteView struct {
	b []byte
}

// Len 返回ByteView的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一个对象拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
