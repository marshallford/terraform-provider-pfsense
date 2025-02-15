package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

const (
	diagDetailPrefix                        = "Underlying error details"
	defaultDomainOverrideTLSQueries         = false
	defaultStaticMappingARPTableStaticEntry = false
	defaultApply                            = true
	applyDescription                        = "Apply change, defaults to 'true'."
	applyMarkdownDescription                = "Apply change, defaults to `true`."
	descriptionDescription                  = "For administrative reference (not parsed)."
)

type attrDescription struct {
	Description         string
	MarkdownDescription string
}

func unknownProviderValue(value path.Path) (string, string) {
	return fmt.Sprintf("Unknown configuration value '%s'", value),
		fmt.Sprintf("The provider cannot be configured as there is an unknown configuration value for '%s'. ", value) +
			"Either target apply the source of the value first or set the value statically in the configuration."
}

func unexpectedConfigureType(value string, providerData any) (string, string) {
	return fmt.Sprintf("Unexpected %s Configure Type", value),
		fmt.Sprintf("Expected *providerOptions, got: %T. Please report this issue to the provider developers.", providerData)
}

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*pfsense.Client, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*pfsense.Client)

	if !ok {
		summary, detail := unexpectedConfigureType("Resource", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*pfsense.Client, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	opts, ok := req.ProviderData.(*pfsense.Client)

	if !ok {
		summary, detail := unexpectedConfigureType("Data Source", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return opts, ok
}

// func configureEphemeralResourceClient(req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) (*pfsense.Client, bool) {
// 	if req.ProviderData == nil {
// 		return nil, false
// 	}

// 	opts, ok := req.ProviderData.(*pfsense.Client)

// 	if !ok {
// 		summary, detail := unexpectedConfigureType("Ephemeral Resource", req.ProviderData)
// 		resp.Diagnostics.AddError(summary, detail)
// 	}

// 	return opts, ok
// }

func addError(diags *diag.Diagnostics, summary string, err error) bool {
	if err != nil {
		diags.AddError(summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func addPathError(diags *diag.Diagnostics, path path.Path, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddAttributeError(path, summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func addWarning(diags *diag.Diagnostics, summary string, err error) bool { //nolint:unparam
	if err != nil {
		diags.AddWarning(summary, fmt.Sprintf("%s: %s", diagDetailPrefix, err))

		return true
	}

	return false
}

func wrapElements(input []string, wrap string) []string {
	output := make([]string, 0, len(input))
	for _, element := range input {
		output = append(output, fmt.Sprintf("%s%s%s", wrap, element, wrap))
	}

	return output
}

func wrapElementsJoin(input []string, wrap string) string {
	return strings.Join(wrapElements(input, wrap), ", ")
}

// ref: https://github.com/coryflucas/terraform-provider-kubernetes/blob/master/internal/framework/provider/functions/decode.go
func convertJSONToTerraform(ctx context.Context, data any) (attr.Value, diag.Diagnostics) { //nolint:ireturn
	switch value := data.(type) {
	case string:
		return types.StringValue(value), nil
	case float64:
		return types.Float64Value(value), nil
	case bool:
		return types.BoolValue(value), nil
	case nil:
		return types.DynamicNull(), nil
	case []any:
		attrValues := make([]attr.Value, len(value))
		attrTypes := make([]attr.Type, len(value))
		for index, v := range value {
			attrValue, diags := convertJSONToTerraform(ctx, v)
			if diags.HasError() {
				return nil, diags
			}
			attrValues[index] = attrValue
			attrTypes[index] = attrValue.Type(ctx)
		}

		return types.TupleValue(attrTypes, attrValues)
	case map[string]any:
		attrValues := make(map[string]attr.Value, len(value))
		attrTypes := make(map[string]attr.Type, len(value))

		for key, v := range value {
			attrValue, diags := convertJSONToTerraform(ctx, v)
			if diags.HasError() {
				return nil, diags
			}
			attrValues[key] = attrValue
			attrTypes[key] = attrValue.Type(ctx)
		}

		return types.ObjectValue(attrTypes, attrValues)
	default:
		var diags diag.Diagnostics
		diags.AddError(
			"Failed to convert JSON to Terraform attribute value",
			fmt.Sprintf("unexpected type: %T for value %#v", value, value),
		)

		return nil, diags
	}
}
