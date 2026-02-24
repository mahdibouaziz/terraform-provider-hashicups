package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Update updates the resource and sets the updated Terraform state on success.
func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loadBalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := lbPayloadFromPlan(plan)

	if err := r.client.Put(ctx, fmt.Sprintf("/loadbalancers/%s", plan.ID.ValueString()), payload, nil); err != nil {
		resp.Diagnostics.AddError("Error updating load balancer", "Could not update load balancer: "+err.Error())
		return
	}

	var lb apiLoadBalancer
	if err := r.client.Get(ctx, "/loadbalancers/"+plan.ID.ValueString(), &lb); err != nil {
		resp.Diagnostics.AddError("Error reading load balancer", "Could not read load balancer "+plan.ID.ValueString()+": "+err.Error())
		return
	}

	plan = lbPlanFromAPI(plan, lb)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
