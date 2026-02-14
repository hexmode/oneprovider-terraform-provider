// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ImageResource{}

type ImageResource struct {
	client *OneProviderClient
}

func NewImageResource() resource.Resource {
	return &ImageResource{}
}

func (r *ImageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oneprovider_image"
}

func (r *ImageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider image resource (like AWS AMI).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Image ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Image display name",
				Optional:            true,
				Computed:            true,
			},
			"vm_id": schema.StringAttribute{
				MarkdownDescription: "Source VM ID to create image from",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Image size in GB",
				Computed:            true,
			},
			"date": schema.StringAttribute{
				MarkdownDescription: "Creation date",
				Computed:            true,
			},
			"os_name": schema.StringAttribute{
				MarkdownDescription: "OS name",
				Computed:            true,
			},
			"os_display": schema.StringAttribute{
				MarkdownDescription: "OS display name",
				Computed:            true,
			},
		},
	}
}

func (r *ImageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ImageModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vmID := plan.VmID.ValueString()
	if vmID == "" {
		resp.Diagnostics.AddError("Missing VM ID", "vm_id is required to create an image")
		return
	}

	result, err := callAPI(ctx, r.client, "POST", "/vm/image/create", map[string]interface{}{
		"vm_id": vmID,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create image: %v", err))
		return
	}

	responseData, ok := result["response"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid response from image create")
		return
	}

	message, _ := responseData["message"].(string)
	plan.ID = types.StringValue(fmt.Sprintf("pending-%s", vmID))
	plan.Name = types.StringValue(fmt.Sprintf("Image from VM %s", vmID))
	plan.Size = types.Int64Value(0)
	plan.Date = types.StringValue("")
	plan.OsName = types.StringValue("")
	plan.OsDisplay = types.StringValue(message)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ImageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	imageID := state.ID.ValueString()
	if imageID == "" {
		resp.Diagnostics.AddError("Missing Image ID", "Image ID is required")
		return
	}

	result, err := callAPI(ctx, r.client, "GET", "/vm/images/list", nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to list images: %v", err))
		return
	}

	images, ok := result["response"].(map[string]interface{})["images"].([]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid images response format")
		return
	}

	found := false
	for _, img := range images {
		imgData, ok := img.(map[string]interface{})
		if !ok {
			continue
		}
		if imgData["id"] == imageID {
			state.ID = types.StringValue(imgData["id"].(string)) // nolint: forcetypeassert

			if name, ok := imgData["name"].(string); ok {
				state.Name = types.StringValue(name)
			}

			if size, ok := imgData["size"].(float64); ok {
				state.Size = types.Int64Value(int64(size))
			}

			if date, ok := imgData["date"].(string); ok {
				state.Date = types.StringValue(date)
			}

			if os, ok := imgData["os"].(map[string]interface{}); ok {
				if osName, ok := os["name"].(string); ok {
					state.OsName = types.StringValue(osName)
				}
				if osDisplay, ok := os["display"].(string); ok {
					state.OsDisplay = types.StringValue(osDisplay)
				}
			}

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

func (r *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ImageModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.IsNull() && plan.Name.ValueString() != state.Name.ValueString() {
		_, err := callAPI(ctx, r.client, "POST", "/vm/image/rename", map[string]interface{}{
			"image":    state.ID.ValueString(),
			"new_name": plan.Name.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to rename image: %v", err))
			return
		}
		state.Name = plan.Name
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ImageModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	imageID := state.ID.ValueString()
	if strings.HasPrefix(imageID, "pending-") {
		return
	}

	_, err := callAPI(ctx, r.client, "POST", "/vm/image/delete", map[string]interface{}{
		"image": imageID,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete image: %v", err))
		return
	}
}

type ImageModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	VmID      types.String `tfsdk:"vm_id"`
	Size      types.Int64  `tfsdk:"size"`
	Date      types.String `tfsdk:"date"`
	OsName    types.String `tfsdk:"os_name"`
	OsDisplay types.String `tfsdk:"os_display"`
}
