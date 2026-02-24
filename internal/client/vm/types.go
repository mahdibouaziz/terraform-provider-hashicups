package vm

// VM represents a virtual machine returned by the API.
type VM struct {
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

// VMPayload is the request body for creating/updating a VM.
type VMPayload struct {
	Name      string   `json:"name"`
	Image     string   `json:"image"`
	CPU       int      `json:"cpu"`
	MemoryMB  int      `json:"memory_mb"`
	NetworkID string   `json:"network_id,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}
