package pfsense

import (
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const MAC48Length = 6

var dnsLabelRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

// TODO implement validation in Set functions
// solve: len == 0 checks will break Get functions, for example when a string is empty

// used for hostname and host override name
func ValidateDNSLabel(dnsLabel string) error {
	if !dnsLabelRegex.MatchString(dnsLabel) {
		return fmt.Errorf("%w, not a valid rfc 1123 dns label", ErrClientValidation)
	}

	return nil
}

// used for FQDN, domain search list, domains, etc
// note: this validation is fairly loose to align with pfSense
func ValidateDomain(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("%w, domain cannot be empty", ErrClientValidation)
	}

	if strings.HasPrefix(domain, ".") {
		return fmt.Errorf("%w, domain cannot start with a dot", ErrClientValidation)
	}

	if strings.Contains(domain, "..") {
		return fmt.Errorf("%w, domain cannot contain consecutive dots", ErrClientValidation)
	}

	dnsLabels := removeEmptyStrings(strings.Split(domain, "."))

	if len(dnsLabels) == 0 {
		// TODO unsure if this can be reached
		return fmt.Errorf("%w, domain cannot be empty", ErrClientValidation)
	}

	for _, dnsLabel := range dnsLabels {
		if !dnsLabelRegex.MatchString(dnsLabel) {
			return fmt.Errorf("%w, domain must contain valid rfc 1123 dns label(s)", ErrClientValidation)
		}
	}

	return nil
}

func ValidateAlias(alias string) error {
	if len(alias) == 0 {
		return fmt.Errorf("%w, alias cannot be empty", ErrClientValidation)
	}

	for _, character := range alias {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '_' {
			return fmt.Errorf("%w, alias can only contain letters (A-Z, a-z), numbers (0-9), and underscores (_)", ErrClientValidation)
		}
	}

	return nil
}

func ValidateConfigFileName(configFileName string) error {
	if len(configFileName) == 0 {
		return fmt.Errorf("%w, config file name cannot be empty", ErrClientValidation)
	}

	if strings.HasPrefix(configFileName, "-") || strings.HasSuffix(configFileName, "-") {
		return fmt.Errorf("%w, config file name cannot start or end with a dash", ErrClientValidation)
	}

	for _, character := range configFileName {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '-' {
			return fmt.Errorf("%w, config file name can only contain letters (A-Z, a-z), numbers (0-9), and dashes (-)", ErrClientValidation)
		}
	}

	return nil
}

func ValidateInterface(iface string) error {
	if len(iface) == 0 {
		return fmt.Errorf("%w, interface cannot be empty", ErrClientValidation)
	}

	if !unicode.IsLetter(rune(iface[0])) {
		return fmt.Errorf("%w, interface must start with a letter (A-Z, a-z)", ErrClientValidation)
	}

	for _, character := range iface {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return fmt.Errorf("%w, interface can only contain letters and numbers (A-Z, a-z, 0-9)", ErrClientValidation)
		}
	}

	return nil
}

func ValidateMACAddress(macAddress string) error {
	if len(macAddress) == 0 {
		return fmt.Errorf("%w, mac address cannot be empty", ErrClientValidation)
	}

	mac, err := net.ParseMAC(macAddress)
	if err != nil {
		return fmt.Errorf("%w, not a valid mac address", ErrClientValidation)
	}

	if len(mac) != MAC48Length {
		return fmt.Errorf("%w, not an ieee 802 mac-48 address (6 bytes)", ErrClientValidation)
	}

	if !strings.Contains(macAddress, ":") {
		return fmt.Errorf("%w, hex octets must be separated by colons", ErrClientValidation)
	}

	return nil
}

func ValidatePort(port string) error {
	if len(port) == 0 {
		return fmt.Errorf("%w, port cannot be empty", ErrClientValidation)
	}

	numericPort, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("%w, port must be a numeric string", ErrClientValidation)
	}

	if numericPort < 1 || numericPort > 65535 {
		return fmt.Errorf("%w, port must be in the range 1-65535", ErrClientValidation)
	}

	return nil
}

func ValidatePortRange(portRange string) error {
	if len(portRange) == 0 {
		return fmt.Errorf("%w, port range cannot be empty", ErrClientValidation)
	}

	ports := strings.Split(portRange, ":")

	if len(ports) != 2 { //nolint:mnd
		return fmt.Errorf("%w, port range must be in the format 'startPort:endPort'", ErrClientValidation)
	}

	startPort, startPortErr := strconv.Atoi(ports[0])
	endPort, endPortErr := strconv.Atoi(ports[1])

	if startPortErr != nil || endPortErr != nil {
		return fmt.Errorf("%w, both ports must be a numeric string", ErrClientValidation)
	}

	if startPort < 1 || startPort > 65535 || endPort < 1 || endPort > 65535 {
		return fmt.Errorf("%w, both ports must be in the range 1-65535", ErrClientValidation)
	}

	// FYI pfSense does not require startPort is <= endPort

	return nil
}

func ValidateIPAddress(addr string, addrFamily string) error {
	if len(addr) == 0 {
		return fmt.Errorf("%w, ip address cannot be empty", ErrClientValidation)
	}

	parsedAddr, err := netip.ParseAddr(addr)

	if err != nil || !parsedAddr.IsValid() {
		return fmt.Errorf("%w, not a valid ip address", ErrClientValidation)
	}

	switch addrFamily {
	case "IPv4":
		if !parsedAddr.Is4() {
			return fmt.Errorf("%w, not a valid ipv4 address", ErrClientValidation)
		}
	case "IPv6":
		if !parsedAddr.Is6() {
			return fmt.Errorf("%w, not a valid ipv6 address", ErrClientValidation)
		}
	}

	return nil
}

func ValidateIPAddressPort(addrPort string) error {
	if len(addrPort) == 0 {
		return fmt.Errorf("%w, address and port cannot be empty", ErrClientValidation)
	}

	parsedAddrPort, err := netip.ParseAddrPort(addrPort)

	if err != nil || !parsedAddrPort.IsValid() {
		return fmt.Errorf("%w, not a valid address and port", ErrClientValidation)
	}

	if parsedAddrPort.Port() < 1 {
		return fmt.Errorf("%w, port must be in the range 1-65535", ErrClientValidation)
	}

	return nil
}

func ValidateNetwork(network string) error {
	if len(network) == 0 {
		return fmt.Errorf("%w, network cannot be empty", ErrClientValidation)
	}

	parsedNetwork, err := netip.ParsePrefix(network)

	if err != nil || !parsedNetwork.IsValid() {
		return fmt.Errorf("%w, not a valid network", ErrClientValidation)
	}

	return nil
}
