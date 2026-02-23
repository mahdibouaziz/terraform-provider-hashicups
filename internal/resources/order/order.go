// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package order

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
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &orderResource{}
	_ resource.ResourceWithConfigure   = &orderResource{}
	_ resource.ResourceWithImportState = &orderResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewOrderResource() resource.Resource {
	return &orderResource{}
}

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	ID          types.String     `tfsdk:"id"`
	LastUpdated types.String     `tfsdk:"last_updated"`
	Items       []orderItemModel `tfsdk:"items"`
}

// orderItemModel maps order item data.
type orderItemModel struct {
	Quantity types.Int64          `tfsdk:"quantity"`
	Coffee   orderItemCoffeeModel `tfsdk:"coffee"`
}

// orderItemCoffeeModel maps coffee order item data.
type orderItemCoffeeModel struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Teaser      types.String  `tfsdk:"teaser"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	Image       types.String  `tfsdk:"image"`
}

// orderResource is the resource implementation.
type orderResource struct {
	client *client.Client
}

// Metadata returns the resource type name.
func (r *orderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order"
}

// Schema defines the schema for the resource.
func (r *orderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the order.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the order.",
				Computed:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of items in the order.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"quantity": schema.Int64Attribute{
							Description: "Count of this item in the order.",
							Required:    true,
						},
						"coffee": schema.SingleNestedAttribute{
							Description: "Coffee item in the order.",
							Required:    true,
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									Description: "Numeric identifier of the coffee.",
									Required:    true,
								},
								"name": schema.StringAttribute{
									Description: "Product name of the coffee.",
									Computed:    true,
								},
								"teaser": schema.StringAttribute{
									Description: "Fun tagline for the coffee.",
									Computed:    true,
								},
								"description": schema.StringAttribute{
									Description: "Product description of the coffee.",
									Computed:    true,
								},
								"price": schema.Float64Attribute{
									Description: "Suggested cost of the coffee.",
									Computed:    true,
								},
								"image": schema.StringAttribute{
									Description: "URI for an image of the coffee.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *orderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Retrieve import ID and save to id attribute
func (r *orderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
