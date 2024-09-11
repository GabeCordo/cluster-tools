package components

type Cache interface {
	Save(data any, expiry ...float64) string
	Swap(identifier string, data any, expiry ...float64) bool
	Get(identifier string) (any, bool)
	Remove(identifier string)
	Clean()
}
