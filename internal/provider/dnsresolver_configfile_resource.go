package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var _ resource.Resource = &DNSResolverConfigFileResource{}
var _ resource.ResourceWithImportState = &DNSResolverConfigFileResource{}

func NewDNSResolverConfigFileResource() resource.Resource {
	return &DNSResolverConfigFileResource{}
}

type DNSResolverConfigFileResource struct {
	client *pfsense.Client
}

type DNSResolverConfigFileResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Apply   types.Bool   `tfsdk:"apply"`
}

func (r *DNSResolverConfigFileResourceModel) SetFromValue(ctx context.Context, configFile *pfsense.ConfigFile) diag.Diagnostics {
	r.Name = types.StringValue(configFile.Name)
	r.Content = types.StringValue(configFile.Content)

	return nil
}

func (r DNSResolverConfigFileResourceModel) Value(ctx context.Context) (*pfsense.ConfigFile, diag.Diagnostics) {
	var configFile pfsense.ConfigFile
	var err error
	var diags diag.Diagnostics

	err = configFile.SetName(r.Name.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("name"),
			"Name cannot be parsed",
			err.Error(),
		)
	}

	err = configFile.SetContent(r.Content.ValueString())
	if err != nil {
		diags.AddAttributeError(
			path.Root("content"),
			"Content cannot be parsed",
			err.Error(),
		)
	}

	return &configFile, diags
}

func (r *DNSResolverConfigFileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_configfile", req.ProviderTypeName)
}

func (r *DNSResolverConfigFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS resolver (Unbound) config file. Prerequisite: Must add the directive 'include-toplevel: /var/unbound/conf.d/*' to the DNS resolver custom options input. Use with caution, content is not checked/validated.",
		MarkdownDescription: "DNS resolver (Unbound) [config file](https://man.freebsd.org/cgi/man.cgi?unbound.conf). **Prerequisite**: Must add the directive `include-toplevel: /var/unbound/conf.d/*` to the DNS resolver custom options input. **Use with caution**, content is not checked/validated.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of config file.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Description:         "Contents of file. Must specify Unbound clause(s). Comments start with '#' and last to the end of line.",
				MarkdownDescription: "Contents of file. Must specify Unbound clause(s). Comments start with `#` and last to the end of line.",
				Required:            true,
			},
			"apply": schema.BoolAttribute{
				Description:         "Apply change, defaults to 'true'.",
				MarkdownDescription: "Apply change, defaults to `true`.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *DNSResolverConfigFileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverConfigFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverConfigFileResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFileReq, d := data.Value(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.CreateDNSResolverConfigFile(ctx, *configFileReq)
	if addError(&resp.Diagnostics, "Error creating config file", err) {
		return
	}

	diags = data.SetFromValue(ctx, configFile)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addError(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverConfigFileResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.GetDNSResolverConfigFile(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error reading config file", err) {
		return
	}

	diags = data.SetFromValue(ctx, configFile)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverConfigFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverConfigFileResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFileReq, d := data.Value(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.UpdateDNSResolverConfigFile(ctx, *configFileReq)
	if addError(&resp.Diagnostics, "Error updating config file", err) {
		return
	}

	diags = data.SetFromValue(ctx, configFile)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addError(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverConfigFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSResolverConfigFile(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error deleting config file", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addError(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
