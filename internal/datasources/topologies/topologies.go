package topologies

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-hashicups/internal/client"
	clienttopo "terraform-provider-hashicups/internal/client/topology"
)

var (
	_ datasource.DataSource              = &topologiesDataSource{}
	_ datasource.DataSourceWithConfigure = &topologiesDataSource{}
)

func NewTopologiesDataSource() datasource.DataSource {
	return &topologiesDataSource{}
}

type topologiesDataSource struct {
	service *clienttopo.Service
}

type topologiesModel struct {
	// Filters
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Ecosystem types.String `tfsdk:"ecosystem"`
	Mock      types.Bool   `tfsdk:"mock"`
	State     types.String `tfsdk:"state"`
	Owner     types.String `tfsdk:"owner"`
	Env       types.String `tfsdk:"env"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	Page      types.Int64  `tfsdk:"page"`
	Size      types.Int64  `tfsdk:"size"`

	// Results
	Items []topologyModel `tfsdk:"items"`
	Total types.Int64     `tfsdk:"total"`
	Pages types.Int64     `tfsdk:"pages"`
}

type topologyModel struct {
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
}

func (d *topologiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topologies"
}

func (d *topologiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Optional: true, Description: "Filter by topology ID."},
			"name":       schema.StringAttribute{Optional: true, Description: "Filter by name."},
			"ecosystem":  schema.StringAttribute{Optional: true, Description: "Filter by ecosystem."},
			"mock":       schema.BoolAttribute{Optional: true, Description: "Filter by mock flag."},
			"state":      schema.StringAttribute{Optional: true, Description: "Filter by state."},
			"owner":      schema.StringAttribute{Optional: true, Description: "Filter by owner."},
			"env":        schema.StringAttribute{Optional: true, Description: "Filter by environment."},
			"created_at": schema.StringAttribute{Optional: true, Description: "Filter by creation datetime."},
			"updated_at": schema.StringAttribute{Optional: true, Description: "Filter by update datetime."},
			"page":       schema.Int64Attribute{Optional: true, Description: "Page number (default 1)."},
			"size":       schema.Int64Attribute{Optional: true, Description: "Page size (default 50)."},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Topologies list.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
					},
				},
			},
			"total": schema.Int64Attribute{Computed: true},
			"pages": schema.Int64Attribute{Computed: true},
		},
	}
}

func (d *topologiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *topologiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg topologiesModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := clienttopo.Filters{
		ID:        cfg.ID.ValueString(),
		Name:      cfg.Name.ValueString(),
		Ecosystem: cfg.Ecosystem.ValueString(),
		State:     cfg.State.ValueString(),
		Owner:     cfg.Owner.ValueString(),
		Env:       cfg.Env.ValueString(),
		CreatedAt: cfg.CreatedAt.ValueString(),
		UpdatedAt: cfg.UpdatedAt.ValueString(),
	}
	if !cfg.Mock.IsNull() && !cfg.Mock.IsUnknown() {
		v := cfg.Mock.ValueBool()
		filters.Mock = &v
	}
	if !cfg.Page.IsNull() && !cfg.Page.IsUnknown() && cfg.Page.ValueInt64() > 0 {
		filters.Page = int(cfg.Page.ValueInt64())
	} else {
		filters.Page = 1
	}
	if !cfg.Size.IsNull() && !cfg.Size.IsUnknown() && cfg.Size.ValueInt64() > 0 {
		filters.Size = int(cfg.Size.ValueInt64())
	} else {
		filters.Size = 50
	}

	list, err := d.service.List(ctx, filters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list topologies", err.Error())
		return
	}

	state := topologiesModel{
		ID:        cfg.ID,
		Name:      cfg.Name,
		Ecosystem: cfg.Ecosystem,
		Mock:      cfg.Mock,
		State:     cfg.State,
		Owner:     cfg.Owner,
		Env:       cfg.Env,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: cfg.UpdatedAt,
		Page:      types.Int64Value(int64(filters.Page)),
		Size:      types.Int64Value(int64(filters.Size)),
		Total:     types.Int64Value(int64(list.Total)),
		Pages:     types.Int64Value(int64(list.Pages)),
	}

	for _, t := range list.Items {
		state.Items = append(state.Items, mapTopologyToModel(t))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func mapTopologyToModel(t clienttopo.Topology) topologyModel {
	metadataJSON := rawToStringFromStruct(t.Metadata)
	ownerJSON := rawToStringFromStruct(t.Metadata.Owner)
	return topologyModel{
		ID:                  types.StringValue(t.ID),
		Ecosystem:           types.StringValue(t.Ecosystem),
		MetadataJSON:        metadataJSON,
		ServiceNowGroup:     types.StringValue(t.Metadata.ServiceNowAssignment),
		TemplateName:        types.StringValue(t.Metadata.TemplateName),
		TemplateAuthor:      types.StringValue(t.Metadata.TemplateAuthor),
		TemplateVersion:     types.StringValue(t.Metadata.TemplateVersion),
		AdditionalPropsJSON: rawToString(t.Metadata.AdditionalProperties),
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
		Owner:               ownerJSON,
	}
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

func rawToString(b []byte) types.String {
	if len(b) == 0 || string(b) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}

func rawToStringFromStruct(v any) types.String {
	if v == nil {
		return types.StringNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return rawToString(b)
}
