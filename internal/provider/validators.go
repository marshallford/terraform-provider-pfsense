package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type stringIsDNSLabelValidator struct{}

func (v stringIsDNSLabelValidator) Description(_ context.Context) string {
	return "string must be a RFC 1123 DNS label"
}

func (v stringIsDNSLabelValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsDNSLabelValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateDNSLabel(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid RFC 1123 DNS label", err)
}

func stringIsDNSLabel() stringIsDNSLabelValidator {
	return stringIsDNSLabelValidator{}
}

type stringIsDomainValidator struct{}

func (v stringIsDomainValidator) Description(_ context.Context) string {
	return "string must be a domain" // TODO improve
}

func (v stringIsDomainValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsDomainValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateDomain(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid domain", err) // TODO improve
}

func stringIsDomain() stringIsDomainValidator {
	return stringIsDomainValidator{}
}

type stringIsAliasValidator struct{}

func (v stringIsAliasValidator) Description(_ context.Context) string {
	return "string must be a pfsense alias"
}

func (v stringIsAliasValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsAliasValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateAlias(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid pfsense alias", err)
}

func stringIsAlias() stringIsAliasValidator {
	return stringIsAliasValidator{}
}

type stringIsConfigFileNameValidator struct{}

func (v stringIsConfigFileNameValidator) Description(_ context.Context) string {
	return "string must be a config file name"
}

func (v stringIsConfigFileNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsConfigFileNameValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateConfigFileName(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid config file name", err)
}

func stringIsConfigFileName() stringIsConfigFileNameValidator {
	return stringIsConfigFileNameValidator{}
}

type stringIsInterfaceValidator struct{}

func (v stringIsInterfaceValidator) Description(_ context.Context) string {
	return "string must be a pfsense interface"
}

func (v stringIsInterfaceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsInterfaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateInterface(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid pfsense interface", err)
}

func stringIsInterface() stringIsInterfaceValidator {
	return stringIsInterfaceValidator{}
}

type stringIsMACAddressValidator struct{}

func (v stringIsMACAddressValidator) Description(_ context.Context) string {
	return "string must be a MAC address"
}

func (v stringIsMACAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsMACAddressValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateMACAddress(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid MAC address", err)
}

func stringIsMACAddress() stringIsMACAddressValidator {
	return stringIsMACAddressValidator{}
}

type stringIsPortValidator struct{}

func (v stringIsPortValidator) Description(_ context.Context) string {
	return "string must be a port number"
}

func (v stringIsPortValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsPortValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidatePort(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid port number", err)
}

func stringIsPort() stringIsPortValidator {
	return stringIsPortValidator{}
}

type stringIsPortRangeValidator struct{}

func (v stringIsPortRangeValidator) Description(_ context.Context) string {
	return "string must be a port range"
}

func (v stringIsPortRangeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsPortRangeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidatePortRange(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid port range", err)
}

func stringIsPortRange() stringIsPortRangeValidator {
	return stringIsPortRangeValidator{}
}

type stringIsIPAddressValidator struct {
	AddressFamily string
}

func (v stringIsIPAddressValidator) Description(_ context.Context) string {
	if v.AddressFamily == "any" {
		return "string must be an ip address"
	}

	return fmt.Sprintf("string must be an %s address", v.AddressFamily)
}

func (v stringIsIPAddressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIPAddressValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateIPAddress(req.ConfigValue.ValueString(), v.AddressFamily)
	summary := "Not a valid ip address"

	if v.AddressFamily != "any" {
		summary = fmt.Sprintf("Not a valid %s address", v.AddressFamily)
	}

	addPathError(&resp.Diagnostics, req.Path, summary, err)
}

func stringIsIPAddress(addrFamily string) stringIsIPAddressValidator {
	return stringIsIPAddressValidator{
		AddressFamily: addrFamily,
	}
}

type stringIsIPAddressPortValidator struct{}

func (v stringIsIPAddressPortValidator) Description(_ context.Context) string {
	return "string must be an ip address port"
}

func (v stringIsIPAddressPortValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsIPAddressPortValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateIPAddressPort(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid ip address and port", err)
}

func stringIsIPAddressPort() stringIsIPAddressPortValidator {
	return stringIsIPAddressPortValidator{}
}

type stringIsNetworkValidator struct{}

func (v stringIsNetworkValidator) Description(_ context.Context) string {
	return "string must be a network"
}

func (v stringIsNetworkValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsNetworkValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	err := pfsense.ValidateNetwork(req.ConfigValue.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid network", err)
}

func stringIsNetwork() stringIsNetworkValidator {
	return stringIsNetworkValidator{}
}
