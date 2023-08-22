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

var _ resource.Resource = &DNSResolverHostOverrideResource{}
var _ resource.ResourceWithImportState = &DNSResolverHostOverrideResource{}

func NewDNSResolverHostOverrideResource() resource.Resource {
	return &DNSResolverHostOverrideResource{}
}

type DNSResolverHostOverrideResource struct {
	client *pfsense.Client
}

type DNSResolverHostOverrideResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Host        types.String   `tfsdk:"host"`
	Domain      types.String   `tfsdk:"domain"`
	IPAddresses []types.String `tfsdk:"ip_addresses"`
	Description types.String   `tfsdk:"description"`
	Apply       types.Bool     `tfsdk:"apply"`
}

func (r *DNSResolverHostOverrideResourceModel) Map(hostOverride *pfsense.HostOverride) {
	if hostOverride.Host != "" {
		r.Host = types.StringValue(hostOverride.Host)
	}

	r.Domain = types.StringValue(hostOverride.Domain)

	ipAddresses := []types.String{}
	for _, ipAddress := range hostOverride.IPAddresses {
		ipAddresses = append(ipAddresses, types.StringValue(ipAddress.String()))
	}
	r.IPAddresses = ipAddresses

	if hostOverride.Description != "" {
		r.Description = types.StringValue(hostOverride.Description)
	}
}

func (r DNSResolverHostOverrideResourceModel) HostOverride(ctx *context.Context, diag *diag.Diagnostics) pfsense.HostOverride {
	var hostOverride pfsense.HostOverride

	if !r.ID.IsUnknown() {
		err := hostOverride.SetID(r.ID.ValueString())

		if err != nil {
			diag.AddAttributeError(
				path.Root("id"),
				"ID cannot be parsed",
				err.Error(),
			)
		}
	}

	if !r.Host.IsNull() {
		err := hostOverride.SetHost(r.Host.ValueString())

		if err != nil {
			diag.AddAttributeError(
				path.Root("host"),
				"Host cannot be parsed",
				err.Error(),
			)
		}
	}

	err := hostOverride.SetDomain(r.Domain.ValueString())

	if err != nil {
		diag.AddAttributeError(
			path.Root("domain"),
			"Domain cannot be parsed",
			err.Error(),
		)
	}

	var ipAddresses []string
	for _, ipAddress := range r.IPAddresses {
		ipAddresses = append(ipAddresses, ipAddress.ValueString())
	}
	err = hostOverride.SetIPAddress(ipAddresses)
	if err != nil {
		diag.AddAttributeError(
			path.Root("ip_addresses"),
			"IP addresses cannot be parsed",
			err.Error(),
		)
	}

	if !r.Description.IsNull() {
		err := hostOverride.SetDescription(r.Description.ValueString())

		if err != nil {
			diag.AddAttributeError(
				path.Root("description"),
				"Description cannot be parsed",
				err.Error(),
			)
		}
	}
	return hostOverride
}

func (r *DNSResolverHostOverrideResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsresolver_hostoverride"
}

// TODO validators
func (r *DNSResolverHostOverrideResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS Resolver Host Override. Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		MarkdownDescription: "DNS Resolver [Host Override](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-host-overrides.html). Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID for host override.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "Name of the host, without the domain part.",
				Optional:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Parent domain of the host.",
				Required:    true,
			},
			"ip_addresses": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "IPv4 or IPv6 addresses to be returned for the host.",
				Required:    true,
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

func (r *DNSResolverHostOverrideResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pfsense.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *pfsense.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *DNSResolverHostOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverHostOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverrideReq := data.HostOverride(&ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.CreateDNSResolverHostOverride(ctx, hostOverrideReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating host override",
			"Could not create host override, unexpected error: "+err.Error(),
		)
		return
	}

	if data.Apply.ValueBool() {
		_, err = r.client.ApplyDNSResolverChanges(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error applying host override",
				"Could not apply host override, unexpected error: "+err.Error(),
			)
			return
		}
	}

	data.ID = types.StringValue(hostOverride.ID.String())
	data.Map(hostOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverHostOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverHostOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.GetDNSResolverHostOverride(ctx, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading host override",
			"Could not read host override, unexpected error: "+err.Error(),
		)
		return
	}

	data.Map(hostOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverHostOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverHostOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverrideReq := data.HostOverride(&ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.UpdateDNSResolverHostOverride(ctx, hostOverrideReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating host override",
			"Could not update host override, unexpected error: "+err.Error(),
		)
		return
	}

	if data.Apply.ValueBool() {
		_, err = r.client.ApplyDNSResolverChanges(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error applying host override",
				"Could not apply host override, unexpected error: "+err.Error(),
			)
			return
		}
	}

	data.Map(hostOverride)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverHostOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverHostOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSResolverHostOverride(ctx, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting host override",
			"Could not delete host override, unexpected error: "+err.Error(),
		)
		return
	}

	if data.Apply.ValueBool() {
		_, err = r.client.ApplyDNSResolverChanges(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error applying host override",
				"Could not apply host override, unexpected error: "+err.Error(),
			)
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
