package loadbalancer

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func lbPayloadFromPlan(plan loadBalancerResourceModel) apiLoadBalancerPayload {
	payload := apiLoadBalancerPayload{
		Name:     plan.Name.ValueString(),
		Protocol: plan.Protocol.ValueString(),
		Port:     int(plan.Port.ValueInt64()),
	}

	if !plan.HealthURL.IsNull() && !plan.HealthURL.IsUnknown() {
		payload.HealthURL = plan.HealthURL.ValueString()
	}

	for _, t := range plan.Targets {
		payload.Targets = append(payload.Targets, apiTarget{
			Address: t.Address.ValueString(),
			Port:    int(t.Port.ValueInt64()),
		})
	}

	return payload
}

func lbPlanFromAPI(plan loadBalancerResourceModel, lb apiLoadBalancer) loadBalancerResourceModel {
	plan.ID = types.StringValue(strconv.Itoa(lb.ID))
	plan.Name = types.StringValue(lb.Name)
	plan.Protocol = types.StringValue(lb.Protocol)
	plan.Port = types.Int64Value(int64(lb.Port))
	if lb.HealthURL != "" {
		plan.HealthURL = types.StringValue(lb.HealthURL)
	}

	plan.Targets = []loadBalancerTargetModel{}
	for _, t := range lb.Targets {
		plan.Targets = append(plan.Targets, loadBalancerTargetModel{
			Address: types.StringValue(t.Address),
			Port:    types.Int64Value(int64(t.Port)),
		})
	}

	plan.Status = types.StringValue(lb.Status)
	plan.CreatedAt = types.StringValue(lb.CreatedAt)
	return plan
}
