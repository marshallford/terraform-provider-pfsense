package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var _ provider.Provider = &pfSenseProvider{}

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
	var data pfSenseProviderModel

	tflog.Info(ctx, "Configuring pfSense client")

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.URL.IsUnknown() {
		path := path.Root("url")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if data.Username.IsUnknown() {
		path := path.Root("username")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if data.Password.IsUnknown() {
		path := path.Root("password")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if data.TLSSkipVerify.IsUnknown() {
		path := path.Root("tls_skip_verify")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if data.MaxAttempts.IsUnknown() {
		path := path.Root("max_attempts")
		summary, detail := unknownProviderValue(path)
		resp.Diagnostics.AddAttributeError(path, summary, detail)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var opts pfsense.Options

	if !data.URL.IsNull() {
		url, err := url.Parse(data.URL.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("url"),
				"pfSense URL cannot be parsed",
				err.Error(),
			)
		}

		opts.URL = url
	}

	if !data.Username.IsNull() {
		opts.Username = data.Username.ValueString()
	}

	opts.Password = data.Password.ValueString()

	if !data.TLSSkipVerify.IsNull() {
		opts.TLSSkipVerify = data.TLSSkipVerify.ValueBoolPointer()
	}

	if !data.MaxAttempts.IsNull() {
		i := int(data.MaxAttempts.ValueInt64())
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
				"pfSense client URL: "+opts.URL.String()+"\n"+
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
	resp.EphemeralResourceData = client

	tflog.Info(ctx, "Configured pfSense client", map[string]any{"success": true})
}

func (p *pfSenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSResolverDomainOverridesDataSource,
		NewDNSResolverHostOverridesDataSource,
		NewFirewallAliasesDataSource,
		NewSystemVersionDataSource,
		NewDHCPv4StaticMappingsDataSource,
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
		NewFirewallPortAliasResource,
		NewDHCPv4ApplyResource,
		NewDHCPv4StaticMappingResource,
	}
}
