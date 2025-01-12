package bcl

func Ref[T any](v T) *T {
	return &v
}
