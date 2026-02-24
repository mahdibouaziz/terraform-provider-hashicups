package topology

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-hashicups/internal/client"
	clienttopo "terraform-provider-hashicups/internal/client/topology"
)

var (
	_ datasource.DataSource              = &topologyDataSource{}
	_ datasource.DataSourceWithConfigure = &topologyDataSource{}
)

func NewTopologyDataSource() datasource.DataSource {
	return &topologyDataSource{}
}

type topologyDataSource struct {
	service *clienttopo.Service
}

type topologyModel struct {
	TopologyID types.String `tfsdk:"topology_id"`
	Nodes      types.Bool   `tfsdk:"nodes"`

	// Topology fields
	ID                  types.String `tfsdk:"id"`
	Ecosystem           types.String `tfsdk:"ecosystem"`
	MetadataJSON        types.String `tfsdk:"metadata_json"`
	ServiceNowGroup     types.String `tfsdk:"service_now_assignment_group"`
	TemplateName        types.String `tfsdk:"template_name"`
	TemplateAuthor      types.String `tfsdk:"template_author"`
	TemplateVersion     types.String `tfsdk:"template_version"`
	AdditionalPropsJSON types.String `tfsdk:"additional_properties_json"`
	State               types.String `tfsdk:"state"`
	Env                 types.String `tfsdk:"env"`
	Name                types.String `tfsdk:"name"`
	SubstitutionNodeID  types.String `tfsdk:"substitution_node_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	Mock                types.Bool   `tfsdk:"mock"`
	GreedyResolution    types.Bool   `tfsdk:"greedy_resolution"`
	PoliciesJSON        types.String `tfsdk:"policies_json"`
	AttributesJSON      types.String `tfsdk:"attributes_json"`
	Owner               types.String `tfsdk:"owner"`

	NodesList []topologyNodeModel `tfsdk:"nodes_list"`
}

type topologyNodeModel struct {
	ID             types.String `tfsdk:"id"`
	Ecosystem      types.String `tfsdk:"ecosystem"`
	AttributesJSON types.String `tfsdk:"attributes_json"`
	MetadataJSON   types.String `tfsdk:"metadata_json"`
	ToscaID        types.String `tfsdk:"tosca_id"`
	ToscaName      types.String `tfsdk:"tosca_name"`
	IsSubstituted  types.Bool   `tfsdk:"is_substituted"`
	State          types.String `tfsdk:"state"`
	Env            types.String `tfsdk:"env"`
}

func (d *topologyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topology"
}

func (d *topologyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"topology_id": schema.StringAttribute{
				Required:    true,
				Description: "Topology ID (UUID).",
			},
			"nodes": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, includes nodes_list in the response.",
			},

			"id":                           schema.StringAttribute{Computed: true},
			"ecosystem":                    schema.StringAttribute{Computed: true},
			"metadata_json":                schema.StringAttribute{Computed: true},
			"service_now_assignment_group": schema.StringAttribute{Computed: true},
			"template_name":                schema.StringAttribute{Computed: true},
			"template_author":              schema.StringAttribute{Computed: true},
			"template_version":             schema.StringAttribute{Computed: true},
			"additional_properties_json":   schema.StringAttribute{Computed: true},
			"state":                        schema.StringAttribute{Computed: true},
			"env":                          schema.StringAttribute{Computed: true},
			"name":                         schema.StringAttribute{Computed: true},
			"substitution_node_id":         schema.StringAttribute{Computed: true},
			"created_at":                   schema.StringAttribute{Computed: true},
			"updated_at":                   schema.StringAttribute{Computed: true},
			"mock":                         schema.BoolAttribute{Computed: true},
			"greedy_resolution":            schema.BoolAttribute{Computed: true},
			"policies_json":                schema.StringAttribute{Computed: true},
			"attributes_json":              schema.StringAttribute{Computed: true},
			"owner":                        schema.StringAttribute{Computed: true},
			"nodes_list": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of nodes returned when nodes=true.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"ecosystem":       schema.StringAttribute{Computed: true},
						"attributes_json": schema.StringAttribute{Computed: true},
						"metadata_json":   schema.StringAttribute{Computed: true},
						"tosca_id":        schema.StringAttribute{Computed: true},
						"tosca_name":      schema.StringAttribute{Computed: true},
						"is_substituted":  schema.BoolAttribute{Computed: true},
						"state":           schema.StringAttribute{Computed: true},
						"env":             schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *topologyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	httpClient, ok := req.ProviderData.(client.HTTPClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected client.HTTPClient.")
		return
	}
	d.service = clienttopo.NewService(httpClient)
}

func (d *topologyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg topologyModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeNodes := false
	if !cfg.Nodes.IsNull() && !cfg.Nodes.IsUnknown() {
		includeNodes = cfg.Nodes.ValueBool()
	}

	topo, err := d.service.Get(ctx, cfg.TopologyID.ValueString(), includeNodes)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read topology", err.Error())
		return
	}

	state := mapTopologyToModel(*topo)
	state.TopologyID = cfg.TopologyID
	state.Nodes = cfg.Nodes

	if includeNodes && len(topo.NodesList) > 0 {
		for _, n := range topo.NodesList {
			state.NodesList = append(state.NodesList, topologyNodeModel{
				ID:             types.StringValue(n.ID),
				Ecosystem:      types.StringValue(n.Ecosystem),
				AttributesJSON: rawToString(n.Attributes),
				MetadataJSON:   rawToString(n.Metadata),
				ToscaID:        types.StringValue(n.ToscaID),
				ToscaName:      types.StringValue(n.ToscaName),
				IsSubstituted:  types.BoolValue(n.IsSubstituted),
				State:          types.StringValue(n.State),
				Env:            types.StringValue(n.Env),
			})
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func mapTopologyToModel(t clienttopo.Topology) topologyModel {
	return topologyModel{
		ID:                  types.StringValue(t.ID),
		Ecosystem:           types.StringValue(t.Ecosystem),
		MetadataJSON:        rawToString(t.Metadata),
		ServiceNowGroup:     types.StringValue(t.ServiceNowAssignment),
		TemplateName:        types.StringValue(t.TemplateName),
		TemplateAuthor:      types.StringValue(t.TemplateAuthor),
		TemplateVersion:     types.StringValue(t.TemplateVersion),
		AdditionalPropsJSON: rawToString(t.AdditionalProperties),
		State:               types.StringValue(t.State),
		Env:                 types.StringValue(t.Env),
		Name:                stringOrNull(t.Name),
		SubstitutionNodeID:  stringOrNull(t.SubstitutionNodeID),
		CreatedAt:           types.StringValue(t.CreatedAt),
		UpdatedAt:           types.StringValue(t.UpdatedAt),
		Mock:                types.BoolValue(t.Mock),
		GreedyResolution:    boolOrNull(t.GreedyResolution),
		PoliciesJSON:        rawToString(t.Policies),
		AttributesJSON:      rawToString(t.Attributes),
		Owner:               types.StringValue(t.Owner),
	}
}

func rawToString(b []byte) types.String {
	if len(b) == 0 || string(b) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

func stringOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func boolOrNull(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}
