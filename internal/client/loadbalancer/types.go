package loadbalancer

// LoadBalancer represents a load balancer returned by the API.
type LoadBalancer struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Protocol  string   `json:"protocol"`
	Port      int      `json:"port"`
	Targets   []Target `json:"targets"`
	HealthURL string   `json:"health_check_path"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

// Target represents a backend target.
type Target struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// Payload is used for create/update requests.
type Payload struct {
	Name      string   `json:"name"`
	Protocol  string   `json:"protocol"`
	Port      int      `json:"port"`
	Targets   []Target `json:"targets"`
	HealthURL string   `json:"health_check_path,omitempty"`
}
