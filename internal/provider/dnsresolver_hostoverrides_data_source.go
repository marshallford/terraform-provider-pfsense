package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = &DNSResolverHostOverridesDataSource{}
	_ datasource.DataSourceWithConfigure = &DNSResolverHostOverridesDataSource{}
)

func NewDNSResolverHostOverridesDataSource() datasource.DataSource {
	return &DNSResolverHostOverridesDataSource{}
}

type DNSResolverHostOverridesDataSource struct {
	client *pfsense.Client
}

type DNSResolverHostOverridesDataSourceModel struct {
	All types.List `tfsdk:"all"`
}

type DNSResolverHostOverrideDataSourceModel struct {
	Host        types.String   `tfsdk:"host"`
	Domain      types.String   `tfsdk:"domain"`
	IPAddresses []types.String `tfsdk:"ip_addresses"`
	Description types.String   `tfsdk:"description"`
	FQDN        types.String   `tfsdk:"fqdn"`
	Aliases     types.List     `tfsdk:"aliases"`
}

func (d DNSResolverHostOverrideDataSourceModel) GetAttrType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"host":         types.StringType,
		"domain":       types.StringType,
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"description":  types.StringType,
		"fqdn":         types.StringType,
		"aliases":      types.ListType{ElemType: DNSResolverHostOverrideAliasDataSourceModel{}.GetAttrType()},
	}}
}

type DNSResolverHostOverrideAliasDataSourceModel struct {
	Host        types.String `tfsdk:"host"`
	Domain      types.String `tfsdk:"domain"`
	Description types.String `tfsdk:"description"`
}

func (d DNSResolverHostOverrideAliasDataSourceModel) GetAttrType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"host":        types.StringType,
		"domain":      types.StringType,
		"description": types.StringType,
	}}
}

func (d *DNSResolverHostOverrideDataSourceModel) SetFromValue(ctx context.Context, hostOverride *pfsense.HostOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if hostOverride.Host != "" {
		d.Host = types.StringValue(hostOverride.Host)
	}

	d.Domain = types.StringValue(hostOverride.Domain)

	var ipAddresses []types.String
	for _, ipAddress := range hostOverride.IPAddresses {
		ipAddresses = append(ipAddresses, types.StringValue(ipAddress.String()))
	}
	d.IPAddresses = ipAddresses

	if hostOverride.Description != "" {
		d.Description = types.StringValue(hostOverride.Description)
	}

	d.FQDN = types.StringValue(hostOverride.FQDN())

	aliases := []DNSResolverHostOverrideAliasDataSourceModel{}

	for _, alias := range hostOverride.Aliases {
		var aliasModel DNSResolverHostOverrideAliasDataSourceModel

		if alias.Host != "" {
			aliasModel.Host = types.StringValue(alias.Host)
		}

		aliasModel.Domain = types.StringValue(alias.Domain)

		if alias.Description != "" {
			aliasModel.Description = types.StringValue(alias.Description)
		}

		aliases = append(aliases, aliasModel)
	}

	d.Aliases, diags = types.ListValueFrom(ctx, DNSResolverHostOverrideAliasDataSourceModel{}.GetAttrType(), aliases)

	return diags
}

func (d *DNSResolverHostOverridesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_hostoverrides", req.ProviderTypeName)
}

func (d *DNSResolverHostOverridesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves all DNS resolver host overrides. Hosts for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		MarkdownDescription: "Retrieves all DNS resolver [host overrides](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-host-overrides.html). Hosts for which the resolver's standard DNS lookup process should be overridden and a specific IPv4 or IPv6 address should automatically be returned by the resolver.",
		Attributes: map[string]schema.Attribute{
			"all": schema.ListNestedAttribute{
				Description: "All host overrides.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Description: "Name of the host, without the domain part.",
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: "Parent domain of the host.",
							Computed:    true,
						},
						"ip_addresses": schema.ListAttribute{
							ElementType: types.StringType,
							Description: "IPv4 or IPv6 addresses to be returned for the host.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "For administrative reference (not parsed).",
							Computed:    true,
						},
						"fqdn": schema.StringAttribute{
							Description: "Fully qualified domain name of host.",
							Computed:    true,
						},
						"aliases": schema.ListNestedAttribute{
							Description:         "List of additional names for this host, defaults to '[]'.",
							MarkdownDescription: "List of additional names for this host, defaults to `[]`.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.StringAttribute{
										Description: "Name of the host, without the domain part.",
										Computed:    true,
									},
									"domain": schema.StringAttribute{
										Description: "Parent domain of the host.",
										Computed:    true,
									},
									"description": schema.StringAttribute{
										Description: "For administrative reference (not parsed).",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *DNSResolverHostOverridesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *DNSResolverHostOverridesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSResolverHostOverridesDataSourceModel
	var diags diag.Diagnostics

	hostOverrides, err := d.client.GetDNSResolverHostOverrides(ctx)
	if addError(&resp.Diagnostics, "Unable to get host overrides", err) {
		return
	}

	hostOverrideModels := []DNSResolverHostOverrideDataSourceModel{}
	for _, hostOverride := range *hostOverrides {
		var hostOverrideModel DNSResolverHostOverrideDataSourceModel
		hostOverride := hostOverride
		diags = hostOverrideModel.SetFromValue(ctx, &hostOverride)
		resp.Diagnostics.Append(diags...)
		hostOverrideModels = append(hostOverrideModels, hostOverrideModel)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.All, diags = types.ListValueFrom(ctx, DNSResolverHostOverrideDataSourceModel{}.GetAttrType(), hostOverrideModels)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
