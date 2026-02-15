// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerDataSource{}

type ServerDataSource struct {
	client *OneProviderClient
}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oneprovider_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider server data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server ID to look up",
				Required:            true,
			},
			"ip_addr": schema.StringAttribute{
				MarkdownDescription: "Server IP address",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Server hostname",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Server status",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Server username",
				Computed:            true,
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Server location",
				Computed:            true,
			},
			"billing_cycle": schema.StringAttribute{
				MarkdownDescription: "Billing cycle",
				Computed:            true,
			},
			"recurring_amount": schema.StringAttribute{
				MarkdownDescription: "Recurring amount",
				Computed:            true,
			},
		},
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ID.ValueString()
	if serverID == "" {
		resp.Diagnostics.AddError("Missing server ID", "Server ID is required.")
		return
	}

	result, err := callAPI(ctx, d.client, "GET", fmt.Sprintf("/server/info/%s", serverID), nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get server info: %v", err))
		return
	}

	serverData, ok := result["response"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid response format from server info endpoint")
		return
	}

	// Use safe type assertions with existence checks
	if v, ok := serverData["server_id"].(string); ok {
		data.ID = types.StringValue(v)
	}

	if v, ok := serverData["ip_addr"].(string); ok {
		data.IpAddr = types.StringValue(v)
	}

	if v, ok := serverData["hostname"].(string); ok {
		data.Hostname = types.StringValue(v)
	}

	if v, ok := serverData["status"].(string); ok {
		data.Status = types.StringValue(v)
	}

	if v, ok := serverData["username"].(string); ok {
		data.Username = types.StringValue(v)
	}

	if v, ok := serverData["location"].(string); ok {
		data.Location = types.StringValue(v)
	}

	if v, ok := serverData["billing_cycle"].(string); ok {
		data.BillingCycle = types.StringValue(v)
	}

	if v, ok := serverData["recurring_amount"]; ok {
		data.RecurringAmount = types.StringValue(fmt.Sprintf("%v", v))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
