package loadbalancer

// apiLoadBalancer mirrors the load balancer object returned by the HashiCups API.
type apiLoadBalancer struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Protocol  string      `json:"protocol"`
	Port      int         `json:"port"`
	Targets   []apiTarget `json:"targets"`
	HealthURL string      `json:"health_check_path"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
}

type apiTarget struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// apiLoadBalancerPayload is the payload shape expected by the HashiCups load balancer endpoint.
type apiLoadBalancerPayload struct {
	Name      string      `json:"name"`
	Protocol  string      `json:"protocol"`
	Port      int         `json:"port"`
	Targets   []apiTarget `json:"targets"`
	HealthURL string      `json:"health_check_path,omitempty"`
}
