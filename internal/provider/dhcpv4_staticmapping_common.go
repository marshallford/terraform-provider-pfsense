package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

// TODO Handle pfSense "form placeholder" defaulting.
// Simple attrs: hardcoded defaults: arp_table_static_entry=false, default_lease_time=2h, maximum_lease_time=24h
// Complicated: wins_servers, dns_servers, gateway, domain_name, domain_search_list
// Since the HTML form placeholders are not actually saved with the static mapping
// the client would need make the same calls that the php page does, which is prone to error/change.
// In addition, the client pkg would need to include an extra func to get all these values
// and the provider pkg would need to merge the result and mark these attributes as computed

type DHCPv4StaticMappingsModel struct {
	Interface types.String `tfsdk:"interface"`
	All       types.List   `tfsdk:"all"`
}

type DHCPv4StaticMappingModel struct {
	Interface           types.String         `tfsdk:"interface"`
	MACAddress          types.String         `tfsdk:"mac_address"`
	ClientIdentifier    types.String         `tfsdk:"client_identifier"`
	IPAddress           types.String         `tfsdk:"ip_address"`
	ARPTableStaticEntry types.Bool           `tfsdk:"arp_table_static_entry"`
	Hostname            types.String         `tfsdk:"hostname"`
	Description         types.String         `tfsdk:"description"`
	WINSServers         types.List           `tfsdk:"wins_servers"`
	DNSServers          types.List           `tfsdk:"dns_servers"`
	Gateway             types.String         `tfsdk:"gateway"`
	DomainName          types.String         `tfsdk:"domain_name"`
	DomainSearchList    types.List           `tfsdk:"domain_search_list"`
	DefaultLeaseTime    timetypes.GoDuration `tfsdk:"default_lease_time"`
	MaximumLeaseTime    timetypes.GoDuration `tfsdk:"maximum_lease_time"`
}

func (DHCPv4StaticMappingModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"interface": {
			Description: "Network interface. Each interface has its own separate DHCP configuration (including static mappings).",
		},
		"mac_address": {
			Description: "MAC address of the client to match.",
		},
		"client_identifier": {
			Description: "Identifier to match based on the value sent by the client (RFC 2132).",
		},
		"ip_address": {
			Description: "IPv4 address to assign this client. Address must be outside of any defined pools. If no IPv4 address is given, one will be dynamically allocated from a pool.",
		},
		"arp_table_static_entry": {
			Description:         fmt.Sprintf("Create an ARP Table Static Entry for this MAC & IP Address pair., defaults to '%t'.", defaultStaticMappingARPTableStaticEntry),
			MarkdownDescription: fmt.Sprintf("Create an ARP Table Static Entry for this MAC & IP Address pair., defaults to `%t`.", defaultStaticMappingARPTableStaticEntry),
		},
		"hostname": {
			Description: "Name of the host, without the domain part.",
		},
		"description": {
			Description: descriptionDescription,
		},
		"wins_servers": {
			Description: "WINS (Windows Internet Name Service) servers provided to the client.",
		},
		"dns_servers": {
			Description: "DNS (Domain Name System) servers provided to the client.",
		},
		"gateway": {
			Description: "IPv4 gateway address.",
		},
		"domain_name": {
			Description: "Domain name passed to the client to form its fully qualified hostname.",
		},
		"domain_search_list": {
			Description: "DNS search domains that are provided to the client.",
		},
		"default_lease_time": {
			Description: "Default lease time for clients that do not ask for a specific lease expiration time.",
		},
		"maximum_lease_time": {
			Description: "Maximum lease time for clients that ask for a specific lease expiration time.",
		},
	}
}

func (DHCPv4StaticMappingModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface":              types.StringType,
		"mac_address":            types.StringType,
		"client_identifier":      types.StringType,
		"ip_address":             types.StringType,
		"arp_table_static_entry": types.BoolType,
		"hostname":               types.StringType,
		"description":            types.StringType,
		"wins_servers":           types.ListType{ElemType: types.StringType},
		"dns_servers":            types.ListType{ElemType: types.StringType},
		"gateway":                types.StringType,
		"domain_name":            types.StringType,
		"domain_search_list":     types.ListType{ElemType: types.StringType},
		"default_lease_time":     timetypes.GoDurationType{},
		"maximum_lease_time":     timetypes.GoDurationType{},
	}
}

func (m *DHCPv4StaticMappingsModel) Set(ctx context.Context, staticMappings pfsense.DHCPv4StaticMappings) diag.Diagnostics {
	var diags diag.Diagnostics

	staticMappingModels := []DHCPv4StaticMappingModel{}
	for _, staticMapping := range staticMappings {
		var staticMappingModel DHCPv4StaticMappingModel
		diags.Append(staticMappingModel.Set(ctx, staticMapping)...)
		staticMappingModels = append(staticMappingModels, staticMappingModel)
	}

	staticMappingsValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DHCPv4StaticMappingModel{}.AttrTypes()}, staticMappingModels)
	diags.Append(newDiags...)
	m.All = staticMappingsValue

	return diags
}

