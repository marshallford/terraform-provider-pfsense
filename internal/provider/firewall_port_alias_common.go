package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

type FirewallPortAliasModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Entries     types.List   `tfsdk:"entries"`
}

type FirewallPortAliasEntryModel struct {
	Port        types.String `tfsdk:"port"`
	Description types.String `tfsdk:"description"`
}

func (FirewallPortAliasModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"name": {
			Description: "Name of port alias.",
		},
		"description": {
			Description: descriptionDescription,
		},
		"entries": {
			Description: "Port(s) or port range(s).",
		},
	}
}

func (FirewallPortAliasEntryModel) descriptions() map[string]attrDescription {
	return map[string]attrDescription{
		"port": {
			Description: "A single port or port range. Port ranges can be expressed by separating with a colon.",
		},
		"description": {
			Description: descriptionDescription,
		},
	}
}

func (FirewallPortAliasModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"entries":     types.ListType{ElemType: types.ObjectType{AttrTypes: FirewallPortAliasEntryModel{}.AttrTypes()}},
	}
}

func (FirewallPortAliasEntryModel) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port":        types.StringType,
		"description": types.StringType,
	}
}

func (m *FirewallPortAliasModel) Set(ctx context.Context, portAlias pfsense.FirewallPortAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(portAlias.Name)

	if portAlias.Description != "" {
		m.Description = types.StringValue(portAlias.Description)
	}

	portAliasEntryModels := []FirewallPortAliasEntryModel{}
	for _, portAlias := range portAlias.Entries {
		var portAliasEntryModel FirewallPortAliasEntryModel
		diags.Append(portAliasEntryModel.Set(ctx, portAlias)...)
		portAliasEntryModels = append(portAliasEntryModels, portAliasEntryModel)
	}

	portAliasEntriesValue, newDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: FirewallPortAliasEntryModel{}.AttrTypes()}, portAliasEntryModels)
	diags.Append(newDiags...)
	m.Entries = portAliasEntriesValue

	return diags
}

func (m *FirewallPortAliasEntryModel) Set(_ context.Context, portAliasEntry pfsense.FirewallPortAliasEntry) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Port = types.StringValue(portAliasEntry.Port)

	if portAliasEntry.Description != "" {
		m.Description = types.StringValue(portAliasEntry.Description)
	}

	return diags
}

func (m FirewallPortAliasModel) Value(ctx context.Context, portAlias *pfsense.FirewallPortAlias) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		path.Root("name"),
		"Name cannot be parsed",
		portAlias.SetName(m.Name.ValueString()),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			path.Root("description"),
			"Description cannot be parsed",
			portAlias.SetDescription(m.Description.ValueString()),
		)
	}

	var portAliasEntryModels []FirewallPortAliasEntryModel
	if !m.Entries.IsNull() {
		diags.Append(m.Entries.ElementsAs(ctx, &portAliasEntryModels, false)...)
	}

	portAlias.Entries = make([]pfsense.FirewallPortAliasEntry, 0, len(portAliasEntryModels))
	for index, portAliasEntryModel := range portAliasEntryModels {
		var portAliasEntry pfsense.FirewallPortAliasEntry

		diags.Append(portAliasEntryModel.Value(ctx, &portAliasEntry, path.Root("entries").AtListIndex(index))...)
		portAlias.Entries = append(portAlias.Entries, portAliasEntry)
	}

	return diags
}

func (m FirewallPortAliasEntryModel) Value(_ context.Context, portAliasEntry *pfsense.FirewallPortAliasEntry, attrPath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	addPathError(
		&diags,
		attrPath.AtName("port"),
		"Entry port cannot be parsed",
		portAliasEntry.SetPort(m.Port.ValueString()),
	)

	if !m.Description.IsNull() {
		addPathError(
			&diags,
			attrPath.AtName("description"),
			"Entry description cannot be parsed",
			portAliasEntry.SetDescription(m.Description.ValueString()),
		)
	}

	return diags
}
