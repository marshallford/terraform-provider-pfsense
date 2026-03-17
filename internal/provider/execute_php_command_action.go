package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ action.Action              = (*ExecutePHPCommandAction)(nil)
	_ action.ActionWithConfigure = (*ExecutePHPCommandAction)(nil)
)

func NewExecutePHPCommandAction() action.Action { //nolint:ireturn
	return &ExecutePHPCommandAction{}
}

type ExecutePHPCommandAction struct {
	client *pfsense.Client
}

type ExecutePHPCommandActionModel struct {
	Command types.String `tfsdk:"command"`
}

func (a *ExecutePHPCommandAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_execute_php_command", req.ProviderTypeName)
}

func (a *ExecutePHPCommandAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Execute PHP command. The command must print exactly one valid JSON value.",
		MarkdownDescription: "[Execute PHP command](https://docs.netgate.com/pfsense/en/latest/diagnostics/command-prompt.html#php-execute). The command must print exactly one valid JSON value." + privilegesMarkdown(pfsense.ExecutePHPCommand{}),
		Attributes: map[string]schema.Attribute{
			"command": schema.StringAttribute{
				Description: ExecutePHPCommandModel{}.descriptions()["command"].Description,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (a *ExecutePHPCommandAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	client, ok := configureActionClient(req, resp)
	if !ok {
		return
	}

	a.client = client
}

func (a *ExecutePHPCommandAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data ExecutePHPCommandActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := a.client.ExecutePHPCommand(ctx, data.Command.ValueString(), true)
	if addError(&resp.Diagnostics, "Failed to execute PHP command", err) {
		return
	}

	resultJson, err := json.Marshal(result)
	if addError(&resp.Diagnostics, "Failed to marshal result", err) {
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "\n" + string(resultJson),
	})
}