func (m *DHCPv4StaticMappingModel) Set(ctx context.Context, staticMapping pfsense.DHCPv4StaticMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Interface = types.StringValue(staticMapping.Interface)
	m.MACAddress = types.StringValue(staticMapping.MACAddress.String())

	if staticMapping.ClientIdentifier != "" {
		m.ClientIdentifier = types.StringValue(staticMapping.ClientIdentifier)
	}

	if staticMapping.StringifyIPAddress() != "" {
		m.IPAddress = types.StringValue(staticMapping.StringifyIPAddress())
	}

	m.ARPTableStaticEntry = types.BoolValue(staticMapping.ARPTableStaticEntry)

	if staticMapping.Hostname != "" {
		m.Hostname = types.StringValue(staticMapping.Hostname)
	}

	if staticMapping.Description != "" {
		m.Description = types.StringValue(staticMapping.Description)
	}

	winsServersValue, newDiags := types.ListValueFrom(ctx, types.StringType, staticMapping.StringifyWINSServers())
	diags.Append(newDiags...)
	m.WINSServers = winsServersValue

	dnsServersValue, newDiags := types.ListValueFrom(ctx, types.StringType, staticMapping.StringifyDNSServers())
	diags.Append(newDiags...)
	m.DNSServers = dnsServersValue

	if staticMapping.StringifyGateway() != "" {
		m.Gateway = types.StringValue(staticMapping.StringifyGateway())
	}

	if staticMapping.DomainName != "" {
		m.DomainName = types.StringValue(staticMapping.DomainName)
	}

	domainSearchListValue, newDiags := types.ListValueFrom(ctx, types.StringType, staticMapping.DomainSearchList)
	diags.Append(newDiags...)
	m.DomainSearchList = domainSearchListValue

	// TODO RFC2131 allows for a 0 second DHCP lease (not sure if pfSense does), consider using *time.Duration to fix.
	if staticMapping.DefaultLeaseTime != 0 {
		m.DefaultLeaseTime = timetypes.NewGoDurationValue(staticMapping.DefaultLeaseTime)
	}

	if staticMapping.MaximumLeaseTime != 0 {
		m.MaximumLeaseTime = timetypes.NewGoDurationValue(staticMapping.MaximumLeaseTime)
	}

	return diags
}

func (m DHCPv4StaticMappingModel) Value(ctx context.Context, staticMapping *pfsense.DHCPv4StaticMapping) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		path.Root("interface"),
		"Interface cannot be parsed",
		staticMapping.SetInterface(m.Interface.ValueString()),
	)

	addPathError(
		&diags,
		path.Root("mac_address"),
		"MAC address cannot be parsed",
		staticMapping.SetMACAddress(m.MACAddress.ValueString()),
	)

	if !m.ClientIdentifier.IsNull() {
		addPathError(
			&diags,
			path.Root("client_identifier"),
			"Client identifier cannot be parsed",
			staticMapping.SetClientIdentifier(m.ClientIdentifier.ValueString()),
		)
	}

	if !m.IPAddress.IsNull() {
		addPathError(
			&diags,
			path.Root("ip_address"),
			"IP address cannot be parsed",
			staticMapping.SetIPAddress(m.IPAddress.ValueString()),
		)
	}

	addPathError(
		&diags,
		path.Root("arp_table_static_entry"),
		"ARP table static entry cannot be parsed",
		staticMapping.SetARPTableStaticEntry(m.ARPTableStaticEntry.ValueBool()),
	)

	if !m.Hostname.IsNull() {
		addPathError(
			&diags,
			path.Root("hostname"),
			"Hostname cannot be parsed",
			staticMapping.SetHostname(m.Hostname.ValueString()),
		)
	}

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			path.Root("description"),
			"Description cannot be parsed",
			staticMapping.SetDescription(m.Description.ValueString()),
		)
	}

	if !m.WINSServers.IsNull() {
		var winsServers []string
		diags.Append(m.WINSServers.ElementsAs(ctx, &winsServers, false)...)
		addPathError(
			&diags,
			path.Root("wins_servers"),
			"WINS servers cannot be parsed",
			staticMapping.SetWINSServers(winsServers),
		)
	}

	if !m.DNSServers.IsNull() {
		var dnsServers []string
		diags.Append(m.DNSServers.ElementsAs(ctx, &dnsServers, false)...)
		addPathError(
			&diags,
			path.Root("dns_servers"),
			"DNS servers cannot be parsed",
			staticMapping.SetDNSServers(dnsServers),
		)
	}

	if !m.Gateway.IsNull() {
		addPathError(
			&diags,
			path.Root("gateway"),
			"Gateway cannot be parsed",
			staticMapping.SetGateway(m.Gateway.ValueString()),
		)
	}

	if !m.DomainName.IsNull() {
		addPathError(
			&diags,
			path.Root("domain_name"),
			"Domain name cannot be parsed",
			staticMapping.SetDomainName(m.DomainName.ValueString()),
		)
	}

	if !m.DomainSearchList.IsNull() {
		var domainSearchList []string
		diags.Append(m.DomainSearchList.ElementsAs(ctx, &domainSearchList, false)...)
		addPathError(
			&diags,
			path.Root("domain_search_list"),
			"Domain search list cannot be parsed",
			staticMapping.SetDomainSearchList(domainSearchList),
		)
	}

	if !m.DefaultLeaseTime.IsNull() {
		addPathError(
			&diags,
			path.Root("default_lease_time"),
			"Default lease time cannot be parsed",
			staticMapping.SetDefaultLeaseTime(m.DefaultLeaseTime.ValueString()),
		)
	}

	if !m.MaximumLeaseTime.IsNull() {
		addPathError(
			&diags,
			path.Root("maximum_lease_time"),
			"Maximum lease time cannot be parsed",
			staticMapping.SetMaximumLeaseTime(m.MaximumLeaseTime.ValueString()),
		)
	}

	return diags
}
