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
	_ resource.Resource                = &FirewallPortAliasResource{}
	_ resource.ResourceWithImportState = &FirewallPortAliasResource{}
)

type FirewallPortAliasResourceModel struct {
	FirewallPortAliasModel
	Apply types.Bool `tfsdk:"apply"`
}

func NewFirewallPortAliasResource() resource.Resource { //nolint:ireturn
	return &FirewallPortAliasResource{}
}

type FirewallPortAliasResource struct {
	client *pfsense.Client
}

func (r *FirewallPortAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_firewall_port_alias", req.ProviderTypeName)
}

func (r *FirewallPortAliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Firewall port alias, defines a group of ports and/or port ranges. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		MarkdownDescription: "Firewall port [alias](https://docs.netgate.com/pfsense/en/latest/firewall/aliases.html), defines a group of ports and/or port ranges. Aliases can be referenced by firewall rules, port forwards, outbound NAT rules, and other places in the firewall.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: FirewallPortAliasModel{}.descriptions()["name"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringIsAlias(),
				},
			},
			"description": schema.StringAttribute{
				Description: FirewallPortAliasModel{}.descriptions()["description"].Description,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
				Description: FirewallPortAliasModel{}.descriptions()["entries"].Description,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: FirewallPortAliasEntryModel{}.AttrTypes()}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							Description: FirewallPortAliasEntryModel{}.descriptions()["port"].Description,
							Required:    true,
							Validators: []validator.String{
								// TODO aliases effectively disables the port validation.
								stringvalidator.Any(stringIsPort(), stringIsPortRange(), stringIsAlias()),
							},
						},
						"description": schema.StringAttribute{
							Description: FirewallPortAliasEntryModel{}.descriptions()["description"].Description,
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

func (r *FirewallPortAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *FirewallPortAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallPortAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var portAliasReq pfsense.FirewallPortAlias
	resp.Diagnostics.Append(data.Value(ctx, &portAliasReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	portAlias, err := r.client.CreateFirewallPortAlias(ctx, portAliasReq)
	if addError(&resp.Diagnostics, "Error creating port alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *portAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		addWarning(&resp.Diagnostics, "Error applying port alias", err)
	}
}

func (r *FirewallPortAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallPortAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	portAlias, err := r.client.GetFirewallPortAlias(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error reading port alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Value(ctx, portAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallPortAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FirewallPortAliasResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var portAliasReq pfsense.FirewallPortAlias
	resp.Diagnostics.Append(data.Value(ctx, &portAliasReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	portAlias, err := r.client.UpdateFirewallPortAlias(ctx, portAliasReq)
	if addError(&resp.Diagnostics, "Error updating port alias", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *portAlias)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		addWarning(&resp.Diagnostics, "Error applying port alias", err)
	}
}

func (r *FirewallPortAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallPortAliasResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewallPortAlias(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error deleting port alias", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ReloadFirewallFilter(ctx)
		addWarning(&resp.Diagnostics, "Error applying port alias", err)
	}
}

func (r *FirewallPortAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
