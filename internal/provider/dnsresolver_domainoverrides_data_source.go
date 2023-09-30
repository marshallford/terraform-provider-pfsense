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
	_ datasource.DataSource              = &DNSResolverDomainOverridesDataSource{}
	_ datasource.DataSourceWithConfigure = &DNSResolverDomainOverridesDataSource{}
)

func NewDNSResolverDomainOverridesDataSource() datasource.DataSource {
	return &DNSResolverDomainOverridesDataSource{}
}

type DNSResolverDomainOverridesDataSource struct {
	client *pfsense.Client
}

type DNSResolverDomainOverridesDataSourceModel struct {
	All types.List `tfsdk:"all"`
}

type DNSResolverDomainOverrideDataSourceModel struct {
	Domain      types.String `tfsdk:"domain"`
	IPAddress   types.String `tfsdk:"ip_address"`
	TLSQueries  types.Bool   `tfsdk:"tls_queries"`
	TLSHostname types.String `tfsdk:"tls_hostname"`
	Description types.String `tfsdk:"description"`
}

func (d DNSResolverDomainOverrideDataSourceModel) GetAttrType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"domain":       types.StringType,
		"ip_address":   types.StringType,
		"tls_queries":  types.BoolType,
		"tls_hostname": types.StringType,
		"description":  types.StringType,
	}}
}

func (d *DNSResolverDomainOverrideDataSourceModel) SetFromValue(ctx context.Context, domainOverride *pfsense.DomainOverride) diag.Diagnostics {
	d.Domain = types.StringValue(domainOverride.Domain)
	d.IPAddress = types.StringValue(domainOverride.IPAddress.String())
	d.TLSQueries = types.BoolValue(domainOverride.TLSQueries)

	if domainOverride.TLSHostname != "" {
		d.TLSHostname = types.StringValue(domainOverride.TLSHostname)
	}

	if domainOverride.Description != "" {
		d.Description = types.StringValue(domainOverride.Description)
	}

	return nil
}

func (d *DNSResolverDomainOverridesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_domainoverrides", req.ProviderTypeName)
}

func (d *DNSResolverDomainOverridesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves all DNS resolver domain overrides. Domains for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		MarkdownDescription: "Retrieves all DNS resolver [domain overrides](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-domain-overrides.html). Domains for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		Attributes: map[string]schema.Attribute{
			"all": schema.ListNestedAttribute{
				Description: "All domain overrides.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "Domain whose lookups will be directed to a user-specified DNS lookup server.",
							Computed:    true,
						},
						"ip_address": schema.StringAttribute{
							Description: "IPv4 or IPv6 address (including port) of the authoritative DNS server for this domain.",
							Computed:    true,
						},
						"tls_queries": schema.BoolAttribute{
							Description:         "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to 'false'.",
							MarkdownDescription: "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to `false`.",
							Computed:            true,
						},
						"tls_hostname": schema.StringAttribute{
							Description: "An optional TLS hostname used to verify the server certificate when performing TLS Queries.",
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
	}
}

func (d *DNSResolverDomainOverridesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *DNSResolverDomainOverridesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSResolverDomainOverridesDataSourceModel
	var diags diag.Diagnostics

	domainOverrides, err := d.client.GetDNSResolverDomainOverrides(ctx)
	if addError(&resp.Diagnostics, "Unable to get domain overrides", err) {
		return
	}

	domainOverrideModels := []DNSResolverDomainOverrideDataSourceModel{}
	for _, domainOverride := range *domainOverrides {
		var domainOverrideModel DNSResolverDomainOverrideDataSourceModel
		domainOverride := domainOverride
		diags = domainOverrideModel.SetFromValue(ctx, &domainOverride)
		resp.Diagnostics.Append(diags...)
		domainOverrideModels = append(domainOverrideModels, domainOverrideModel)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.All, diags = types.ListValueFrom(ctx, DNSResolverDomainOverrideDataSourceModel{}.GetAttrType(), domainOverrideModels)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
