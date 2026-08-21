package topics

// Topic represents a strongly-typed string for allowed message topics.
type Topic string

const (
	Email           Topic = "EMAIL"
	Order           Topic = "ORDER"
	Drone           Topic = "DRONE"
	OrderCreated    Topic = "order.created"
	OrderCancelled  Topic = "order.cancelled"
	DroneRegistered Topic = "drone.registered"
	DroneReserved   Topic = "drone.reserved"
	DroneAssigned   Topic = "drone.assigned"
)

// IsValid validates whether a string represents a registered topic.
func (t Topic) IsValid() bool {
	switch t {
	case Email, Order, Drone, OrderCreated, OrderCancelled, DroneRegistered, DroneReserved, DroneAssigned:
		return true
	default:
		return false
	}
}
