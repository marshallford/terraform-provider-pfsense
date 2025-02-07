package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ resource.Resource                = &FirewallIPAliasResource{}
	_ resource.ResourceWithImportState = &FirewallIPAliasResource{}
)

type FirewallIPAliasResourceModel struct {
	FirewallIPAliasModel
	Apply types.Bool `tfsdk:"apply"`
}

func NewFirewallIPAliasResource() resource.Resource { //nolint:ireturn
	return &FirewallIPAliasResource{}
}

type FirewallIPAliasResource struct {
	client *pfsense.Client
}

func (r *FirewallIPAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_ip_alias", req.ProviderTypeName)
}

func (r *FirewallIPAliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Firewall IP alias, defines a group of hosts and/or networks. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		MarkdownDescription: "Firewall IP [alias](https://docs.netgate.com/pfsense/en/latest/firewall/aliases.html), defines a group of hosts and/or networks. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: FirewallIPAliasModel{}.descriptions()["name"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringIsAlias(),
				},
			},
			"description": schema.StringAttribute{
				Description: FirewallIPAliasModel{}.descriptions()["description"].Description,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"type": schema.StringAttribute{
				Description:         FirewallIPAliasModel{}.descriptions()["type"].Description,
				MarkdownDescription: FirewallIPAliasModel{}.descriptions()["type"].MarkdownDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(pfsense.FirewallIPAlias{}.Types()...),
				},
			},
			"apply": schema.BoolAttribute{
				Description:         applyDescription,
				MarkdownDescription: applyMarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(defaultApply),
			},
			"entries": schema.ListNestedAttribute{
				Description: FirewallIPAliasModel{}.descriptions()["entries"].Description,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: FirewallIPAliasEntryModel{}.AttrTypes()}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Description: FirewallIPAliasEntryModel{}.descriptions()["address"].Description,
							Required:    true,
							Validators: []validator.String{
								// https://github.com/hashicorp/terraform-plugin-framework-validators/issues/113
								stringvalidator.Any(stringIsNetwork(), stringIsIPAddress("any"), stringIsDomain(), stringIsAlias()),
							},
						},
						"description": schema.StringAttribute{
							Description: FirewallIPAliasEntryModel{}.descriptions()["description"].Description,
							Computed:    true,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func (r *FirewallIPAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *FirewallIPAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallIPAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ipAliasReq pfsense.FirewallIPAlias
	resp.Diagnostics.Append(data.Value(ctx, &ipAliasReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.CreateFirewallIPAlias(ctx, ipAliasReq)
	if addError(&resp.Diagnostics, "Error creating IP alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *ipAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		addWarning(&resp.Diagnostics, "Error applying IP alias", err)
	}
}

func (r *FirewallIPAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallIPAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.GetFirewallIPAlias(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error reading IP alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Value(ctx, ipAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallIPAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FirewallIPAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var ipAliasReq pfsense.FirewallIPAlias
	resp.Diagnostics.Append(data.Value(ctx, &ipAliasReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAlias, err := r.client.UpdateFirewallIPAlias(ctx, ipAliasReq)
	if addError(&resp.Diagnostics, "Error updating IP alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *ipAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		addWarning(&resp.Diagnostics, "Error applying IP alias", err)
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
		addWarning(&resp.Diagnostics, "Error applying IP alias", err)
	}
}

func (r *FirewallIPAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
