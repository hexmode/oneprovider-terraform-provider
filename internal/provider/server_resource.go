// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ServerResource{}

type ServerResource struct {
	client *OneProviderClient
}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oneprovider_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider server resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server ID",
				Computed:            true,
			},
			"ip_addr": schema.StringAttribute{
				MarkdownDescription: "Server IP address",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Server hostname",
				Optional:            true,
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

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*OneProviderClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("expected *OneProviderClient, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Server resource create not implemented - API is read-only for servers")
	resp.Diagnostics.AddError("Not Implemented", "Server creation is not supported. Use the OneProvider panel to provision servers.")
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ID.ValueString()
	if serverID == "" {
		resp.Diagnostics.AddError("Missing server ID", "Server ID is required to read server data.")
		return
	}

	result, err := callAPI(ctx, r.client, "GET", fmt.Sprintf("/server/info/%s", serverID), nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get server info: %v", err))
		return
	}

	serverData, ok := result["response"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid response format from server info endpoint")
		return
	}

	data.ID = types.StringValue(serverData["server_id"].(string))
	data.IpAddr = types.StringValue(serverData["ip_addr"].(string))
	data.Hostname = types.StringValue(serverData["hostname"].(string))
	data.Status = types.StringValue(serverData["status"].(string))
	data.Username = types.StringValue(serverData["username"].(string))
	data.Location = types.StringValue(serverData["location"].(string))
	data.BillingCycle = types.StringValue(serverData["billing_cycle"].(string))
	data.RecurringAmount = types.StringValue(fmt.Sprintf("%v", serverData["recurring_amount"]))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ID.ValueString()
	if data.Hostname.ValueString() != "" {
		_, err := callAPI(ctx, r.client, "POST", "/server/hostname", map[string]interface{}{
			"server_id": serverID,
			"hostname":  data.Hostname.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update hostname: %v", err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/server/cancel", map[string]interface{}{
		"server_id": data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to cancel server: %v", err))
		return
	}
}

type ServerModel struct {
	ID              types.String `tfsdk:"id"`
	IpAddr          types.String `tfsdk:"ip_addr"`
	Hostname        types.String `tfsdk:"hostname"`
	Status          types.String `tfsdk:"status"`
	Username        types.String `tfsdk:"username"`
	Location        types.String `tfsdk:"location"`
	BillingCycle    types.String `tfsdk:"billing_cycle"`
	RecurringAmount types.String `tfsdk:"recurring_amount"`
}

func callAPI(ctx context.Context, client *OneProviderClient, method, endpoint string, params map[string]interface{}) (map[string]interface{}, error) {
	url := client.Endpoint + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Api-Key", client.ApiKey)
	req.Header.Set("Client-Key", client.ClientKey)
	req.Header.Set("User-Agent", "OneProvider-Terraform/1.0")

	if params != nil && (method == "POST" || method == "PUT") {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	tflog.Debug(ctx, fmt.Sprintf("API Request: %s %s", method, url))

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result["result"] != "success" {
		return nil, fmt.Errorf("API returned error: %v", result)
	}

	return result, nil
}
