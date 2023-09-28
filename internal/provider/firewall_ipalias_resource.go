package provider

import (
	"context"
	"fmt"

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

var _ resource.Resource = &FirewallIPAliasResource{}
var _ resource.ResourceWithImportState = &FirewallIPAliasResource{}

func NewFirewallIPAliasResource() resource.Resource {
	return &FirewallIPAliasResource{}
}

type FirewallIPAliasResource struct {
	client *pfsense.Client
}

type FirewallIPAliasResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Apply       types.Bool   `tfsdk:"apply"`
	Entries     types.List   `tfsdk:"entries"`
}

type FirewallIPAliasEntryResourceModel struct {
	Address     types.String `tfsdk:"address"`
	Description types.String `tfsdk:"description"`
}

func (r FirewallIPAliasEntryResourceModel) GetType() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"address":     types.StringType,
		"description": types.StringType,
	}}
}

func (r *FirewallIPAliasResourceModel) SetFromClient(ctx context.Context, ipAlias *pfsense.FirewallIPAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	r.Name = types.StringValue(ipAlias.Name)

	if ipAlias.Description != "" {
		r.Description = types.StringValue(ipAlias.Description)
	}

	r.Type = types.StringValue(ipAlias.Type)

	entries := []FirewallIPAliasEntryResourceModel{}
	for _, entry := range ipAlias.Entries {
		var entryModel FirewallIPAliasEntryResourceModel

		entryModel.Address = types.StringValue(entry.Address)

		if entry.Description != "" {
			entryModel.Description = types.StringValue(entry.Description)
		}

		entries = append(entries, entryModel)
	}

	r.Entries, diags = types.ListValueFrom(ctx, FirewallIPAliasEntryResourceModel{}.GetType(), entries)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r FirewallIPAliasResourceModel) GetClientValue(ctx context.Context) (*pfsense.FirewallIPAlias, diag.Diagnostics) {
	var ipAlias pfsense.FirewallIPAlias
	var err error
	var diags diag.Diagnostics

	var entryModels []*FirewallIPAliasEntryResourceModel
	diags = r.Entries.ElementsAs(ctx, &entryModels, false)
	if diags.HasError() {
		return nil, diags
	}

	err = ipAlias.SetName(r.Name.ValueString())

	if err != nil {
		diags.AddAttributeError(
			path.Root("name"),
			"Name cannot be parsed",
			err.Error(),
		)
	}

	if !r.Description.IsNull() {
		err = ipAlias.SetDescription(r.Description.ValueString())

		if err != nil {
			diags.AddAttributeError(
				path.Root("description"),
				"Description cannot be parsed",
				err.Error(),
			)
		}
	}

	err = ipAlias.SetType(r.Type.ValueString())

	if err != nil {
		diags.AddAttributeError(
			path.Root("type"),
			"Type cannot be parsed",
			err.Error(),
		)
	}

	for i, entryModel := range entryModels {
		var entry pfsense.FirewallIPAliasEntry

		err = entry.SetAddress(entryModel.Address.ValueString())

		if err != nil {
			diags.AddAttributeError(
				path.Root("entries").AtListIndex(i).AtName("address"),
				"Entry address cannot be parsed",
				err.Error(),
			)
		}

		if !entryModel.Description.IsNull() {
			err = entry.SetDescription(entryModel.Description.ValueString())

			if err != nil {
				diags.AddAttributeError(
					path.Root("entries").AtListIndex(i).AtName("description"),
					"Entry description cannot be parsed",
					err.Error(),
				)
			}
		}

		ipAlias.Entries = append(ipAlias.Entries, entry)
	}

	return &ipAlias, diags
}

func (r *FirewallIPAliasResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_ipalias", req.ProviderTypeName)
}

func (r *FirewallIPAliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Firewall IP alias, defines a group of hosts or networks. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		MarkdownDescription: "Firewall IP [alias](https://docs.netgate.com/pfsense/en/latest/firewall/aliases.html), defines a group of hosts or networks. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of alias.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "For administrative reference (not parsed).",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of alias.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"apply": schema.BoolAttribute{
				Description:         "Apply change, defaults to 'true'.",
				MarkdownDescription: "Apply change, defaults to `true`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			"entries": schema.ListNestedAttribute{
				Description: "Host(s) or network(s).",
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(FirewallIPAliasEntryResourceModel{}.GetType(), []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: "Hosts must be specified by their IP address or fully qualified domain name (FQDN). Networks are specified in CIDR format.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "For administrative reference (not parsed).",
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *FirewallIPAliasResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)

	if !ok {
		return
	}

	r.client = client
}

func (r *FirewallIPAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallIPAliasResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAliasReq, d := data.GetClientValue(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.CreateFirewallIPAlias(ctx, *ipAliasReq)
	if addError(&resp.Diagnostics, "Error creating IP alias", err) {
		return
	}

	diags = data.SetFromClient(ctx, ipAlias)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		if addError(&resp.Diagnostics, "Error applying IP alias", err) {
			return
		}
	}
}

func (r *FirewallIPAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallIPAliasResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.GetFirewallIPAlias(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error reading IP alias", err) {
		return
	}

	diags = data.SetFromClient(ctx, ipAlias)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallIPAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FirewallIPAliasResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAliasReq, d := data.GetClientValue(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.UpdateFirewallIPAlias(ctx, *ipAliasReq)
	if addError(&resp.Diagnostics, "Error updating IP alias", err) {
		return
	}

	diags = data.SetFromClient(ctx, ipAlias)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		if addError(&resp.Diagnostics, "Error applying IP alias", err) {
			return
		}
	}
}

func (r *FirewallIPAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallIPAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewallIPAlias(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error deleting IP alias", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		if addError(&resp.Diagnostics, "Error applying IP alias", err) {
			return
		}
	}
}

func (r *FirewallIPAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
