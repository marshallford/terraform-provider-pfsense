package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/marshallford/terraform-provider-pfsense/pkg/pfsense"
)

var (
	_ resource.Resource                = (*DNSResolverDomainOverrideResource)(nil)
	_ resource.ResourceWithConfigure   = (*DNSResolverDomainOverrideResource)(nil)
	_ resource.ResourceWithImportState = (*DNSResolverDomainOverrideResource)(nil)
)

type DNSResolverDomainOverrideResourceModel struct {
	DNSResolverDomainOverrideModel
	Apply types.Bool `tfsdk:"apply"`
}

func NewDNSResolverDomainOverrideResource() resource.Resource { //nolint:ireturn
	return &DNSResolverDomainOverrideResource{}
}

type DNSResolverDomainOverrideResource struct {
	client *pfsense.Client
}

func (r *DNSResolverDomainOverrideResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_dnsresolver_domainoverride", req.ProviderTypeName)
}

func (r *DNSResolverDomainOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "DNS resolver domain override. Domain for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		MarkdownDescription: "DNS resolver [domain override](https://docs.netgate.com/pfsense/en/latest/services/dns/resolver-domain-overrides.html). Domain for which the resolver's standard DNS lookup process should be overridden and a different (non-standard) lookup server should be queried instead.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: DNSResolverDomainOverrideModel{}.descriptions()["domain"].Description,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringIsDomain(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: DNSResolverDomainOverrideModel{}.descriptions()["ip_address"].Description,
				Required:    true,
				Validators: []validator.String{
					stringIsIPAddressPort(),
				},
			},
			"tls_queries": schema.BoolAttribute{
				Description:         DNSResolverDomainOverrideModel{}.descriptions()["tls_queries"].Description,
				MarkdownDescription: DNSResolverDomainOverrideModel{}.descriptions()["tls_queries"].MarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(defaultDomainOverrideTLSQueries),
			},
			"tls_hostname": schema.StringAttribute{
				Description: DNSResolverDomainOverrideModel{}.descriptions()["tls_hostname"].Description,
				Optional:    true,
				Validators: []validator.String{
					stringIsDomain(),
				},
			},
			"description": schema.StringAttribute{
				Description: DNSResolverDomainOverrideModel{}.descriptions()["description"].Description,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"apply": schema.BoolAttribute{
				Description:         applyDescription,
				MarkdownDescription: applyMarkdownDescription,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(defaultApply),
			},
		},
	}
}

func (r *DNSResolverDomainOverrideResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, ok := configureResourceClient(req, resp)
	if !ok {
		return
	}

	r.client = client
}

func (r *DNSResolverDomainOverrideResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var domainOverrideReq pfsense.DomainOverride
	resp.Diagnostics.Append(data.Value(ctx, &domainOverrideReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.CreateDNSResolverDomainOverride(ctx, domainOverrideReq)
	if addError(&resp.Diagnostics, "Error creating domain override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *domainOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		addWarning(&resp.Diagnostics, "Error applying domain override", err)
	}
}

func (r *DNSResolverDomainOverrideResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.GetDNSResolverDomainOverride(ctx, data.Domain.ValueString())

	if errors.Is(err, pfsense.ErrNotFound) {
		resp.State.RemoveResource(ctx)

		return
	}

	if addError(&resp.Diagnostics, "Error reading domain override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *domainOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSResolverDomainOverrideResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var domainOverrideReq pfsense.DomainOverride
	resp.Diagnostics.Append(data.Value(ctx, &domainOverrideReq)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainOverride, err := r.client.UpdateDNSResolverDomainOverride(ctx, domainOverrideReq)
	if addError(&resp.Diagnostics, "Error updating domain override", err) {
		return
	}

	resp.Diagnostics.Append(data.Set(ctx, *domainOverride)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		addWarning(&resp.Diagnostics, "Error applying domain override", err)
	}
}

func (r *DNSResolverDomainOverrideResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DNSResolverDomainOverrideResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSResolverDomainOverride(ctx, data.Domain.ValueString())
	if addError(&resp.Diagnostics, "Error deleting domain override", err) {
		return
	}

	resp.State.RemoveResource(ctx)

	if data.Apply.ValueBool() {
		err = r.client.ApplyDNSResolverChanges(ctx)
		addWarning(&resp.Diagnostics, "Error applying domain override", err)
	}
}

func (r *DNSResolverDomainOverrideResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("apply"), types.BoolValue(defaultApply))...)
}
