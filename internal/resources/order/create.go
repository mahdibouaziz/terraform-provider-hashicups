// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package order

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Create creates the resource and sets the initial Terraform state.
func (r *orderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API Request body from plan
	var items []hashicups.OrderItem
	for _, item := range plan.Items {
		newItem := hashicups.OrderItem{
			Quantity: int(item.Quantity.ValueInt64()),
			Coffee: hashicups.Coffee{
				ID: int(item.Coffee.ID.ValueInt64()),
			},
		}
		items = append(items, newItem)
	}

	// Create a new Order
	order, err := r.client.CreateOrder(items)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(order.ID))
	for index, item := range order.Items {
		plan.Items[index] = orderItemModel{
			Quantity: types.Int64Value(int64(item.Quantity)),
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(item.Coffee.ID)),
				Name:        types.StringValue(item.Coffee.Name),
				Teaser:      types.StringValue(item.Coffee.Teaser),
				Description: types.StringValue(item.Coffee.Description),
				Price:       types.Float64Value(item.Coffee.Price),
				Image:       types.StringValue(item.Coffee.Image),
			},
		}
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated daa
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}
