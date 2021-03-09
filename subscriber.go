package publisher

// Subscriber represents a user who subscribes from the publisher.
type Subscriber struct {
	// Remarks is the user alias, which is used for user-defined remarks.
	Remarks string `cfg:"remarks"`

	// Key represents the user's key and is used to get the subscription from the publisher.
	Key string `cfg:"key" validate:"gte=10,lte=32,alphanum"`

	// VMessServers the vmess server ID list of the user's subscription.
	VMessServers []string `cfg:"vmessServers"`

	// RoutingRules the routing rules ID list.
	RoutingRules []string `cfg:"routingRules"`
}
