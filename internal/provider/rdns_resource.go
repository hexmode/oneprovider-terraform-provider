// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RdnsResource{}

type RdnsResource struct {
	client *OneProviderClient
}

func NewRdnsResource() resource.Resource {
	return &RdnsResource{}
}

func (r *RdnsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oneprovider_rdns"
}

func (r *RdnsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider reverse DNS resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The IP address",
				Required:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The IP address",
				Required:            true,
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The reverse DNS hostname",
				Optional:            true,
				Computed:            true,
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "VM ID (required for create)",
				Optional:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live in seconds",
				Computed:            true,
			},
		},
	}
}

func (r *RdnsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RdnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RdnsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAddress := plan.IpAddress.ValueString()
	domain := plan.Domain.ValueString()

	if ipAddress == "" {
		resp.Diagnostics.AddError("Missing IP Address", "ip_address is required")
		return
	}

	if domain == "" {
		resp.Diagnostics.AddError("Missing Domain", "domain is required")
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/vm/rdns/add", map[string]interface{}{
		"ip":     ipAddress,
		"domain": domain,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to add reverse DNS: %v", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RdnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RdnsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAddress := state.IpAddress.ValueString()
	if ipAddress == "" {
		resp.Diagnostics.AddError("Missing IP Address", "ip_address is required")
		return
	}

	result, err := callAPI(ctx, r.client, "GET", "/vm/rdns/get", map[string]interface{}{
		"ip": ipAddress,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get reverse DNS: %v", err))
		return
	}

	record, ok := result["response"].(map[string]interface{})["record"].(map[string]interface{})
	if !ok || record == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if content, ok := record["content"].(string); ok {
		state.Domain = types.StringValue(content)
	}

	if ttl, ok := record["ttl"].(string); ok {
		var ttlInt int64
		fmt.Sscanf(ttl, "%d", &ttlInt)
		state.Ttl = types.Int64Value(ttlInt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RdnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state RdnsModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAddress := state.IpAddress.ValueString()
	domain := plan.Domain.ValueString()

	if domain == "" {
		_, err := callAPI(ctx, r.client, "POST", "/vm/rdns/delete", map[string]interface{}{
			"ip": ipAddress,
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete reverse DNS: %v", err))
			return
		}
	} else {
		_, err := callAPI(ctx, r.client, "POST", "/vm/rdns/edit", map[string]interface{}{
			"ip":     ipAddress,
			"domain": domain,
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update reverse DNS: %v", err))
			return
		}
		state.Domain = plan.Domain
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RdnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RdnsModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ipAddress := state.IpAddress.ValueString()
	if ipAddress == "" {
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/vm/rdns/delete", map[string]interface{}{
		"ip": ipAddress,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete reverse DNS: %v", err))
		return
	}
}

type RdnsModel struct {
	ID        types.String `tfsdk:"id"`
	IpAddress types.String `tfsdk:"ip_address"`
	Domain    types.String `tfsdk:"domain"`
	VmID      types.String `tfsdk:"vm_id"`
	Ttl       types.Int64  `tfsdk:"ttl"`
}
