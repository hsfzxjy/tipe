package tipe

func (r Result[T]) Alloc() bool {
	return r.field2 == uintptr(kAlloc)
}
