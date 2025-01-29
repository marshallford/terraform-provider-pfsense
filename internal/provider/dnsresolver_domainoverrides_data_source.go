package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = &DNSResolverDomainOverridesDataSource{}
	_ datasource.DataSourceWithConfigure = &DNSResolverDomainOverridesDataSource{}
)

func NewDNSResolverDomainOverridesDataSource() datasource.DataSource { //nolint:ireturn
	return &DNSResolverDomainOverridesDataSource{}
}

type DNSResolverDomainOverridesDataSource struct {
	client *pfsense.Client
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
							Description: DNSResolverDomainOverrideModel{}.descriptions()["domain"].Description,
							Computed:    true,
						},
						"ip_address": schema.StringAttribute{
							Description: DNSResolverDomainOverrideModel{}.descriptions()["ip_address"].Description,
							Computed:    true,
						},
						"tls_queries": schema.BoolAttribute{
							Description:         DNSResolverDomainOverrideModel{}.descriptions()["tls_queries"].Description,
							MarkdownDescription: DNSResolverDomainOverrideModel{}.descriptions()["tls_queries"].MarkdownDescription,
							Computed:            true,
						},
						"tls_hostname": schema.StringAttribute{
							Description: DNSResolverDomainOverrideModel{}.descriptions()["tls_hostname"].Description,
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: DNSResolverDomainOverrideModel{}.descriptions()["description"].Description,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSResolverDomainOverridesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *DNSResolverDomainOverridesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSResolverDomainOverridesModel

	domainOverrides, err := d.client.GetDNSResolverDomainOverrides(ctx)
	if addError(&resp.Diagnostics, "Unable to get domain overrides", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *domainOverrides)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
