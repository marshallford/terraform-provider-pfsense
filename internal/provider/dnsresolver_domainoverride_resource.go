package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var _ resource.Resource = &DNSResolverDomainOverrideResource{}
var _ resource.ResourceWithImportState = &DNSResolverDomainOverrideResource{}

func NewDNSResolverDomainOverrideResource() resource.Resource {
	return &DNSResolverDomainOverrideResource{}
}

type DNSResolverDomainOverrideResource struct {
	client *pfsense.Client
}

type DNSResolverDomainOverrideResourceModel struct {
	Domain      types.String `tfsdk:"domain"`
	IPAddress   types.String `tfsdk:"ip_address"`
	TLSHostname types.String `tfsdk:"tls_hostname"`
	Description types.String `tfsdk:"description"`
	TLSQueries  types.Bool   `tfsdk:"tls_queries"`
	Apply       types.Bool   `tfsdk:"apply"`
}

func (r *DNSResolverDomainOverrideResourceModel) Map(domainOverride *pfsense.DomainOverride) {
	r.Domain = types.StringValue(domainOverride.Domain)
	r.IPAddress = types.StringValue(domainOverride.IPAddress.String())
	r.TLSQueries = types.BoolValue(domainOverride.TLSQueries)

	if domainOverride.TLSHostname != "" {
		r.TLSHostname = types.StringValue(domainOverride.TLSHostname)
	}

	if domainOverride.Description != "" {
		r.Description = types.StringValue(domainOverride.Description)
	}
}

func (r DNSResolverDomainOverrideResourceModel) DomainOverride(diag *diag.Diagnostics) pfsense.DomainOverride {
	var domainOverride pfsense.DomainOverride
	var err error

	err = domainOverride.SetDomain(r.Domain.ValueString())

	if err != nil {
		diag.AddAttributeError(
			path.Root("domain"),
			"Domain cannot be parsed",
			err.Error(),
		)
	}

	err = domainOverride.SetIPAddress(r.IPAddress.ValueString())

	if err != nil {
		diag.AddAttributeError(
			path.Root("ip_address"),
			"IP address cannot be parsed",
			err.Error(),
		)
	}

	err = domainOverride.SetTLSQueries(r.TLSQueries.ValueBool())

	if err != nil {
		diag.AddAttributeError(
			path.Root("tls_queries"),
			"TLS Queries cannot be parsed",
			err.Error(),
		)
	}

	if !r.TLSHostname.IsNull() {
		err = domainOverride.SetTLSHostname(r.TLSHostname.ValueString())

		if err != nil {
			diag.AddAttributeError(
				path.Root("tls_hostname"),
				"TLS Hostname cannot be parsed",
				err.Error(),
			)
		}
	}

	if !r.Description.IsNull() {
		err = domainOverride.SetDescription(r.Description.ValueString())

		if err != nil {
			diag.AddAttributeError(
				path.Root("description"),
				"Description cannot be parsed",
				err.Error(),
			)
		}
	}

	return domainOverride
}

func (r *DNSResolverDomainOverrideResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_domainoverride", req.ProviderTypeName)
}

// TODO validators
func (r *DNSResolverDomainOverrideResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS Resolver Domain Override. Domain for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		MarkdownDescription: "DNS Resolver [Domain Override](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-domain-overrides.html). Domain for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain whose lookups will be directed to a user-specified DNS lookup server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "IPv4 or IPv6 address (including port) of the authoritative DNS server for this domain.",
				Required:    true,
			},
			"tls_queries": schema.BoolAttribute{
				Description:         "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to 'false'.",
				MarkdownDescription: "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to `false`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tls_hostname": schema.StringAttribute{
				Description: "An optional TLS hostname used to verify the server certificate when performing TLS Queries.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "For administrative reference (not parsed).",
				Optional:    true,
			},
			"apply": schema.BoolAttribute{
				Description:         "Apply change, defaults to 'true'.",
				MarkdownDescription: "Apply change, defaults to `true`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *DNSResolverDomainOverrideResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)

	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverDomainOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverrideReq := data.DomainOverride(&resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.CreateDNSResolverDomainOverride(ctx, domainOverrideReq)

	if addError(&resp.Diagnostics, "Error creating domain override", err) {
		return
	}

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)

		if addError(&resp.Diagnostics, "Error applying domain override", err) {
			return
		}
	}

	data.Map(domainOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverDomainOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.GetDNSResolverDomainOverride(ctx, data.Domain.ValueString())

	if addError(&resp.Diagnostics, "Error reading domain override", err) {
		return
	}

	data.Map(domainOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverDomainOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverrideReq := data.DomainOverride(&resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.UpdateDNSResolverDomainOverride(ctx, domainOverrideReq)

	if addError(&resp.Diagnostics, "Error updating domain override", err) {
		return
	}

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)

		if addError(&resp.Diagnostics, "Error applying domain override", err) {
			return
		}
	}

	data.Map(domainOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverDomainOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSResolverDomainOverride(ctx, data.Domain.ValueString())

	if addError(&resp.Diagnostics, "Error deleting domain override", err) {
		return
	}

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)

		if addError(&resp.Diagnostics, "Error applying domain override", err) {
			return
		}
	}
}

func (r *DNSResolverDomainOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}
