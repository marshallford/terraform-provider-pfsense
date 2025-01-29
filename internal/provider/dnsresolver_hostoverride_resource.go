package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ resource.Resource                = &DNSResolverHostOverrideResource{}
	_ resource.ResourceWithImportState = &DNSResolverHostOverrideResource{}
)

func NewDNSResolverHostOverrideResource() resource.Resource { //nolint:ireturn
	return &DNSResolverHostOverrideResource{}
}

type DNSResolverHostOverrideResource struct {
	client *pfsense.Client
}

func (r *DNSResolverHostOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_hostoverride", req.ProviderTypeName)
}

func (r *DNSResolverHostOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS resolver host override. Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		MarkdownDescription: "DNS resolver [host override](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-host-overrides.html). Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: DNSResolverHostOverrideModel{}.descriptions()["host"].Description,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Description: DNSResolverHostOverrideModel{}.descriptions()["domain"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_addresses": schema.ListAttribute{
				Description: DNSResolverHostOverrideModel{}.descriptions()["ip_addresses"].Description,
				Required:    true,
				ElementType: types.StringType,
			},
			"description": schema.StringAttribute{
				Description: DNSResolverHostOverrideModel{}.descriptions()["description"].Description,
				Optional:    true,
			},
			"apply": schema.BoolAttribute{
				Description:         DNSResolverHostOverrideModel{}.descriptions()["apply"].Description,
				MarkdownDescription: DNSResolverHostOverrideModel{}.descriptions()["apply"].MarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"fqdn": schema.StringAttribute{
				Description: DNSResolverHostOverrideModel{}.descriptions()["fqdn"].Description,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aliases": schema.ListNestedAttribute{
				Description:         DNSResolverHostOverrideModel{}.descriptions()["aliases"].Description,
				MarkdownDescription: DNSResolverHostOverrideModel{}.descriptions()["aliases"].MarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: DNSResolverHostOverrideAliasModel{}.AttrTypes()}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Description: DNSResolverHostOverrideAliasModel{}.descriptions()["host"].Description,
							Optional:    true,
						},
						"domain": schema.StringAttribute{
							Description: DNSResolverHostOverrideAliasModel{}.descriptions()["domain"].Description,
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: DNSResolverHostOverrideAliasModel{}.descriptions()["description"].Description,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *DNSResolverHostOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverHostOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverHostOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var hostOverrideReq pfsense.HostOverride
	resp.Diagnostics.Append(data.Value(ctx, &hostOverrideReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.CreateDNSResolverHostOverride(ctx, hostOverrideReq)
	if addError(&resp.Diagnostics, "Error creating host override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *hostOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addWarning(&resp.Diagnostics, "Error applying host override", err) {
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverHostOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.GetDNSResolverHostOverride(ctx, data.FQDN.ValueString())
	if addError(&resp.Diagnostics, "Error reading host override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *hostOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverHostOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverHostOverrideModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var hostOverrideReq pfsense.HostOverride
	resp.Diagnostics.Append(data.Value(ctx, &hostOverrideReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.UpdateDNSResolverHostOverride(ctx, hostOverrideReq)
	if addError(&resp.Diagnostics, "Error updating host override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *hostOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addWarning(&resp.Diagnostics, "Error applying host override", err) {
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverHostOverrideModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSResolverHostOverride(ctx, data.FQDN.ValueString())
	if addError(&resp.Diagnostics, "Error deleting host override", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addWarning(&resp.Diagnostics, "Error applying host override", err) {
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: host,domain. Got: %q", req.ID),
		)

		return
	}

	var hostOverride pfsense.HostOverride

	if idParts[0] != "" {
		if addError(&resp.Diagnostics, "Host cannot be parsed", hostOverride.SetHost(idParts[0])) {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host"), hostOverride.Host)...)
	}

	if addError(&resp.Diagnostics, "Domain cannot be parsed", hostOverride.SetDomain(idParts[1])) {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), hostOverride.Domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fqdn"), hostOverride.FQDN())...)
}
