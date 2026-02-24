package vm

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func vmPayloadFromPlan(plan vmResourceModel) apiVMPayload {
	payload := apiVMPayload{
		Name:     plan.Name.ValueString(),
		Image:    plan.Image.ValueString(),
		CPU:      int(plan.CPU.ValueInt64()),
		MemoryMB: int(plan.MemoryMB.ValueInt64()),
	}

	if !plan.NetworkID.IsNull() && !plan.NetworkID.IsUnknown() {
		payload.NetworkID = plan.NetworkID.ValueString()
	}

	for _, tag := range plan.Tags {
		if !tag.IsNull() && !tag.IsUnknown() {
			payload.Tags = append(payload.Tags, tag.ValueString())
		}
	}

	return payload
}

func vmPlanFromAPI(plan vmResourceModel, vm apiVM) vmResourceModel {
	plan.ID = types.StringValue(strconv.Itoa(vm.ID))
	plan.Name = types.StringValue(vm.Name)
	plan.Image = types.StringValue(vm.Image)
	plan.CPU = types.Int64Value(int64(vm.CPU))
	plan.MemoryMB = types.Int64Value(int64(vm.MemoryMB))
	if vm.NetworkID != "" {
		plan.NetworkID = types.StringValue(vm.NetworkID)
	}

	plan.Tags = []types.String{}
	for _, tag := range vm.Tags {
		plan.Tags = append(plan.Tags, types.StringValue(tag))
	}

	plan.PrivateIP = types.StringValue(vm.PrivateIP)
	plan.PublicIP = types.StringValue(vm.PublicIP)
	plan.Status = types.StringValue(vm.Status)
	plan.CreatedAt = types.StringValue(vm.CreatedAt)

	return plan
}
