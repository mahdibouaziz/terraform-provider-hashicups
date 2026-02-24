package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Read refreshes the Terraform state with the latest data.
func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var lb apiLoadBalancer
	if err := r.client.Get(ctx, "/loadbalancers/"+state.ID.ValueString(), &lb); err != nil {
		resp.Diagnostics.AddError("Error reading load balancer", "Could not read load balancer "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state = lbPlanFromAPI(state, lb)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
