package nodes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-hashicups/internal/client"
	clientnode "terraform-provider-hashicups/internal/client/node"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &nodesDataSource{}
	_ datasource.DataSourceWithConfigure = &nodesDataSource{}
)

// NewNodesDataSource is a helper function to simplify provider implementation.
func NewNodesDataSource() datasource.DataSource {
	return &nodesDataSource{}
}

type nodesDataSource struct {
	service *clientnode.Service
}

type nodesDataSourceModel struct {
	// Filters
	Mock          types.Bool     `tfsdk:"mock"`
	Ecosystem     types.String   `tfsdk:"ecosystem"`
	IsSubstituted types.Bool     `tfsdk:"is_substituted"`
	Env           types.String   `tfsdk:"env"`
	State         types.String   `tfsdk:"state"`
	ToscaID       types.String   `tfsdk:"tosca_id"`
	ToscaName     types.String   `tfsdk:"tosca_name"`
	ToscaTypes    []types.String `tfsdk:"tosca_types"`
	Name          types.String   `tfsdk:"name"`
	CreatedAt     types.String   `tfsdk:"created_at"`
	UpdatedAt     types.String   `tfsdk:"updated_at"`
	Page          types.Int64    `tfsdk:"page"`
	Size          types.Int64    `tfsdk:"size"`

	// Results
	Items []nodeModel `tfsdk:"items"`
	Total types.Int64 `tfsdk:"total"`
	Pages types.Int64 `tfsdk:"pages"`
}

