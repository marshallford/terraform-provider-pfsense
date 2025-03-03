package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type ExecutePHPCommandModel struct {
	Command types.String  `tfsdk:"command"`
	Result  types.Dynamic `tfsdk:"result"`
}

func (ExecutePHPCommandModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"command": {
			Description: "PHP command.",
		},
		"result": {
			Description: "Result of command.",
		},
	}
}
