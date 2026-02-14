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

var _ resource.Resource = &SSHKeyResource{}

type SSHKeyResource struct {
	client *OneProviderClient
}

func NewSSHKeyResource() resource.Resource {
	return &SSHKeyResource{}
}

func (r *SSHKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oneprovider_ssh_key"
}

func (r *SSHKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider SSH key resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "SSH key UUID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SSH key name",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "SSH public key value",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *SSHKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SSHKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := callAPI(ctx, r.client, "POST", "/vm/sshkey/new", map[string]interface{}{
		"key_name":  plan.Name.ValueString(),
		"key_value": plan.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create SSH key: %v", err))
		return
	}

	keys, ok := result["response"].(map[string]interface{})["keys"].([]interface{})
	if !ok || len(keys) == 0 {
		resp.Diagnostics.AddError("API Error", "Failed to get created SSH key UUID")
		return
	}

	keyData, ok := keys[0].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid SSH key response format")
		return
	}

	plan.ID = types.StringValue(keyData["uuid"].(string))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SSHKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := callAPI(ctx, r.client, "GET", "/vm/sshkeys/list", nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to list SSH keys: %v", err))
		return
	}

	keys, ok := result["response"].(map[string]interface{})["keys"].([]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid SSH keys response format")
		return
	}

	found := false
	for _, k := range keys {
		keyData, ok := k.(map[string]interface{})
		if !ok {
			continue
		}
		if keyData["uuid"] == state.ID.ValueString() {
			state.ID = types.StringValue(keyData["uuid"].(string))
			state.Name = types.StringValue(keyData["name"].(string))
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SSHKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/vm/sshkey/edit", map[string]interface{}{
		"ssh_key":  state.ID.ValueString(),
		"key_name": plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update SSH key: %v", err))
		return
	}

	state.Name = plan.Name
	state.Value = plan.Value

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SSHKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/vm/sshkey/delete", map[string]interface{}{
		"ssh_key": state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete SSH key: %v", err))
		return
	}
}

type SSHKeyModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}
