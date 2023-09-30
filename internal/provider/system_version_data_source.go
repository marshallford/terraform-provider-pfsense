package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ datasource.DataSource              = &SystemVersionDataSource{}
	_ datasource.DataSourceWithConfigure = &SystemVersionDataSource{}
)

func NewSystemVersionDataSource() datasource.DataSource {
	return &SystemVersionDataSource{}
}

type SystemVersionDataSource struct {
	client *pfsense.Client
}

type SystemVersionDataSourceModel struct {
	Current types.String `tfsdk:"current"`
	Latest  types.String `tfsdk:"latest"`
}

func (d *SystemVersionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_system_version", req.ProviderTypeName)
}

func (d *SystemVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves current and latest system version.",
		Attributes: map[string]schema.Attribute{
			"current": schema.StringAttribute{
				Description: "Current pfSense system version.",
				Computed:    true,
			},
			"latest": schema.StringAttribute{
				Description: "Latest pfSense system version.",
				Computed:    true,
			},
		},
	}
}

func (d *SystemVersionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, ok := configureDataSourceClient(req, resp)
	if !ok {
		return
	}

	d.client = client
}

func (d *SystemVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SystemVersionDataSourceModel

	version, err := d.client.GetSystemVersion(ctx)
	if addError(&resp.Diagnostics, "Unable to get system version", err) {
		return
	}

	data.Current = types.StringValue(version.Current)
	data.Latest = types.StringValue(version.Latest)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
