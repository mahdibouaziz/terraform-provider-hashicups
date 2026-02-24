package vm

import (
	"context"
	"fmt"
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

	if err := r.client.Put(ctx, fmt.Sprintf("/vms/%s", plan.ID.ValueString()), payload, nil); err != nil {
		resp.Diagnostics.AddError("Error updating VM", "Could not update VM: "+err.Error())
		return
	}

	var vm apiVM
	if err := r.client.Get(ctx, "/vms/"+plan.ID.ValueString(), &vm); err != nil {
		resp.Diagnostics.AddError("Error reading VM", "Could not read VM "+plan.ID.ValueString()+": "+err.Error())
		return
	}

	plan = vmPlanFromAPI(plan, vm)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}
