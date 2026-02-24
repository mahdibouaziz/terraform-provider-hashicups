package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Delete deletes the resource and removes the Terraform state on success.
func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.service.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting load balancer", "Could not delete load balancer: "+err.Error())
		return
	}
}
