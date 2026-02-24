package vm

// apiVM mirrors the VM object returned by the HashiCups API.
type apiVM struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	CPU       int      `json:"cpu"`
	MemoryMB  int      `json:"memory_mb"`
	NetworkID string   `json:"network_id"`
	Tags      []string `json:"tags"`

	PrivateIP string `json:"private_ip"`
	PublicIP  string `json:"public_ip"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// apiVMPayload is the payload shape expected by the HashiCups VM endpoint.
type apiVMPayload struct {
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	CPU       int      `json:"cpu"`
	MemoryMB  int      `json:"memory_mb"`
	NetworkID string   `json:"network_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}
