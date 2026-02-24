package vm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Read refreshes the Terraform state with the latest data.
func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.service.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading VM", "Could not read VM "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state = vmPlanFromAPI(state, *vm)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
