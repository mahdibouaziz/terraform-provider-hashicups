package vm

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
	clientvm "terraform-provider-hashicups/internal/client/vm"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &vmResource{}
	_ resource.ResourceWithConfigure   = &vmResource{}
	_ resource.ResourceWithImportState = &vmResource{}
)

// NewVMResource is a helper function to simplify the provider implementation.
func NewVMResource() resource.Resource {
	return &vmResource{}
}

// vmResourceModel maps the resource schema data.
type vmResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	LastUpdated types.String   `tfsdk:"last_updated"`
	Name        types.String   `tfsdk:"name"`
	Image       types.String   `tfsdk:"image"`
	CPU         types.Int64    `tfsdk:"cpu"`
	MemoryMB    types.Int64    `tfsdk:"memory_mb"`
	NetworkID   types.String   `tfsdk:"network_id"`
	Tags        []types.String `tfsdk:"tags"`

	PrivateIP types.String `tfsdk:"private_ip"`
	PublicIP  types.String `tfsdk:"public_ip"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

// vmResource is the resource implementation.
type vmResource struct {
	client  client.HTTPClient
	service *clientvm.Service
}

// Metadata returns the resource type name.
func (r *vmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

// Schema defines the schema for the resource.
func (r *vmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the virtual machine.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the VM.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the virtual machine.",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "Image identifier to boot the VM.",
				Required:    true,
			},
			"cpu": schema.Int64Attribute{
				Description: "Number of virtual CPUs.",
				Required:    true,
			},
			"memory_mb": schema.Int64Attribute{
				Description: "Memory size in megabytes.",
				Required:    true,
			},
			"network_id": schema.StringAttribute{
				Description: "Optional network identifier to attach the VM.",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Optional set of tags applied to the VM.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"private_ip": schema.StringAttribute{
				Description: "Private IP assigned to the VM.",
				Computed:    true,
			},
			"public_ip": schema.StringAttribute{
				Description: "Public IP assigned to the VM, if any.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current power/state of the VM.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp of the VM.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.service = clientvm.NewService(client)
}

// ImportState allows `terraform import` to work for this resource.
func (r *vmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
