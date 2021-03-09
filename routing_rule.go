package publisher

// RoutingRule represents a routing rule supported by v2rayN.
type RoutingRule struct {
	// domain rule list
	Domain []string `json:"domain,omitempty"`

	// IP rule list
	IP []string `json:"ip,omitempty"`

	// Port rule, e.g. 80, 443, 1000-2000
	Port string `json:"port,omitempty"`

	// protocol rule, e.g. http, tls, bittorrent
	Protocol []string `json:"protocol,omitempty"`

	// outboundTag, e.g. direct, proxy, block
	OutboundTag string `json:"outboundTag"`
}
