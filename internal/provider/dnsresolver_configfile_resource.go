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

var (
	_ resource.Resource                = &DNSResolverConfigFileResource{}
	_ resource.ResourceWithImportState = &DNSResolverConfigFileResource{}
)

func NewDNSResolverConfigFileResource() resource.Resource { //nolint:ireturn
	return &DNSResolverConfigFileResource{}
}

type DNSResolverConfigFileResource struct {
	client *pfsense.Client
}

type DNSResolverConfigFileModel struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
	Apply   types.Bool   `tfsdk:"apply"`
}

func (r *DNSResolverConfigFileModel) Set(_ context.Context, configFile pfsense.ConfigFile) diag.Diagnostics {
	r.Name = types.StringValue(configFile.Name)
	r.Content = types.StringValue(configFile.Content)

	return nil
}

func (r DNSResolverConfigFileModel) Value(_ context.Context, configFile *pfsense.ConfigFile) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		path.Root("name"),
		"Name cannot be parsed",
		configFile.SetName(r.Name.ValueString()),
	)

	addPathError(
		&diags,
		path.Root("content"),
		"Content cannot be parsed",
		configFile.SetContent(r.Content.ValueString()),
	)

	return diags
}

func (r *DNSResolverConfigFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_configfile", req.ProviderTypeName)
}

func (r *DNSResolverConfigFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description:         applyDescription,
				MarkdownDescription: applyMarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *DNSResolverConfigFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverConfigFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverConfigFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var configFileReq pfsense.ConfigFile
	resp.Diagnostics.Append(data.Value(ctx, &configFileReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.CreateDNSResolverConfigFile(ctx, configFileReq)
	if addError(&resp.Diagnostics, "Error creating config file", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *configFile)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addWarning(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverConfigFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.GetDNSResolverConfigFile(ctx, data.Name.ValueString())
	if addError(&resp.Diagnostics, "Error reading config file", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *configFile)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverConfigFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverConfigFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var configFileReq pfsense.ConfigFile
	resp.Diagnostics.Append(data.Value(ctx, &configFileReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFile, err := r.client.UpdateDNSResolverConfigFile(ctx, configFileReq)
	if addError(&resp.Diagnostics, "Error updating config file", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *configFile)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		if addWarning(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverConfigFileModel
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
		if addWarning(&resp.Diagnostics, "Error applying config file", err) {
			return
		}
	}
}

func (r *DNSResolverConfigFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