type nodeModel struct {
	ID                   types.String   `tfsdk:"id"`
	TemplateAuthor       types.String   `tfsdk:"template_author"`
	Mock                 types.Bool     `tfsdk:"mock"`
	ToscaName            types.String   `tfsdk:"tosca_name"`
	ToscaID              types.String   `tfsdk:"tosca_id"`
	Env                  types.String   `tfsdk:"env"`
	CreatedAt            types.String   `tfsdk:"created_at"`
	UpdatedAt            types.String   `tfsdk:"updated_at"`
	Ecosystem            types.String   `tfsdk:"ecosystem"`
	State                types.String   `tfsdk:"state"`
	IsSubstituted        types.Bool     `tfsdk:"is_substituted"`
	ToscaTypes           []types.String `tfsdk:"tosca_types"`
	TopologyID           types.String   `tfsdk:"topology_id"`
	SubstituteTopologyID types.String   `tfsdk:"substitute_topology_id"`
	Name                 types.String   `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *nodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
}

// Schema defines the schema for the data source.
func (d *nodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"mock": schema.BoolAttribute{
				Optional:    true,
				Description: "Filter by mock flag.",
			},
			"ecosystem": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by ecosystem.",
			},
			"is_substituted": schema.BoolAttribute{
				Optional:    true,
				Description: "Filter by substitution state.",
			},
			"env": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by environment.",
			},
			"state": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by state.",
			},
			"tosca_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by TOSCA ID.",
			},
			"tosca_name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by TOSCA name.",
			},
			"tosca_types": schema.ListAttribute{
				Optional:    true,
				Description: "Filter by TOSCA types.",
				ElementType: types.StringType,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by node name.",
			},
			"created_at": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by creation datetime.",
			},
			"updated_at": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by update datetime.",
			},
			"page": schema.Int64Attribute{
				Optional:    true,
				Description: "Page number (default 1).",
			},
			"size": schema.Int64Attribute{
				Optional:    true,
				Description: "Page size (default 50).",
			},
			"items": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of nodes.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"template_author": schema.StringAttribute{Computed: true},
						"mock":            schema.BoolAttribute{Computed: true},
						"tosca_name":      schema.StringAttribute{Computed: true},
						"tosca_id":        schema.StringAttribute{Computed: true},
						"env":             schema.StringAttribute{Computed: true},
						"created_at":      schema.StringAttribute{Computed: true},
						"updated_at":      schema.StringAttribute{Computed: true},
						"ecosystem":       schema.StringAttribute{Computed: true},
						"state":           schema.StringAttribute{Computed: true},
						"is_substituted":  schema.BoolAttribute{Computed: true},
						"tosca_types": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"topology_id":            schema.StringAttribute{Computed: true},
						"substitute_topology_id": schema.StringAttribute{Computed: true},
						"name":                   schema.StringAttribute{Computed: true},
					},
				},
			},
			"total": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of nodes matching the filter.",
			},
			"pages": schema.Int64Attribute{
				Computed:    true,
				Description: "Total number of pages.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *nodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	httpClient, ok := req.ProviderData.(client.HTTPClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected client.HTTPClient.",
		)
		return
	}

	d.service = clientnode.NewService(httpClient)
}

// Read refreshes the Terraform state with the latest data.
func (d *nodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nodesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters := clientnode.Filters{
		Ecosystem: config.Ecosystem.ValueString(),
		Env:       config.Env.ValueString(),
		State:     config.State.ValueString(),
		ToscaID:   config.ToscaID.ValueString(),
		ToscaName: config.ToscaName.ValueString(),
		Name:      config.Name.ValueString(),
		CreatedAt: config.CreatedAt.ValueString(),
		UpdatedAt: config.UpdatedAt.ValueString(),
	}

	if !config.Mock.IsNull() && !config.Mock.IsUnknown() {
		val := config.Mock.ValueBool()
		filters.Mock = &val
	}
	if !config.IsSubstituted.IsNull() && !config.IsSubstituted.IsUnknown() {
		val := config.IsSubstituted.ValueBool()
		filters.IsSubstituted = &val
	}
	if len(config.ToscaTypes) > 0 {
		for _, t := range config.ToscaTypes {
			if !t.IsNull() && !t.IsUnknown() {
				filters.ToscaTypes = append(filters.ToscaTypes, t.ValueString())
			}
		}
	}
	if !config.Page.IsNull() && !config.Page.IsUnknown() && config.Page.ValueInt64() > 0 {
		filters.Page = int(config.Page.ValueInt64())
	} else {
		filters.Page = 1
	}
	if !config.Size.IsNull() && !config.Size.IsUnknown() && config.Size.ValueInt64() > 0 {
		filters.Size = int(config.Size.ValueInt64())
	} else {
		filters.Size = 50
	}

	nodesResp, err := d.service.List(ctx, filters)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list nodes", err.Error())
		return
	}

	state := nodesDataSourceModel{
		Mock:          config.Mock,
		Ecosystem:     config.Ecosystem,
		IsSubstituted: config.IsSubstituted,
		Env:           config.Env,
		State:         config.State,
		ToscaID:       config.ToscaID,
		ToscaName:     config.ToscaName,
		ToscaTypes:    config.ToscaTypes,
		Name:          config.Name,
		CreatedAt:     config.CreatedAt,
		UpdatedAt:     config.UpdatedAt,
		Page:          types.Int64Value(int64(filters.Page)),
		Size:          types.Int64Value(int64(filters.Size)),
		Total:         types.Int64Value(int64(nodesResp.Total)),
		Pages:         types.Int64Value(int64(nodesResp.Pages)),
	}

	for _, n := range nodesResp.Items {
		nodeState := nodeModel{
			ID:             types.StringValue(n.ID),
			TemplateAuthor: types.StringValue(n.Metadata.TemplateAuthor),
			Mock:           types.BoolValue(n.Mock),
			ToscaName:      types.StringValue(n.ToscaName),
			ToscaID:        types.StringValue(n.ToscaID),
			Env:            types.StringValue(n.Env),
			CreatedAt:      types.StringValue(n.CreatedAt),
			UpdatedAt:      types.StringValue(n.UpdatedAt),
			Ecosystem:      types.StringValue(n.Ecosystem),
			State:          types.StringValue(n.State),
			IsSubstituted:  types.BoolValue(n.IsSubstituted),
			TopologyID:     types.StringValue(n.TopologyID),
			Name:           types.StringValue(n.Name),
		}
		if n.SubstituteTopologyID != nil {
			nodeState.SubstituteTopologyID = types.StringValue(*n.SubstituteTopologyID)
		} else {
			nodeState.SubstituteTopologyID = types.StringNull()
		}
		for _, tt := range n.ToscaTypes {
			nodeState.ToscaTypes = append(nodeState.ToscaTypes, types.StringValue(tt))
		}
		state.Items = append(state.Items, nodeState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
