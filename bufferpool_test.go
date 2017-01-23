package bidirpc

import "testing"

func TestBufferPool(t *testing.T) {
	bp := newBufferPool(1)

	buf0 := bp.Get()
	if buf0 == nil {
		t.Fatal("Get returns nil")
	}
	buf1 := bp.Get()
	if buf1 == nil {
		t.Fatal("Get returns nil")
	}

	bp.Put(buf0)
	bp.Put(buf1)

	buf2 := bp.Get()
	if buf2 != buf0 {
		t.Fatal("buf2 should be equal to buf0")
	}
}
