package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ action.Action              = (*FirewallFilterReloadAction)(nil)
	_ action.ActionWithConfigure = (*FirewallFilterReloadAction)(nil)
)

func NewFirewallFilterReloadAction() action.Action { //nolint:ireturn
	return &FirewallFilterReloadAction{}
}

type FirewallFilterReloadAction struct {
	client *pfsense.Client
}

func (a *FirewallFilterReloadAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_filter_reload", req.ProviderTypeName)
}

func (a *FirewallFilterReloadAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reload firewall filter.",
		MarkdownDescription: "Reload firewall filter." + privilegesMarkdown(pfsense.FirewallFilter{}),
	}
}

func (a *FirewallFilterReloadAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	client, ok := configureActionClient(req, resp)
	if !ok {
		return
	}

	a.client = client
}

func (a *FirewallFilterReloadAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	err := a.client.ReloadFirewallFilter(ctx)
	if addError(&resp.Diagnostics, "Error reloading firewall filter", err) {
		return
	}
}
