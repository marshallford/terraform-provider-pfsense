package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type DNSResolverHostOverridesModel struct {
	All types.List `tfsdk:"all"`
}

type DNSResolverHostOverrideModel struct {
	Host        types.String `tfsdk:"host"`
	Domain      types.String `tfsdk:"domain"`
	IPAddresses types.List   `tfsdk:"ip_addresses"`
	Description types.String `tfsdk:"description"`
	FQDN        types.String `tfsdk:"fqdn"`
	Aliases     types.List   `tfsdk:"aliases"`
}

type DNSResolverHostOverrideAliasModel struct {
	Host        types.String `tfsdk:"host"`
	Domain      types.String `tfsdk:"domain"`
	Description types.String `tfsdk:"description"`
}

func (DNSResolverHostOverrideModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"host": {
			Description: "Name of the host, without the domain part.",
		},
		"domain": {
			Description: "Parent domain of the host.",
		},
		"ip_addresses": {
			Description: "IPv4 or IPv6 addresses to be returned for the host.",
		},
		"description": {
			Description: "For administrative reference (not parsed).",
		},
		"fqdn": {
			Description: "Fully qualified domain name of host.",
		},
		"aliases": {
			Description:         "List of additional names for this host, defaults to '[]'.",
			MarkdownDescription: "List of additional names for this host, defaults to `[]`.",
		},
	}
}

func (DNSResolverHostOverrideAliasModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"host": {
			Description: "Name of the host, without the domain part.",
		},
		"domain": {
			Description: "Parent domain of the host.",
		},
		"description": {
			Description: "For administrative reference (not parsed).",
		},
	}
}

func (DNSResolverHostOverrideModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":         types.StringType,
		"domain":       types.StringType,
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"description":  types.StringType,
		"fqdn":         types.StringType,
		"aliases":      types.ListType{ElemType: types.ObjectType{AttrTypes: DNSResolverHostOverrideAliasModel{}.AttrTypes()}},
	}
}

func (DNSResolverHostOverrideAliasModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"domain":      types.StringType,
		"description": types.StringType,
	}
}

func (m *DNSResolverHostOverridesModel) Set(ctx context.Context, hostOverrides pfsense.HostOverrides) diag.Diagnostics {
	var diags diag.Diagnostics

	hostOverrideModels := []DNSResolverHostOverrideModel{}
	for _, hostOverride := range hostOverrides {
		var hostOverrideModel DNSResolverHostOverrideModel
		diags.Append(hostOverrideModel.Set(ctx, hostOverride)...)
		hostOverrideModels = append(hostOverrideModels, hostOverrideModel)
	}

	hostOverridesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DNSResolverHostOverrideModel{}.AttrTypes()}, hostOverrideModels)
	diags.Append(newDiags...)
	m.All = hostOverridesValue

	return diags
}

func (m *DNSResolverHostOverrideModel) Set(ctx context.Context, hostOverride pfsense.HostOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if hostOverride.Host != "" {
		m.Host = types.StringValue(hostOverride.Host)
	}

	m.Domain = types.StringValue(hostOverride.Domain)

	ipAddressesValue, newDiags := types.ListValueFrom(ctx, types.StringType, hostOverride.IPAddresses)
	diags.Append(newDiags...)
	m.IPAddresses = ipAddressesValue

	if hostOverride.Description != "" {
		m.Description = types.StringValue(hostOverride.Description)
	}

	m.FQDN = types.StringValue(hostOverride.FQDN())

	hostOverrideAliasModels := []DNSResolverHostOverrideAliasModel{}
	for _, hostOverrideAlias := range hostOverride.Aliases {
		var hostOverrideAliasModel DNSResolverHostOverrideAliasModel
		diags.Append(hostOverrideAliasModel.Set(ctx, hostOverrideAlias)...)
		hostOverrideAliasModels = append(hostOverrideAliasModels, hostOverrideAliasModel)
	}

	hostOverrideAliasesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DNSResolverHostOverrideAliasModel{}.AttrTypes()}, hostOverrideAliasModels)
	diags.Append(newDiags...)
	m.Aliases = hostOverrideAliasesValue

	return diags
}

func (m *DNSResolverHostOverrideAliasModel) Set(_ context.Context, hostOverrideAlias pfsense.HostOverrideAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	if hostOverrideAlias.Host != "" {
		m.Host = types.StringValue(hostOverrideAlias.Host)
	}

	m.Domain = types.StringValue(hostOverrideAlias.Domain)

	if hostOverrideAlias.Description != "" {
		m.Description = types.StringValue(hostOverrideAlias.Description)
	}

	return diags
}

func (m DNSResolverHostOverrideModel) Value(ctx context.Context, hostOverride *pfsense.HostOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	if !m.Host.IsNull() {
		addPathError(
			&diags,
			path.Root("host"),
			"Host cannot be parsed",
			hostOverride.SetHost(m.Host.ValueString()),
		)
	}

	addPathError(
		&diags,
		path.Root("domain"),
		"Domain cannot be parsed",
		hostOverride.SetDomain(m.Domain.ValueString()),
	)

	var ipAddresses []string
	if !m.IPAddresses.IsNull() {
		diags.Append(m.IPAddresses.ElementsAs(ctx, &ipAddresses, false)...)
	}

	addPathError(
		&diags,
		path.Root("ip_addresses"),
		"IP addresses cannot be parsed",
		hostOverride.SetIPAddresses(ipAddresses),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			path.Root("description"),
			"Description cannot be parsed",
			hostOverride.SetDescription(m.Description.ValueString()),
		)
	}

	var hostOverrideAliasModels []DNSResolverHostOverrideAliasModel
	if !m.Aliases.IsNull() {
		diags.Append(m.Aliases.ElementsAs(ctx, &hostOverrideAliasModels, false)...)
	}

	hostOverride.Aliases = make([]pfsense.HostOverrideAlias, 0, len(hostOverrideAliasModels))
	for index, hostOverrideAliasModel := range hostOverrideAliasModels {
		var hostOverrideAlias pfsense.HostOverrideAlias

		diags.Append(hostOverrideAliasModel.Value(ctx, &hostOverrideAlias, path.Root("aliases").AtListIndex(index))...)
		hostOverride.Aliases = append(hostOverride.Aliases, hostOverrideAlias)
	}

	return diags
}

func (m DNSResolverHostOverrideAliasModel) Value(_ context.Context, hostOverrideAlias *pfsense.HostOverrideAlias, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !m.Host.IsNull() {
		addPathError(
			&diags,
			attrPath.AtName("host"),
			"Alias host cannot be parsed",
			hostOverrideAlias.SetDomain(m.Domain.ValueString()),
		)
	}

	addPathError(
		&diags,
		attrPath.AtName("domain"),
		"Alias domain cannot be parsed",
		hostOverrideAlias.SetDomain(m.Domain.ValueString()),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			attrPath.AtName("description"),
			"Alias description cannot be parsed",
			hostOverrideAlias.SetDescription(m.Description.ValueString()),
		)
	}

	return diags
}
