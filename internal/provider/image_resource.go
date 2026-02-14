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
		MarkdownDescription: "OneProvider image resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Image ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Image name",
				Optional:            true,
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: "Operating system",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Image type",
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
	resp.Diagnostics.AddError("Not Implemented", "Image creation is not yet implemented.")
}

func (r *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError("Not Implemented", "Image read is not yet implemented.")
}

func (r *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Implemented", "Image update is not yet implemented.")
}

func (r *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Not Implemented", "Image deletion is not yet implemented.")
}

type ImageModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	OS   types.String `tfsdk:"os"`
	Type types.String `tfsdk:"type"`
}
