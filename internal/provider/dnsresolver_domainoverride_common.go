package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type DNSResolverDomainOverridesModel struct {
	All types.List `tfsdk:"all"`
}

type DNSResolverDomainOverrideModel struct {
	Domain      types.String `tfsdk:"domain"`
	IPAddress   types.String `tfsdk:"ip_address"`
	TLSHostname types.String `tfsdk:"tls_hostname"`
	Description types.String `tfsdk:"description"`
	TLSQueries  types.Bool   `tfsdk:"tls_queries"` // unordered to avoid maligned error
	Apply       types.Bool   `tfsdk:"apply"`
}

func (DNSResolverDomainOverrideModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"domain": {
			Description: "Domain whose lookups will be directed to a user-specified DNS lookup server.",
		},
		"ip_address": {
			Description: "IPv4 or IPv6 address (including port) of the authoritative DNS server for this domain.",
		},
		"tls_queries": {
			Description:         "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to 'false'.",
			MarkdownDescription: "Queries to all DNS servers for this domain will be sent using SSL/TLS, defaults to `false`.",
		},
		"tls_hostname": {
			Description: "A TLS hostname used to verify the server certificate when performing TLS Queries.",
		},
		"description": {
			Description: "For administrative reference (not parsed).",
		},
		"apply": {
			Description:         "Apply change, defaults to 'true'.",
			MarkdownDescription: "Apply change, defaults to `true`.",
		},
	}
}

func (DNSResolverDomainOverrideModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":       types.StringType,
		"ip_address":   types.StringType,
		"tls_queries":  types.BoolType,
		"tls_hostname": types.StringType,
		"description":  types.StringType,
	}
}

func (m *DNSResolverDomainOverridesModel) Set(ctx context.Context, domainOverrides pfsense.DomainOverrides) diag.Diagnostics {
	var diags diag.Diagnostics

	domainOverrideModels := []DNSResolverDomainOverrideModel{}
	for _, domainOverride := range domainOverrides {
		var domainOverrideModel DNSResolverDomainOverrideModel
		diags.Append(domainOverrideModel.Set(ctx, domainOverride)...)
		domainOverrideModels = append(domainOverrideModels, domainOverrideModel)
	}

	domainOverridesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DNSResolverDomainOverrideModel{}.AttrTypes()}, domainOverrideModels)
	diags.Append(newDiags...)
	m.All = domainOverridesValue

	return diags
}

func (m *DNSResolverDomainOverrideModel) Set(_ context.Context, domainOverride pfsense.DomainOverride) diag.Diagnostics {
	m.Domain = types.StringValue(domainOverride.Domain)
	m.IPAddress = types.StringValue(domainOverride.IPAddress.String())
	m.TLSQueries = types.BoolValue(domainOverride.TLSQueries)

	if domainOverride.TLSHostname != "" {
		m.TLSHostname = types.StringValue(domainOverride.TLSHostname)
	}

	if domainOverride.Description != "" {
		m.Description = types.StringValue(domainOverride.Description)
	}

	return nil
}

func (m DNSResolverDomainOverrideModel) Value(_ context.Context, domainOverride *pfsense.DomainOverride) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		path.Root("domain"),
		"Domain cannot be parsed",
		domainOverride.SetDomain(m.Domain.ValueString()),
	)

	addPathError(
		&diags,
		path.Root("ip_address"),
		"IP address cannot be parsed",
		domainOverride.SetIPAddress(m.IPAddress.ValueString()),
	)

	addPathError(
		&diags,
		path.Root("tls_queries"),
		"TLS Queries cannot be parsed",
		domainOverride.SetTLSQueries(m.TLSQueries.ValueBool()),
	)

	if !m.TLSHostname.IsNull() {
		addPathError(
			&diags,
			path.Root("tls_hostname"),
			"TLS Hostname cannot be parsed",
			domainOverride.SetTLSHostname(m.TLSHostname.ValueString()),
		)
	}

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			path.Root("description"),
			"Description cannot be parsed",
			domainOverride.SetDescription(m.Description.ValueString()),
		)
	}

	return diags
}
