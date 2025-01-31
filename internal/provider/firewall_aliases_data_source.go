package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = &FirewallAliasesDataSource{}
	_ datasource.DataSourceWithConfigure = &FirewallAliasesDataSource{}
)

type FirewallAliasesModel struct {
	IP   types.List `tfsdk:"ip"`
	Port types.List `tfsdk:"port"`
}

func NewFirewallAliasesDataSource() datasource.DataSource { //nolint:ireturn
	return &FirewallAliasesDataSource{}
}

type FirewallAliasesDataSource struct {
	client *pfsense.Client
}

func (m *FirewallAliasesModel) Set(ctx context.Context, ipAliases pfsense.FirewallIPAliases, portAliases pfsense.FirewallPortAliases) diag.Diagnostics {
	var diags diag.Diagnostics

	ipAliasModels := []FirewallIPAliasModel{}
	for _, ipAlias := range ipAliases {
		var ipAliasModel FirewallIPAliasModel
		diags.Append(ipAliasModel.Set(ctx, ipAlias)...)
		ipAliasModels = append(ipAliasModels, ipAliasModel)
	}

	ipAliasesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallIPAliasModel{}.AttrTypes()}, ipAliasModels)
	diags.Append(newDiags...)
	m.IP = ipAliasesValue

	portAliasModels := []FirewallPortAliasModel{}
	for _, portAlias := range portAliases {
		var portAliasModel FirewallPortAliasModel
		diags.Append(portAliasModel.Set(ctx, portAlias)...)
		portAliasModels = append(portAliasModels, portAliasModel)
	}

	portAliasesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallPortAliasModel{}.AttrTypes()}, portAliasModels)
	diags.Append(newDiags...)
	m.Port = portAliasesValue

	return diags
}

func (d *FirewallAliasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_aliases", req.ProviderTypeName)
}

func (d *FirewallAliasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves all firewall aliases. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		MarkdownDescription: "Retrieves all firewall [aliases](https://docs.netgate.com/pfsense/en/latest/firewall/aliases.html). Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		Attributes: map[string]schema.Attribute{
			"ip": schema.ListNestedAttribute{
				Description: "IP aliases (hosts and networks).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: FirewallIPAliasModel{}.descriptions()["name"].Description,
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: FirewallIPAliasModel{}.descriptions()["description"].Description,
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: FirewallIPAliasModel{}.descriptions()["type"].Description,
							Computed:    true,
						},
						"entries": schema.ListNestedAttribute{
							Description: FirewallIPAliasModel{}.descriptions()["entries"].Description,
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Description: FirewallIPAliasEntryModel{}.descriptions()["address"].Description,
										Computed:    true,
									},
									"description": schema.StringAttribute{
										Description: FirewallIPAliasEntryModel{}.descriptions()["description"].Description,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"port": schema.ListNestedAttribute{
				Description: "Port aliases.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: FirewallPortAliasModel{}.descriptions()["name"].Description,
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: FirewallPortAliasModel{}.descriptions()["description"].Description,
							Computed:    true,
						},
						"entries": schema.ListNestedAttribute{
							Description: FirewallPortAliasModel{}.descriptions()["entries"].Description,
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"port": schema.StringAttribute{
										Description: FirewallPortAliasEntryModel{}.descriptions()["port"].Description,
										Computed:    true,
									},
									"description": schema.StringAttribute{
										Description: FirewallPortAliasEntryModel{}.descriptions()["description"].Description,
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

func (d *FirewallAliasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *FirewallAliasesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallAliasesModel

	ipAliases, err := d.client.GetFirewallIPAliases(ctx)
	if addError(&resp.Diagnostics, "Unable to get IP aliases", err) {
		return
	}

	portAliases, err := d.client.GetFirewallPortAliases(ctx)
	if addError(&resp.Diagnostics, "Unable to get IP aliases", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *ipAliases, *portAliases)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
