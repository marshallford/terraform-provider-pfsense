package pfsense

const (
	PrivDiagnosticsCommand            = "WebCfg - Diagnostics: Command"
	PrivDiagnosticsEditFile           = "WebCfg - Diagnostics: Edit File"
	PrivDNSResolver                   = "WebCfg - Services: DNS Resolver"
	PrivDNSResolverEditHost           = "WebCfg - Services: DNS Resolver: Edit host"
	PrivDNSResolverEditDomainOverride = "WebCfg - Services: DNS Resolver: Edit Domain Override"
	PrivDHCPServer                    = "WebCfg - Services: DHCP Server"
	PrivDHCPServerEditStaticMapping   = "WebCfg - Services: DHCP Server: Edit static mapping"
	PrivFirewallAliases               = "WebCfg - Firewall: Aliases"
	PrivFirewallAliasEdit             = "WebCfg - Firewall: Alias: Edit"
	PrivFilterReloadStatus            = "WebCfg - Status: Filter Reload Status"
	PrivPackageManagerInstall         = "WebCfg - System: Package Manager: Install Package"
)

type Privileges struct {
	Create []string
	Read   []string
	Update []string
	Delete []string
}
