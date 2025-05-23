package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ resource.Resource              = (*FirewallFilterReloadResource)(nil)
	_ resource.ResourceWithConfigure = (*FirewallFilterReloadResource)(nil)
)

func NewFirewallFilterReloadResource() resource.Resource { //nolint:ireturn
	return &FirewallFilterReloadResource{}
}

type FirewallFilterReloadResource struct {
	client *pfsense.Client
}

type FirewallFilterReloadResourceModel struct {
	ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

func (r *FirewallFilterReloadResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_filter_reload", req.ProviderTypeName)
}

func (r *FirewallFilterReloadResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reload firewall filter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for firewall filter reload.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Description: "Last updated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FirewallFilterReloadResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *FirewallFilterReloadResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallFilterReloadResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.ReloadFirewallFilter(ctx)
	if addError(&resp.Diagnostics, "Error reloading firewall filter", err) {
		return
	}

	data.ID = types.StringValue(uuid.New().String())
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallFilterReloadResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *FirewallFilterReloadResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *FirewallFilterReloadResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
