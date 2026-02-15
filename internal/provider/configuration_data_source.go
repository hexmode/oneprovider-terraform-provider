// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ConfigurationDataSource{}

type ConfigurationDataSource struct {
	client *OneProviderClient
}

func NewConfigurationDataSource() datasource.DataSource {
	return &ConfigurationDataSource{}
}

func (d *ConfigurationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oneprovider_configuration"
}

func (d *ConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider configuration data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Config ID (optional)",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Configuration type (configs, sizes, templates, locations)",
				Optional:            true,
			},
			"configs": schema.ListAttribute{
				MarkdownDescription: "List of available configurations",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"sizes": schema.ListAttribute{
				MarkdownDescription: "List of available sizes",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"templates": schema.ListAttribute{
				MarkdownDescription: "List of available templates",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"locations": schema.ListAttribute{
				MarkdownDescription: "List of available locations",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *ConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OneProviderClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("expected *OneProviderClient, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigurationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var configs, sizes, templates, locations []string

	result, err := callAPI(ctx, d.client, "GET", "/store/configs", nil)
	if err == nil {
		if configsData, ok := result["response"].([]interface{}); ok {
			for _, c := range configsData {
				if m, ok := c.(map[string]interface{}); ok {
					configs = append(configs, fmt.Sprintf("%v", m))
				}
			}
		}
	}

	result, err = callAPI(ctx, d.client, "GET", "/vm/sizes", nil)
	if err == nil {
		if sizesData, ok := result["response"].([]interface{}); ok {
			for _, s := range sizesData {
				sizes = append(sizes, fmt.Sprintf("%v", s))
			}
		}
	}

	result, err = callAPI(ctx, d.client, "GET", "/vm/templates", nil)
	if err == nil {
		if tmplData, ok := result["response"].([]interface{}); ok {
			for _, t := range tmplData {
				templates = append(templates, fmt.Sprintf("%v", t))
			}
		}
	}

	result, err = callAPI(ctx, d.client, "GET", "/vm/locations", nil)
	if err == nil {
		if locData, ok := result["response"].([]interface{}); ok {
			for _, l := range locData {
				locations = append(locations, fmt.Sprintf("%v", l))
			}
		}
	}

	data.Configs = types.ListValueMust(types.StringType, stringSliceToTypeList(configs))
	data.Sizes = types.ListValueMust(types.StringType, stringSliceToTypeList(sizes))
	data.Templates = types.ListValueMust(types.StringType, stringSliceToTypeList(templates))
	data.Locations = types.ListValueMust(types.StringType, stringSliceToTypeList(locations))

	// Set a fixed ID since this is a singleton data source
	data.ID = types.StringValue("oneprovider_configuration")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func stringSliceToTypeList(s []string) []attr.Value {
	result := make([]attr.Value, len(s))
	for i, v := range s {
		result[i] = types.StringValue(v)
	}
	return result
}

type ConfigurationModel struct {
	ID        types.String `tfsdk:"id"`
	Type      types.String `tfsdk:"type"`
	Configs   types.List   `tfsdk:"configs"`
	Sizes     types.List   `tfsdk:"sizes"`
	Templates types.List   `tfsdk:"templates"`
	Locations types.List   `tfsdk:"locations"`
}
