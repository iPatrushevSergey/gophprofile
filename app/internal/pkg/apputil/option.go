package apputil

// Option is a generic functional option type for application constructors.
type Option[T any] func(*T)

// Apply applies all options to the target.
func Apply[T any](target *T, opts ...Option[T]) {
	for _, o := range opts {
		if o != nil {
			o(target)
		}
	}
}
