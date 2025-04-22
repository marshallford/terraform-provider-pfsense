package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = (*DNSResolverHostOverridesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DNSResolverHostOverridesDataSource)(nil)
)

func NewDNSResolverHostOverridesDataSource() datasource.DataSource { //nolint:ireturn
	return &DNSResolverHostOverridesDataSource{}
}

type DNSResolverHostOverridesDataSource struct {
	client *pfsense.Client
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
							Description: DNSResolverHostOverrideModel{}.descriptions()["host"].Description,
							Computed:    true,
						},
						"domain": schema.StringAttribute{
							Description: DNSResolverHostOverrideModel{}.descriptions()["domain"].Description,
							Computed:    true,
						},
						"ip_addresses": schema.ListAttribute{
							Description: DNSResolverHostOverrideModel{}.descriptions()["ip_addresses"].Description,
							Computed:    true,
							ElementType: types.StringType,
						},
						"description": schema.StringAttribute{
							Description: DNSResolverHostOverrideModel{}.descriptions()["description"].Description,
							Computed:    true,
						},
						"fqdn": schema.StringAttribute{
							Description: DNSResolverHostOverrideModel{}.descriptions()["fqdn"].Description,
							Computed:    true,
						},
						"aliases": schema.ListNestedAttribute{
							Description: DNSResolverHostOverrideModel{}.descriptions()["aliases"].Description,
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.StringAttribute{
										Description: DNSResolverHostOverrideAliasModel{}.descriptions()["host"].Description,
										Computed:    true,
									},
									"domain": schema.StringAttribute{
										Description: DNSResolverHostOverrideAliasModel{}.descriptions()["domain"].Description,
										Computed:    true,
									},
									"description": schema.StringAttribute{
										Description: DNSResolverHostOverrideAliasModel{}.descriptions()["description"].Description,
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

func (d *DNSResolverHostOverridesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *DNSResolverHostOverridesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSResolverHostOverridesModel

	hostOverrides, err := d.client.GetDNSResolverHostOverrides(ctx)
	if addError(&resp.Diagnostics, "Unable to get host overrides", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *hostOverrides)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
