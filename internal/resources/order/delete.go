// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package order

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Delete deletes the resource and removes the Terraform state on success.
func (r *orderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.DeleteOrder(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HashiCups Order",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}
