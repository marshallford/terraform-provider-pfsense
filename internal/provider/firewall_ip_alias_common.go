package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type FirewallIPAliasModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Entries     types.List   `tfsdk:"entries"`
}

type FirewallIPAliasEntryModel struct {
	IP          types.String `tfsdk:"ip"`
	Description types.String `tfsdk:"description"`
}

func (FirewallIPAliasModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"name": {
			Description: "Name of IP alias.",
		},
		"description": {
			Description: descriptionDescription,
		},
		"type": {
			Description:         fmt.Sprintf("Type of alias. Options: %s.", wrapElementsJoin(pfsense.FirewallIPAlias{}.Types(), "'")),
			MarkdownDescription: fmt.Sprintf("Type of alias. Options: %s.", wrapElementsJoin(pfsense.FirewallIPAlias{}.Types(), "`")),
		},
		"entries": {
			Description: "Host(s) or network(s).",
		},
	}
}

func (FirewallIPAliasEntryModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"ip": {
			Description: "Hosts must be specified by their IP address or fully qualified domain name (FQDN). Networks are specified in CIDR format.",
		},
		"description": {
			Description: descriptionDescription,
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
		"ip":          types.StringType,
		"description": types.StringType,
	}
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

	ipAliasEntriesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallIPAliasEntryModel{}.AttrTypes()}, ipAliasEntryModels)
	diags.Append(newDiags...)
	m.Entries = ipAliasEntriesValue

	return diags
}

func (m *FirewallIPAliasEntryModel) Set(_ context.Context, ipAliasEntry pfsense.FirewallIPAliasEntry) diag.Diagnostics {
	var diags diag.Diagnostics

	m.IP = types.StringValue(ipAliasEntry.IP)

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
		attrPath.AtName("ip"),
		"Entry ip cannot be parsed",
		ipAliasEntry.SetIP(m.IP.ValueString()),
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
