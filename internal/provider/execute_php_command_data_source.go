package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = (*ExecutePHPCommandDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ExecutePHPCommandDataSource)(nil)
)

func NewExecutePHPCommandDataSource() datasource.DataSource { //nolint:ireturn
	return &ExecutePHPCommandDataSource{}
}

type ExecutePHPCommandDataSource struct {
	client *pfsense.Client
}

func (d *ExecutePHPCommandDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_execute_php_command", req.ProviderTypeName)
}

func (d *ExecutePHPCommandDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Execute PHP command. The command must print exactly one valid JSON value. Only execute commands without observable side-effects.",
		MarkdownDescription: "[Execute PHP command](https://docs.netgate.com/pfsense/en/latest/diagnostics/command-prompt.html#php-execute). The command must print exactly one valid JSON value. Only execute commands without observable side-effects.",
		Attributes: map[string]schema.Attribute{
			"command": schema.StringAttribute{
				Description: ExecutePHPCommandModel{}.descriptions()["command"].Description,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"result": schema.DynamicAttribute{
				Description: ExecutePHPCommandModel{}.descriptions()["result"].Description,
				Computed:    true,
			},
		},
	}
}

func (d *ExecutePHPCommandDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *ExecutePHPCommandDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExecutePHPCommandModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.ExecutePHPCommand(ctx, data.Command.ValueString(), "read")
	if addError(&resp.Diagnostics, "Failed to execute PHP command", err) {
		return
	}

	resultValue, newDiags := convertJSONToTerraform(ctx, result)
	resp.Diagnostics.Append(newDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Result = types.DynamicValue(resultValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
