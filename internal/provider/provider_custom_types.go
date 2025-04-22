package provider

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_             basetypes.StringTypable                    = (*macAddressType)(nil)
	_             basetypes.StringValuable                   = (*macAddressValue)(nil)
	_             basetypes.StringValuableWithSemanticEquals = (*macAddressValue)(nil)
	_             xattr.ValidateableAttribute                = (*macAddressValue)(nil)
	errCustomType                                            = errors.New("custom type")
)

type macAddressType struct {
	basetypes.StringType
}

func (t macAddressType) String() string {
	return "macAddressType"
}

func (t macAddressType) ValueType(ctx context.Context) attr.Value { //nolint:ireturn
	return macAddressValue{}
}

func (t macAddressType) Equal(o attr.Type) bool {
	other, ok := o.(macAddressType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t macAddressType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) { //nolint:ireturn
	return macAddressValue{
		StringValue: in,
	}, nil
}

func (t macAddressType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) { //nolint:ireturn
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("mac address %w: unexpected value type of %T", errCustomType, attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("mac address %w: unexpected error converting StringValue to StringValuable: %v", errCustomType, diags)
	}

	return stringValuable, nil
}

type macAddressValue struct {
	basetypes.StringValue
}

func (v macAddressValue) Type(_ context.Context) attr.Type { //nolint:ireturn
	return macAddressType{}
}

func (v macAddressValue) Equal(o attr.Value) bool {
	other, ok := o.(macAddressValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v macAddressValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(macAddressValue)
	if !ok {
		summary, detail := unexpectedValueTypeSemanticEquality(v, newValuable)
		diags.AddError(summary, detail)

		return false, diags
	}

	newValueHwAddr, err := pfsense.ParseMACAddress(v.ValueString())
	if err != nil {
		summary, detail := unexpectedErrorSemanticEquality(err)
		diags.AddError(summary, detail)
	}

	valueHwAddr, err := pfsense.ParseMACAddress(newValue.ValueString())
	if err != nil {
		summary, detail := unexpectedErrorSemanticEquality(err)
		diags.AddError(summary, detail)
	}

	if diags.HasError() {
		return false, diags
	}

	return pfsense.CompareMACAddresses(newValueHwAddr, valueHwAddr), diags
}

func (v macAddressValue) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	err := pfsense.ValidateMACAddress(v.ValueString())
	addPathError(&resp.Diagnostics, req.Path, "Not a valid MAC address", err)
}

func (v macAddressValue) parseMACAddress() (net.HardwareAddr, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() {
		addError(&diags, "MAC Address Parse Error", fmt.Errorf("mac address %w: value is null", errCustomType))

		return nil, diags
	}

	if v.IsUnknown() {
		addError(&diags, "MAC Address Parse Error", fmt.Errorf("mac address %w: value is unknown", errCustomType))

		return nil, diags
	}

	hwAddr, err := pfsense.ParseMACAddress(v.ValueString())
	if err != nil {
		addError(&diags, "MAC Address Parse Error", err)

		return nil, diags
	}

	return hwAddr, diags
}

// func newMACAddressNull() macAddressValue {
// 	return macAddressValue{
// 		StringValue: basetypes.NewStringNull(),
// 	}
// }

// func newMACAddressUnknown() macAddressValue {
// 	return macAddressValue{
// 		StringValue: basetypes.NewStringUnknown(),
// 	}
// }

func newMACAddressValue(value string) macAddressValue {
	return macAddressValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// func newMACAddressPointerValue(value *string) macAddressValue {
// 	return macAddressValue{
// 		StringValue: basetypes.NewStringPointerValue(value),
// 	}
// }
