package node

// ListResponse represents the paginated response for nodes.
type ListResponse struct {
	Items []Node `json:"items"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Pages int    `json:"pages"`
}

// Node represents a node item returned by the API.
type Node struct {
	ID                   string   `json:"id"`
	Metadata             Metadata `json:"metadata"`
	Mock                 bool     `json:"mock"`
	ToscaName            string   `json:"tosca_name"`
	ToscaID              string   `json:"tosca_id"`
	Env                  string   `json:"env"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
	Ecosystem            string   `json:"ecosystem"`
	State                string   `json:"state"`
	IsSubstituted        bool     `json:"is_substituted"`
	ToscaTypes           []string `json:"tosca_types"`
	TopologyID           string   `json:"topology_id"`
	SubstituteTopologyID *string  `json:"substitute_topology_id"`
	Name                 string   `json:"name"`
}

type Metadata struct {
	TemplateAuthor string `json:"template_author"`
}

// Filters captures the query parameters for listing nodes.
type Filters struct {
	Mock          *bool
	Ecosystem     string
	IsSubstituted *bool
	Env           string
	State         string
	ToscaID       string
	ToscaName     string
	ToscaTypes    []string
	Name          string
	CreatedAt     string
	UpdatedAt     string
	Page          int
	Size          int
}
