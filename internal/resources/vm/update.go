package vm

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Update updates the resource and sets the updated Terraform state on success.
func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := vmPayloadFromPlan(plan)

	vm, err := r.service.Update(ctx, plan.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating VM", "Could not update/read VM: "+err.Error())
		return
	}

	plan = vmPlanFromAPI(plan, *vm)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
