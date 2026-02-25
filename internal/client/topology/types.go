package topology

import "encoding/json"

// ListResponse represents the paginated response for topologies.
type ListResponse struct {
	Items []Topology `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
	Pages int        `json:"pages"`
}

// Topology represents a topology item.
type Topology struct {
	ID                 string          `json:"id"`
	Ecosystem          string          `json:"ecosystem"`
	Metadata           Metadata        `json:"metadata"`
	State              string          `json:"state"`
	Env                string          `json:"env"`
	Name               *string         `json:"name"`
	SubstitutionNodeID *string         `json:"substitution_node_id"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
	Mock               bool            `json:"mock"`
	GreedyResolution   *bool           `json:"greedy_resolution"`
	Policies           json.RawMessage `json:"policies"`
	Attributes         json.RawMessage `json:"attributes"`
	NodesList          []TopologyNode  `json:"nodes_list"`
}

// Metadata groups template and ownership information.
type Metadata struct {
	Owner                Owner           `json:"owner"`
	TemplateName         string          `json:"template_name"`
	TemplateAuthor       string          `json:"template_author"`
	TemplateVersion      string          `json:"template_version"`
	AdditionalProperties json.RawMessage `json:"additional_properties"`
}

type Owner struct {
	DL                      string   `json:"dl"`
	Po                      string   `json:"po"`
	Ecosystems              []string `json:"ecosystems"`
	ServiceNowAssignmentGrp string   `json:"service_now_assignment_group"`
}

// TopologyNode represents nodes included when nodes=true.
type TopologyNode struct {
	ID            string          `json:"id"`
	Ecosystem     string          `json:"ecosystem"`
	Attributes    json.RawMessage `json:"attributes"`
	Metadata      json.RawMessage `json:"metadata"`
	ToscaID       string          `json:"tosca_id"`
	ToscaName     string          `json:"tosca_name"`
	IsSubstituted bool            `json:"is_substituted"`
	State         string          `json:"state"`
	Env           string          `json:"env"`
}

// Filters captures query parameters for listing topologies.
type Filters struct {
	ID        string
	Name      string
	Ecosystem string
	Mock      *bool
	State     string
	Owner     string
	Env       string
	CreatedAt string
	UpdatedAt string
	Page      int
	Size      int
}
