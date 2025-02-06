package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var (
	_ resource.Resource                = &DHCPDV4StaticMappingResource{}
	_ resource.ResourceWithImportState = &DHCPDV4StaticMappingResource{}
)

type DHCPDV4StaticMappingResourceModel struct {
	DHCPDV4StaticMappingModel
	Apply types.Bool `tfsdk:"apply"`
}

func NewDHCPDV4StaticMappingResource() resource.Resource { //nolint:ireturn
	return &DHCPDV4StaticMappingResource{}
}

type DHCPDV4StaticMappingResource struct {
	client *pfsense.Client
}

func (r *DHCPDV4StaticMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dhcpdv4_staticmapping", req.ProviderTypeName)
}

func (r *DHCPDV4StaticMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DHCPD v4 static mapping. Static DHCP mappings express a preference for which IP address will be assigned to a given client based on its MAC address. In a network where unknown clients are denied, this also serves as a list of known clients which are allowed to receive leases or have static ARP entries.",
		MarkdownDescription: "DHCPD v4 [static mapping](https://docs.netgate.com/pfsense/en/latest/services/dhcp/ipv4.html#static-mappings). Static DHCP mappings express a preference for which IP address will be assigned to a given client based on its MAC address. In a network where unknown clients are denied, this also serves as a list of known clients which are allowed to receive leases or have static ARP entries.",
		Attributes: map[string]schema.Attribute{
			"interface": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["interface"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mac_address": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["mac_address"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_identifier": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["client_identifier"].Description,
				Optional:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["ip_address"].Description,
				Optional:    true,
			},
			"arp_table_static_entry": schema.BoolAttribute{
				Description:         DHCPDV4StaticMappingModel{}.descriptions()["arp_table_static_entry"].Description,
				MarkdownDescription: DHCPDV4StaticMappingModel{}.descriptions()["arp_table_static_entry"].MarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(defaultStaticMappingARPTableStaticEntry),
			},
			"hostname": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["hostname"].Description,
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["description"].Description,
				Optional:    true,
			},
			"wins_servers": schema.ListAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["wins_servers"].Description,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				ElementType: types.StringType,
			},
			"dns_servers": schema.ListAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["dns_servers"].Description,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				ElementType: types.StringType,
			},
			"gateway": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["gateway"].Description,
				Optional:    true,
			},
			"domain_name": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["domain_name"].Description,
				Optional:    true,
			},
			"domain_search_list": schema.ListAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["domain_search_list"].Description,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				ElementType: types.StringType,
			},
			"default_lease_time": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["default_lease_time"].Description,
				Optional:    true,
				CustomType:  timetypes.GoDurationType{},
			},
			"maximum_lease_time": schema.StringAttribute{
				Description: DHCPDV4StaticMappingModel{}.descriptions()["maximum_lease_time"].Description,
				Optional:    true,
				CustomType:  timetypes.GoDurationType{},
			},
			"apply": schema.BoolAttribute{
				Description:         applyDescription,
				MarkdownDescription: applyMarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(defaultApply),
			},
		},
	}
}

func (r *DHCPDV4StaticMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *DHCPDV4StaticMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DHCPDV4StaticMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var staticMappingReq pfsense.DHCPDV4StaticMapping
	resp.Diagnostics.Append(data.Value(ctx, &staticMappingReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	staticMapping, err := r.client.CreateDHCPDV4StaticMapping(ctx, staticMappingReq)
	if addError(&resp.Diagnostics, "Error creating static mapping", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *staticMapping)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDHCPDV4Changes(ctx, data.Interface.ValueString())
		addWarning(&resp.Diagnostics, "Error applying static mapping", err)
	}
}

func (r *DHCPDV4StaticMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DHCPDV4StaticMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	staticMapping, err := r.client.GetDHCPDV4StaticMapping(ctx, data.Interface.ValueString(), data.MACAddress.ValueString())
	if addError(&resp.Diagnostics, "Error reading static mapping", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *staticMapping)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DHCPDV4StaticMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DHCPDV4StaticMappingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var staticMappingReq pfsense.DHCPDV4StaticMapping
	resp.Diagnostics.Append(data.Value(ctx, &staticMappingReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	staticMapping, err := r.client.UpdateDHCPDV4StaticMapping(ctx, staticMappingReq)
	if addError(&resp.Diagnostics, "Error updating static mapping", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *staticMapping)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDHCPDV4Changes(ctx, data.Interface.ValueString())
		addWarning(&resp.Diagnostics, "Error applying static mapping", err)
	}
}

func (r *DHCPDV4StaticMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DHCPDV4StaticMappingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDHCPDV4StaticMapping(ctx, data.Interface.ValueString(), data.MACAddress.ValueString())
	if addError(&resp.Diagnostics, "Error deleting static mapping", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDHCPDV4Changes(ctx, data.Interface.ValueString())
		addWarning(&resp.Diagnostics, "Error applying static mapping", err)
	}
}

func (r *DHCPDV4StaticMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: interface,mac_address. Got: %q", req.ID),
		)

		return
	}

	var staticMapping pfsense.DHCPDV4StaticMapping

	if addError(&resp.Diagnostics, "Interface cannot be parsed", staticMapping.SetInterface(idParts[0])) {
		return
	}

	if addError(&resp.Diagnostics, "MAC address cannot be parsed", staticMapping.SetMACAddress(idParts[1])) {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("interface"), staticMapping.Interface)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mac_address"), staticMapping.MACAddress)...)
}
