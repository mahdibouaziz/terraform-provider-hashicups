package loadbalancer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-hashicups/internal/client"
	clientlb "terraform-provider-hashicups/internal/client/loadbalancer"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

// NewLoadBalancerResource is a helper function to simplify the provider implementation.
func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

// loadBalancerResourceModel maps the resource schema data.
type loadBalancerResourceModel struct {
	ID          types.String              `tfsdk:"id"`
	LastUpdated types.String              `tfsdk:"last_updated"`
	Name        types.String              `tfsdk:"name"`
	Protocol    types.String              `tfsdk:"protocol"`
	Port        types.Int64               `tfsdk:"port"`
	HealthURL   types.String              `tfsdk:"health_check_path"`
	Targets     []loadBalancerTargetModel `tfsdk:"targets"`
	Status      types.String              `tfsdk:"status"`
	CreatedAt   types.String              `tfsdk:"created_at"`
}

type loadBalancerTargetModel struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}

// loadBalancerResource is the resource implementation.
type loadBalancerResource struct {
	client  client.HTTPClient
	service *clientlb.Service
}

// Metadata returns the resource type name.
func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

// Schema defines the schema for the resource.
func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the load balancer.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the load balancer.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the load balancer.",
				Required:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "Protocol handled by the load balancer (e.g. http, https, tcp).",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Listener port of the load balancer.",
				Required:    true,
			},
			"health_check_path": schema.StringAttribute{
				Description: "Optional health check path or URL.",
				Optional:    true,
			},
			"targets": schema.ListNestedAttribute{
				Description: "Backend targets behind the load balancer.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "IP or DNS of the backend target.",
							Required:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Port of the backend target.",
							Required:    true,
						},
					},
				},
			},
			"status": schema.StringAttribute{
				Description: "Operational status of the load balancer.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp of the load balancer.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.HTTPClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.HTTPClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
	r.service = clientlb.NewService(client)
}

// ImportState allows `terraform import` to work for this resource.
func (r *loadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
