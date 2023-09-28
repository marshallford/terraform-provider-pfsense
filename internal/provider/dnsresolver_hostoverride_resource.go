package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

var _ resource.Resource = &DNSResolverHostOverrideResource{}
var _ resource.ResourceWithImportState = &DNSResolverHostOverrideResource{}

func NewDNSResolverHostOverrideResource() resource.Resource {
	return &DNSResolverHostOverrideResource{}
}

type DNSResolverHostOverrideResource struct {
	client *pfsense.Client
}

type DNSResolverHostOverrideResourceModel struct {
	Host        types.String   `tfsdk:"host"`
	Domain      types.String   `tfsdk:"domain"`
	IPAddresses []types.String `tfsdk:"ip_addresses"`
	Description types.String   `tfsdk:"description"`
	Apply       types.Bool     `tfsdk:"apply"`
	FQDN        types.String   `tfsdk:"fqdn"`
	Aliases     types.List     `tfsdk:"aliases"`
}

type DNSResolverHostOverrideAliasResourceModel struct {
	Host        types.String `tfsdk:"host"`
	Domain      types.String `tfsdk:"domain"`
	Description types.String `tfsdk:"description"`
}

func (r DNSResolverHostOverrideAliasResourceModel) GetType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"host":        types.StringType,
		"domain":      types.StringType,
		"description": types.StringType,
	}}
}

func (r *DNSResolverHostOverrideResourceModel) SetFromClient(ctx context.Context, hostOverride *pfsense.HostOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if hostOverride.Host != "" {
		r.Host = types.StringValue(hostOverride.Host)
	}

	r.Domain = types.StringValue(hostOverride.Domain)

	var ipAddresses []types.String
	for _, ipAddress := range hostOverride.IPAddresses {
		ipAddresses = append(ipAddresses, types.StringValue(ipAddress.String()))
	}
	r.IPAddresses = ipAddresses

	if hostOverride.Description != "" {
		r.Description = types.StringValue(hostOverride.Description)
	}

	r.FQDN = types.StringValue(hostOverride.FQDN())

	aliases := []DNSResolverHostOverrideAliasResourceModel{}

	for _, alias := range hostOverride.Aliases {
		var aliasModel DNSResolverHostOverrideAliasResourceModel

		if alias.Host != "" {
			aliasModel.Host = types.StringValue(alias.Host)
		}

		aliasModel.Domain = types.StringValue(alias.Domain)

		if alias.Description != "" {
			aliasModel.Description = types.StringValue(alias.Description)
		}

		aliases = append(aliases, aliasModel)
	}

	r.Aliases, diags = types.ListValueFrom(ctx, DNSResolverHostOverrideAliasResourceModel{}.GetType(), aliases)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r DNSResolverHostOverrideResourceModel) GetClientValue(ctx context.Context) (*pfsense.HostOverride, diag.Diagnostics) {
	var hostOverride pfsense.HostOverride
	var err error
	var diags diag.Diagnostics

	var aliasModels []*DNSResolverHostOverrideAliasResourceModel
	diags = r.Aliases.ElementsAs(ctx, &aliasModels, false)
	if diags.HasError() {
		return nil, diags
	}

	if !r.Host.IsNull() {
		err = hostOverride.SetHost(r.Host.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("host"),
				"Host cannot be parsed",
				err.Error(),
			)
		}
	}

	err = hostOverride.SetDomain(r.Domain.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("domain"),
			"Domain cannot be parsed",
			err.Error(),
		)
	}

	var ipAddresses []string
	for _, ipAddress := range r.IPAddresses {
		ipAddresses = append(ipAddresses, ipAddress.ValueString())
	}

	err = hostOverride.SetIPAddresses(ipAddresses)
	if err != nil {
		diags.AddAttributeError(
			path.Root("ip_addresses"),
			"IP addresses cannot be parsed",
			err.Error(),
		)
	}

	if !r.Description.IsNull() {
		err = hostOverride.SetDescription(r.Description.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("description"),
				"Description cannot be parsed",
				err.Error(),
			)
		}
	}

	for i, aliasModel := range aliasModels {
		var alias pfsense.HostOverrideAlias

		if !aliasModel.Host.IsNull() {
			err = alias.SetHost(aliasModel.Host.ValueString())
			if err != nil {
				diags.AddAttributeError(
					path.Root("aliases").AtListIndex(i).AtName("host"),
					"Alias host cannot be parsed",
					err.Error(),
				)
			}
		}

		err = alias.SetDomain(aliasModel.Domain.ValueString())
		if err != nil {
			diags.AddAttributeError(
				path.Root("aliases").AtListIndex(i).AtName("domain"),
				"Alias domain cannot be parsed",
				err.Error(),
			)
		}

		if !aliasModel.Description.IsNull() {
			err = alias.SetDescription(aliasModel.Description.ValueString())
			if err != nil {
				diags.AddAttributeError(
					path.Root("aliases").AtListIndex(i).AtName("description"),
					"Alias description cannot be parsed",
					err.Error(),
				)
			}
		}

		hostOverride.Aliases = append(hostOverride.Aliases, alias)
	}

	return &hostOverride, diags
}

func (r *DNSResolverHostOverrideResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_hostoverride", req.ProviderTypeName)
}

func (r *DNSResolverHostOverrideResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS Resolver Host Override. Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		MarkdownDescription: "DNS Resolver [Host Override](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-host-overrides.html). Host for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Name of the host, without the domain part.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Parent domain of the host.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"fqdn": schema.StringAttribute{
				Description: "Fully qualified domain name of host.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aliases": schema.ListNestedAttribute{
				Description:         "List of additional names for this host, defaults to '[]'.",
				MarkdownDescription: "List of additional names for this host, defaults to `[]`.",
				Computed:            true,
				Optional:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(DNSResolverHostOverrideAliasResourceModel{}.GetType(), []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Description: "Name of the host, without the domain part.",
							Optional:    true,
						},
						"domain": schema.StringAttribute{
							Description: "Parent domain of the host.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "For administrative reference (not parsed).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *DNSResolverHostOverrideResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)

	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverHostOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverHostOverrideResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverrideReq, d := data.GetClientValue(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.CreateDNSResolverHostOverride(ctx, *hostOverrideReq)
	if addError(&resp.Diagnostics, "Error creating host override", err) {
		return
	}

	diags = data.SetFromClient(ctx, hostOverride)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addError(&resp.Diagnostics, "Error applying host override", err) {
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverHostOverrideResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.GetDNSResolverHostOverride(ctx, data.FQDN.ValueString())
	if addError(&resp.Diagnostics, "Error reading host override", err) {
		return
	}

	diags = data.SetFromClient(ctx, hostOverride)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverHostOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverHostOverrideResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverrideReq, d := data.GetClientValue(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	hostOverride, err := r.client.UpdateDNSResolverHostOverride(ctx, *hostOverrideReq)
	if addError(&resp.Diagnostics, "Error updating host override", err) {
		return
	}

	diags = data.SetFromClient(ctx, hostOverride)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addError(&resp.Diagnostics, "Error applying host override", err) {
			return
		}
	}
}

func (r *DNSResolverHostOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverHostOverrideResourceModel
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
		if addError(&resp.Diagnostics, "Error applying host override", err) {
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

	var ho pfsense.HostOverride
	var err error

	if idParts[0] != "" {
		err = ho.SetHost(idParts[0])
		if err != nil {
			resp.Diagnostics.AddError(
				"Host cannot be parsed",
				err.Error(),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host"), ho.Host)...)
	}

	err = ho.SetDomain(idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Domain cannot be parsed",
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), ho.Domain)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("fqdn"), ho.FQDN())...)
}
