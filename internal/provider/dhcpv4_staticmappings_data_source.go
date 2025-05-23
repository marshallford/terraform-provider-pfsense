package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = (*DHCPv4StaticMappingsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DHCPv4StaticMappingsDataSource)(nil)
)

func NewDHCPv4StaticMappingsDataSource() datasource.DataSource { //nolint:ireturn
	return &DHCPv4StaticMappingsDataSource{}
}

type DHCPv4StaticMappingsDataSource struct {
	client *pfsense.Client
}

func (d *DHCPv4StaticMappingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dhcpv4_staticmappings", req.ProviderTypeName)
}

func (d *DHCPv4StaticMappingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves all DHCPv4 static mappings. Static mappings express a preference for which IP address will be assigned to a given client based on its MAC address. In a network where unknown clients are denied, this also serves as a list of known clients which are allowed to receive leases or have static ARP entries.",
		MarkdownDescription: "Retrieves all DHCPv4 [static mappings](https://docs.netgate.com/pfsense/en/latest/services/dhcp/ipv4.html#static-mappings). Static mappings express a preference for which IP address will be assigned to a given client based on its MAC address. In a network where unknown clients are denied, this also serves as a list of known clients which are allowed to receive leases or have static ARP entries.",
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Description: "Network interface.",
				Required:    true,
				Validators: []validator.String{
					stringIsInterface(),
				},
			},
			"all": schema.ListNestedAttribute{
				Description: "All static mappings.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["interface"].Description,
							Computed:    true,
						},
						"mac_address": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["mac_address"].Description,
							CustomType:  macAddressType{},
							Computed:    true,
						},
						"client_identifier": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["client_identifier"].Description,
							Computed:    true,
						},
						"ip_address": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["ip_address"].Description,
							Computed:    true,
						},
						"arp_table_static_entry": schema.BoolAttribute{
							Description:         DHCPv4StaticMappingModel{}.descriptions()["arp_table_static_entry"].Description,
							MarkdownDescription: DHCPv4StaticMappingModel{}.descriptions()["arp_table_static_entry"].MarkdownDescription,
							Computed:            true,
						},
						"hostname": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["hostname"].Description,
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["description"].Description,
							Computed:    true,
						},
						"wins_servers": schema.ListAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["wins_servers"].Description,
							Computed:    true,
							ElementType: types.StringType,
						},
						"dns_servers": schema.ListAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["dns_servers"].Description,
							Computed:    true,
							ElementType: types.StringType,
						},
						"gateway": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["gateway"].Description,
							Computed:    true,
						},
						"domain_name": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["domain_name"].Description,
							Computed:    true,
						},
						"domain_search_list": schema.ListAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["domain_search_list"].Description,
							Computed:    true,
							ElementType: types.StringType,
						},
						"default_lease_time": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["default_lease_time"].Description,
							Computed:    true,
							CustomType:  timetypes.GoDurationType{},
						},
						"maximum_lease_time": schema.StringAttribute{
							Description: DHCPv4StaticMappingModel{}.descriptions()["maximum_lease_time"].Description,
							Computed:    true,
							CustomType:  timetypes.GoDurationType{},
						},
					},
				},
			},
		},
	}
}

func (d *DHCPv4StaticMappingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *DHCPv4StaticMappingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DHCPv4StaticMappingsModel
	resp.Diagnostics.Append(resp.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	staticMappings, err := d.client.GetDHCPv4StaticMappings(ctx, data.Interface.ValueString())
	if addError(&resp.Diagnostics, "Unable to get static mappings", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *staticMappings)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
