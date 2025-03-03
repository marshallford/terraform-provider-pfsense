package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/dynamicplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var _ resource.Resource = &ExecutePHPCommandResource{}

type ExecutePHPCommandResourceModel struct {
	ExecutePHPCommandModel
	DestroyCommand types.String `tfsdk:"destroy_command"`
}

func NewExecutePHPCommandResource() resource.Resource { //nolint:ireturn
	return &ExecutePHPCommandResource{}
}

type ExecutePHPCommandResource struct {
	client *pfsense.Client
}

func (r *ExecutePHPCommandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_execute_php_command", req.ProviderTypeName)
}

func (r *ExecutePHPCommandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Execute PHP command. The command must print exactly one valid JSON value.",
		MarkdownDescription: "[Execute PHP command](https://docs.netgate.com/pfsense/en/latest/diagnostics/command-prompt.html#php-execute). The command must print exactly one valid JSON value.",
		Attributes: map[string]schema.Attribute{
			"command": schema.StringAttribute{
				Description: ExecutePHPCommandModel{}.descriptions()["command"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"destroy_command": schema.StringAttribute{
				Description: "PHP command to run on destroy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"result": schema.DynamicAttribute{
				Description: ExecutePHPCommandModel{}.descriptions()["result"].Description,
				Computed:    true,
				PlanModifiers: []planmodifier.Dynamic{
					dynamicplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ExecutePHPCommandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *ExecutePHPCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExecutePHPCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ExecutePHPCommand(ctx, data.Command.ValueString(), "create")
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

func (r *ExecutePHPCommandResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *ExecutePHPCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExecutePHPCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExecutePHPCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExecutePHPCommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.DestroyCommand.IsNull() {
		_, err := r.client.ExecutePHPCommand(ctx, data.DestroyCommand.ValueString(), "delete")
		addError(&resp.Diagnostics, "Failed to execute PHP command", err)
	}
}
