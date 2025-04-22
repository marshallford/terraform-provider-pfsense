package pfsense

import (
	"fmt"
	"net"
)

func ParseMACAddress(macAddress string) (net.HardwareAddr, error) {
	hwAddr, err := net.ParseMAC(macAddress)
	if err != nil {
		return nil, fmt.Errorf("%w, not a valid mac address", ErrClientValidation)
	}

	return hwAddr, nil
}

func CompareMACAddresses(macAddress1 net.HardwareAddr, macAddress2 net.HardwareAddr) bool {
	return macAddress1.String() == macAddress2.String()
}
