package namespaces

// CacheNamespace represents a strongly-typed string for allowed Redis cache namespaces.
type CacheNamespace string

const (
	Users     CacheNamespace = "users_ns"
	Sessions  CacheNamespace = "sessions_ns"
	QueueData CacheNamespace = "queue_data_ns"
	Orders    CacheNamespace = "orders_ns"
	Drones    CacheNamespace = "drones_ns"
)

// IsValid validates whether a string represents a registered cache namespace.
func (ns CacheNamespace) IsValid() bool {
	switch ns {
	case Users, Sessions, QueueData, Orders, Drones:
		return true
	default:
		return false
	}
}
