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

	data.ID = types.StringValue(serverData["server_id"].(string))               // nolint: forcetypeassert
	data.IpAddr = types.StringValue(serverData["ip_addr"].(string))             // nolint: forcetypeassert
	data.Hostname = types.StringValue(serverData["hostname"].(string))          // nolint: forcetypeassert
	data.Status = types.StringValue(serverData["status"].(string))              // nolint: forcetypeassert
	data.Username = types.StringValue(serverData["username"].(string))          // nolint: forcetypeassert
	data.Location = types.StringValue(serverData["location"].(string))          // nolint: forcetypeassert
	data.BillingCycle = types.StringValue(serverData["billing_cycle"].(string)) // nolint: forcetypeassert
	data.RecurringAmount = types.StringValue(fmt.Sprintf("%v", serverData["recurring_amount"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
