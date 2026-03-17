package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ action.Action              = (*DHCPv4ApplyAction)(nil)
	_ action.ActionWithConfigure = (*DHCPv4ApplyAction)(nil)
)

func NewDHCPv4ApplyAction() action.Action { //nolint:ireturn
	return &DHCPv4ApplyAction{}
}

type DHCPv4ApplyAction struct {
	client *pfsense.Client
}

type DHCPv4ApplyActionModel struct {
	Interface types.String `tfsdk:"interface"`
}

func (a *DHCPv4ApplyAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dhcpv4_apply", req.ProviderTypeName)
}

func (a *DHCPv4ApplyAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Apply DHCPv4 configuration.",
		MarkdownDescription: "Apply DHCPv4 configuration." + privilegesMarkdown(pfsense.DHCPv4Changes{}),
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Description: "Network interface.",
				Required:    true,
				Validators: []validator.String{
					stringIsInterface(),
				},
			},
		},
	}
}

func (a *DHCPv4ApplyAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	client, ok := configureActionClient(req, resp)
	if !ok {
		return
	}

	a.client = client
}

func (a *DHCPv4ApplyAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data DHCPv4ApplyActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.ApplyDHCPv4Changes(ctx, data.Interface.ValueString())
	if addError(&resp.Diagnostics, "Error applying dhcpv4 changes", err) {
		return
	}
}
