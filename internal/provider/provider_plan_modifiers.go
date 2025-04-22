package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type macAddressModifier struct{}

var (
	_ planmodifier.String = (*macAddressModifier)(nil)
)

func (m macAddressModifier) Description(_ context.Context) string {
	return "Compares MAC addresses semantically regardless of format"
}

func (m macAddressModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m macAddressModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	stateValue := req.StateValue.ValueString()
	planValue := req.PlanValue.ValueString()

	if stateValue == planValue {
		return
	}

	stateHwAddr, err := pfsense.ParseMACAddress(stateValue)
	if addPathWarning(&resp.Diagnostics, req.Path, "Invalid MAC address in state", err) {
		return
	}

	planHwAddr, err := pfsense.ParseMACAddress(planValue)
	if addPathError(&resp.Diagnostics, req.Path, "Invalid MAC address in plan", err) {
		return
	}

	if pfsense.CompareMACAddresses(stateHwAddr, planHwAddr) {
		resp.PlanValue = req.StateValue
	}
}

func macAddressPlanModifier() macAddressModifier {
	return macAddressModifier{}
}
