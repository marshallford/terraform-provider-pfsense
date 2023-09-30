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
	_ datasource.DataSource              = &FirewallAliasesDataSource{}
	_ datasource.DataSourceWithConfigure = &FirewallAliasesDataSource{}
)

func NewFirewallAliasesDataSource() datasource.DataSource {
	return &FirewallAliasesDataSource{}
}

type FirewallAliasesDataSource struct {
	client *pfsense.Client
}

type FirewallAliasesDataSourceModel struct {
	IP types.List `tfsdk:"ip"`
}

type FirewallIPAliasDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Entries     types.List   `tfsdk:"entries"`
}

func (d FirewallIPAliasDataSourceModel) GetAttrType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"type":        types.StringType,
		"entries":     types.ListType{ElemType: FirewallIPAliasEntryDataSourceModel{}.GetAttrType()},
	}}
}

type FirewallIPAliasEntryDataSourceModel struct {
	Address     types.String `tfsdk:"address"`
	Description types.String `tfsdk:"description"`
}

func (d FirewallIPAliasEntryDataSourceModel) GetAttrType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"address":     types.StringType,
		"description": types.StringType,
	}}
}

func (d *FirewallIPAliasDataSourceModel) SetFromValue(ctx context.Context, ipAlias *pfsense.FirewallIPAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Name = types.StringValue(ipAlias.Name)

	if ipAlias.Description != "" {
		d.Description = types.StringValue(ipAlias.Description)
	}

	d.Type = types.StringValue(ipAlias.Type)

	entries := []FirewallIPAliasEntryDataSourceModel{}
	for _, entry := range ipAlias.Entries {
		var entryModel FirewallIPAliasEntryDataSourceModel

		entryModel.Address = types.StringValue(entry.Address)

		if entry.Description != "" {
			entryModel.Description = types.StringValue(entry.Description)
		}

		entries = append(entries, entryModel)
	}

	d.Entries, diags = types.ListValueFrom(ctx, FirewallIPAliasEntryDataSourceModel{}.GetAttrType(), entries)

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
				Description: "IP aliases (hosts and networks)",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of alias.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "For administrative reference (not parsed).",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Type of alias.",
							Computed:    true,
						},
						"entries": schema.ListNestedAttribute{
							Description: "Host(s) or network(s).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Description: "Hosts must be specified by their IP address or fully qualified domain name (FQDN). Networks are specified in CIDR format.",
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

func (d *FirewallAliasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *FirewallAliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallAliasesDataSourceModel
	var diags diag.Diagnostics

	ipAliases, err := d.client.GetFirewallIPAliases(ctx)
	if addError(&resp.Diagnostics, "Unable to get IP aliases", err) {
		return
	}

	ipAliasModels := []FirewallIPAliasDataSourceModel{}
	for _, ipAlias := range *ipAliases {
		var ipAliasModel FirewallIPAliasDataSourceModel
		ipAlias := ipAlias
		diags = ipAliasModel.SetFromValue(ctx, &ipAlias)
		resp.Diagnostics.Append(diags...)
		ipAliasModels = append(ipAliasModels, ipAliasModel)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.IP, diags = types.ListValueFrom(ctx, FirewallIPAliasDataSourceModel{}.GetAttrType(), ipAliasModels)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
