package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type FirewallAliasesModel struct {
	IP types.List `tfsdk:"ip"`
}

type FirewallIPAliasModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Apply       types.Bool   `tfsdk:"apply"`
	Entries     types.List   `tfsdk:"entries"`
}

type FirewallIPAliasEntryModel struct {
	Address     types.String `tfsdk:"address"`
	Description types.String `tfsdk:"description"`
}

func (FirewallIPAliasModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"name": {
			Description: "Name of alias.",
		},
		"description": {
			Description: "For administrative reference (not parsed).",
		},
		"type": {
			Description: "Type of alias.",
		},
		"apply": {
			Description:         "Apply change, defaults to 'true'.",
			MarkdownDescription: "Apply change, defaults to `true`.",
		},
		"entries": {
			Description: "Host(s) or network(s).",
		},
	}
}

func (FirewallIPAliasEntryModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"address": {
			Description: "Hosts must be specified by their IP address or fully qualified domain name (FQDN). Networks are specified in CIDR format.",
		},
		"description": {
			Description: "For administrative reference (not parsed).",
		},
	}
}

func (FirewallIPAliasModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"type":        types.StringType,
		"entries":     types.ListType{ElemType: types.ObjectType{AttrTypes: FirewallIPAliasEntryModel{}.AttrTypes()}},
	}
}

func (FirewallIPAliasEntryModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address":     types.StringType,
		"description": types.StringType,
	}
}

func (m *FirewallAliasesModel) Set(ctx context.Context, ipAliases pfsense.FirewallIPAliases) diag.Diagnostics {
	var diags diag.Diagnostics

	ipAliasModels := []FirewallIPAliasModel{}
	for _, ipAlias := range ipAliases {
		var ipAliasModel FirewallIPAliasModel
		diags.Append(ipAliasModel.Set(ctx, ipAlias)...)
		ipAliasModels = append(ipAliasModels, ipAliasModel)
	}

	ipAliasesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallIPAliasModel{}.AttrTypes()}, ipAliasModels)
	diags.Append(newDiags...)
	m.IP = ipAliasesValue

	return diags
}

func (m *FirewallIPAliasModel) Set(ctx context.Context, ipAlias pfsense.FirewallIPAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(ipAlias.Name)

	if ipAlias.Description != "" {
		m.Description = types.StringValue(ipAlias.Description)
	}

	m.Type = types.StringValue(ipAlias.Type)

	ipAliasEntryModels := []FirewallIPAliasEntryModel{}
	for _, ipAlias := range ipAlias.Entries {
		var ipAliasEntryModel FirewallIPAliasEntryModel
		diags.Append(ipAliasEntryModel.Set(ctx, ipAlias)...)
		ipAliasEntryModels = append(ipAliasEntryModels, ipAliasEntryModel)
	}

	ipAliasesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallIPAliasModel{}.AttrTypes()}, ipAliasEntryModels)
	diags.Append(newDiags...)
	m.Entries = ipAliasesValue

	return diags
}

func (m *FirewallIPAliasEntryModel) Set(_ context.Context, ipAliasEntry pfsense.FirewallIPAliasEntry) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Address = types.StringValue(ipAliasEntry.Address)

	if ipAliasEntry.Description != "" {
		m.Description = types.StringValue(ipAliasEntry.Description)
	}

	return diags
}

func (m FirewallIPAliasModel) Value(ctx context.Context, ipAlias *pfsense.FirewallIPAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		path.Root("name"),
		"Name cannot be parsed",
		ipAlias.SetName(m.Name.ValueString()),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			path.Root("description"),
			"Description cannot be parsed",
			ipAlias.SetDescription(m.Description.ValueString()),
		)
	}

	addPathError(
		&diags,
		path.Root("type"),
		"Type cannot be parsed",
		ipAlias.SetType(m.Type.ValueString()),
	)

	var ipAliasEntryModels []FirewallIPAliasEntryModel
	if !m.Entries.IsNull() {
		diags.Append(m.Entries.ElementsAs(ctx, &ipAliasEntryModels, false)...)
	}

	ipAlias.Entries = make([]pfsense.FirewallIPAliasEntry, 0, len(ipAliasEntryModels))
	for index, ipAliasEntryModel := range ipAliasEntryModels {
		var ipAliasEntry pfsense.FirewallIPAliasEntry

		diags.Append(ipAliasEntryModel.Value(ctx, &ipAliasEntry, path.Root("entries").AtListIndex(index))...)
		ipAlias.Entries = append(ipAlias.Entries, ipAliasEntry)
	}

	return diags
}

func (m FirewallIPAliasEntryModel) Value(_ context.Context, ipAliasEntry *pfsense.FirewallIPAliasEntry, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		attrPath.AtName("address"),
		"Entry address cannot be parsed",
		ipAliasEntry.SetAddress(m.Address.ValueString()),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			attrPath.AtName("description"),
			"Entry description cannot be parsed",
			ipAliasEntry.SetDescription(m.Description.ValueString()),
		)
	}

	return diags
}
