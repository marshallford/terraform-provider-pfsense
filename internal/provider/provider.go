package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ provider.Provider = &pfSenseProvider{}
)

func unknownProviderValue(value string) (string, string) {
	return fmt.Sprintf("Unknown pfSense %s", value),
		fmt.Sprintf("The provider cannot create the pfSense client as there is an unknown configuration value for the %s. ", value) +
			"Either target apply the source of the value first, set the value statically in the configuration."
}

func unexpectedConfigureType(value string, providerData any) (string, string) {
	return fmt.Sprintf("Unexpected %s Configure Type", value),
		fmt.Sprintf("Expected *pfsense.Client, got: %T. Please report this issue to the provider developers.", providerData)
}

func configureDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) (*pfsense.Client, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	client, ok := req.ProviderData.(*pfsense.Client)

	if !ok {
		summary, detail := unexpectedConfigureType("Data Source", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return client, ok
}

func configureResourceClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) (*pfsense.Client, bool) {
	if req.ProviderData == nil {
		return nil, false
	}

	client, ok := req.ProviderData.(*pfsense.Client)

	if !ok {
		summary, detail := unexpectedConfigureType("Resource", req.ProviderData)
		resp.Diagnostics.AddError(summary, detail)
	}

	return client, ok
}

func addError(diag *diag.Diagnostics, summary string, err error) bool {
	if err != nil {
		diag.AddError(summary, fmt.Sprintf("unexpected error: %v", err))
		return true
	}
	return false
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pfSenseProvider{
			version: version,
		}
	}
}

type pfSenseProvider struct {
	version string
}

type pfSenseProviderModel struct {
	URL           types.String `tfsdk:"url"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	TLSSkipVerify types.Bool   `tfsdk:"tls_skip_verify"`
	MaxAttempts   types.Int64  `tfsdk:"max_attempts"`
}

func (p *pfSenseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pfsense"
	resp.Version = p.version
}

func (p *pfSenseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Interact with pfSense firewall/router.",
		MarkdownDescription: "Interact with [pfSense](https://www.pfsense.org/) firewall/router.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description:         fmt.Sprintf("pfSense administration URL, defaults to '%s'.", pfsense.DefaultURL),
				MarkdownDescription: fmt.Sprintf("pfSense administration URL, defaults to `%s`.", pfsense.DefaultURL),
				Optional:            true,
			},
			"username": schema.StringAttribute{
				Description:         fmt.Sprintf("pfSense administration username, defaults to '%s'.", pfsense.DefaultUsername),
				MarkdownDescription: fmt.Sprintf("pfSense administration username, defaults to `%s`.", pfsense.DefaultUsername),
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description: "pfSense administration password.",
				Required:    true,
				Sensitive:   true,
			},
			"tls_skip_verify": schema.BoolAttribute{
				Description:         fmt.Sprintf("Skip verification of TLS certificates, defaults to '%t'.", pfsense.DefaultTLSSkipVerify),
				MarkdownDescription: fmt.Sprintf("Skip verification of TLS certificates, defaults to `%t`.", pfsense.DefaultTLSSkipVerify),
				Optional:            true,
			},
			"max_attempts": schema.Int64Attribute{
				Description:         fmt.Sprintf("Maximum number of attempts (only applicable for retryable errors), defaults to '%d'.", pfsense.DefaultMaxAttempts),
				MarkdownDescription: fmt.Sprintf("Maximum number of attempts (only applicable for retryable errors), defaults to `%d`.", pfsense.DefaultMaxAttempts),
				Optional:            true,
			},
		},
	}
}

func (p *pfSenseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring pfSense client")

	var config pfSenseProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.IsUnknown() {
		summary, detail := unknownProviderValue("URL")
		resp.Diagnostics.AddAttributeError(path.Root("url"), summary, detail)
	}

	if config.Username.IsUnknown() {
		summary, detail := unknownProviderValue("username")
		resp.Diagnostics.AddAttributeError(path.Root("username"), summary, detail)
	}

	if config.Username.IsUnknown() {
		summary, detail := unknownProviderValue("password")
		resp.Diagnostics.AddAttributeError(path.Root("password"), summary, detail)
	}

	if config.TLSSkipVerify.IsUnknown() {
		summary, detail := unknownProviderValue("tls_skip_verify")
		resp.Diagnostics.AddAttributeError(path.Root("tls_skip_verify"), summary, detail)
	}

	if config.MaxAttempts.IsUnknown() {
		summary, detail := unknownProviderValue("max_attempts")
		resp.Diagnostics.AddAttributeError(path.Root("max_attempts"), summary, detail)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var opts pfsense.Options

	if !config.URL.IsNull() {
		url, err := url.Parse(config.URL.ValueString())

		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"pfSense URL cannot be parsed",
				err.Error(),
			)
		}

		opts.URL = url
	}

	if !config.Username.IsNull() {
		opts.Username = config.Username.ValueString()
	}

	opts.Password = config.Password.ValueString()

	if !config.TLSSkipVerify.IsNull() {
		opts.TLSSkipVerify = config.TLSSkipVerify.ValueBoolPointer()
	}

	if !config.MaxAttempts.IsNull() {
		i := int(config.MaxAttempts.ValueInt64())
		opts.MaxAttempts = &i
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating pfSense client")

	client, err := pfsense.NewClient(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create pfSense client",
			"An unexpected error occurred when creating the pfSense client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"pfSense client Error: "+err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "pfsense_url", client.Options.URL.String())
	ctx = tflog.SetField(ctx, "pfsense_username", client.Options.Username)
	ctx = tflog.SetField(ctx, "pfsense_password", client.Options.Password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "pfsense_password")

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured pfSense client", map[string]any{"success": true})
}

func (p *pfSenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSystemVersionDataSource,
	}
}

func (p *pfSenseProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSResolverApplyResource,
		NewDNSResolverConfigFileResource,
		NewDNSResolverDomainOverrideResource,
		NewDNSResolverHostOverrideResource,
		NewFirewallFilterReloadResource,
		NewFirewallIPAliasResource,
	}
}
