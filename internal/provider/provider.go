// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &OneProviderProvider{}

type OneProviderProvider struct {
	version string
}

type OneProviderProviderModel struct {
	Endpoint  types.String `tfsdk:"endpoint"`
	ApiKey    types.String `tfsdk:"api_key"`
	ClientKey types.String `tfsdk:"client_key"`
}

type OneProviderClient struct {
	httpClient *http.Client
	Endpoint   string
	ApiKey     string
	ClientKey  string
}

func (p *OneProviderProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oneprovider"
	resp.Version = p.version
}

func (p *OneProviderProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The OneProvider API endpoint. Defaults to https://api.oneprovider.com",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The OneProvider API key.",
				Required:            true,
				Sensitive:           true,
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: "The OneProvider client key.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *OneProviderProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OneProviderProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := "https://api.oneprovider.com"
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	apiKey := data.ApiKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("ONEPROVIDER_API_KEY")
	}

	clientKey := data.ClientKey.ValueString()
	if clientKey == "" {
		clientKey = os.Getenv("ONEPROVIDER_CLIENT_KEY")
	}

	if apiKey == "" || clientKey == "" {
		resp.Diagnostics.AddError("Missing credentials", "API key and client key are required. Set ONEPROVIDER_API_KEY and ONEPROVIDER_CLIENT_KEY environment variables or configure in provider block.")
		return
	}

	client := &OneProviderClient{
		httpClient: &http.Client{},
		Endpoint:   endpoint,
		ApiKey:     apiKey,
		ClientKey:  clientKey,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OneProviderProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewVmResource,
		NewImageResource,
		NewSSHKeyResource,
		NewRdnsResource,
	}
}

func (p *OneProviderProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
		NewConfigurationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OneProviderProvider{
			version: version,
		}
	}
}
