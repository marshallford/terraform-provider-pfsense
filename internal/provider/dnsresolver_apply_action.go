package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ action.Action              = (*DNSResolverApplyAction)(nil)
	_ action.ActionWithConfigure = (*DNSResolverApplyAction)(nil)
)

func NewDNSResolverApplyAction() action.Action { //nolint:ireturn
	return &DNSResolverApplyAction{}
}

type DNSResolverApplyAction struct {
	client *pfsense.Client
}

func (a *DNSResolverApplyAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_apply", req.ProviderTypeName)
}

func (a *DNSResolverApplyAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Apply DNS resolver configuration.",
		MarkdownDescription: "Apply DNS resolver configuration." + privilegesMarkdown(pfsense.DNSResolverChanges{}),
	}
}

func (a *DNSResolverApplyAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	client, ok := configureActionClient(req, resp)
	if !ok {
		return
	}

	a.client = client
}

func (a *DNSResolverApplyAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	err := a.client.ApplyDNSResolverChanges(ctx)
	if addError(&resp.Diagnostics, "Error applying dns resolver changes", err) {
		return
	}
}
