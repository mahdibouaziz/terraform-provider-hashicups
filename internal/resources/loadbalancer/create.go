package loadbalancer

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Create creates the resource and sets the initial Terraform state.
func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := lbPayloadFromPlan(plan)

	var lb apiLoadBalancer
	if err := r.client.Post(ctx, "/loadbalancers", payload, &lb); err != nil {
		resp.Diagnostics.AddError("Error creating load balancer", "Could not create load balancer: "+err.Error())
		return
	}

	plan = lbPlanFromAPI(plan, lb)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	plan.ID = types.StringValue(strconv.Itoa(lb.ID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
