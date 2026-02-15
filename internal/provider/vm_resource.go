// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &VmResource{}

type VmResource struct {
	client *OneProviderClient
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oneprovider_vm"
}

func (r *VmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "OneProvider VM (OneCloud) resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VM ID",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "VM hostname",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "VM status",
				Computed:            true,
			},
			"ip_addr": schema.StringAttribute{
				MarkdownDescription: "VM IP address",
				Computed:            true,
			},
			"project_uuid": schema.StringAttribute{
				MarkdownDescription: "Project UUID",
				Optional:            true,
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: "Operating system template",
				Computed:            true,
			},
			"ram": schema.StringAttribute{
				MarkdownDescription: "RAM in MB",
				Computed:            true,
			},
			"cpu": schema.StringAttribute{
				MarkdownDescription: "Number of CPUs",
				Computed:            true,
			},
			"disk": schema.StringAttribute{
				MarkdownDescription: "Disk in GB",
				Computed:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Location ID for VM creation",
				Optional:            true,
			},
			"instance_size": schema.Int64Attribute{
				MarkdownDescription: "Instance size ID for VM creation",
				Optional:            true,
			},
			"template": schema.StringAttribute{
				MarkdownDescription: "OS template ID or image UUID for VM creation",
				Optional:            true,
			},
			"ssh_keys": schema.ListAttribute{
				MarkdownDescription: "Array of SSH Key UUIDs",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"enable_ipv6": schema.BoolAttribute{
				MarkdownDescription: "Enable IPv6 on the VM",
				Optional:            true,
			},
			"root_password": schema.StringAttribute{
				MarkdownDescription: "Root password (only returned on create)",
				Computed:            true,
				Sensitive:           true,
			},
			"iso_image": schema.StringAttribute{
				MarkdownDescription: "ISO image to mount (use empty string to unmount)",
				Optional:            true,
			},
			"reinstall_template": schema.StringAttribute{
				MarkdownDescription: "Template ID to reinstall VM with",
				Optional:            true,
			},
			"rescue": schema.BoolAttribute{
				MarkdownDescription: "Enable rescue mode (true) or disable (false)",
				Optional:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "VM action: boot, shutdown, reboot, poweroff",
				Optional:            true,
			},
		},
	}
}

func (r *VmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VmModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := map[string]interface{}{
		"hostname": plan.Hostname.ValueString(),
	}

	if !plan.LocationID.IsNull() {
		params["location_id"] = plan.LocationID.ValueInt64()
	}

	if !plan.InstanceSize.IsNull() {
		params["instance_size"] = plan.InstanceSize.ValueInt64()
	}

	if !plan.Template.IsNull() {
		params["template"] = plan.Template.ValueString()
	}

	if !plan.ProjectUUID.IsNull() {
		params["project_uuid"] = plan.ProjectUUID.ValueString()
	}

	if !plan.EnableIPv6.IsNull() && plan.EnableIPv6.ValueBool() {
		params["enable_ipv6"] = "1"
	}

	sshKeys := make([]string, 0, len(plan.SSHKeys.Elements()))
	if len(plan.SSHKeys.Elements()) > 0 {
		for _, key := range plan.SSHKeys.Elements() {
			if keyVal, ok := key.(types.String); ok {
				sshKeys = append(sshKeys, keyVal.ValueString())
			}
		}
		if len(sshKeys) > 0 {
			params["ssh_keys"] = sshKeys
		}
	}

	tflog.Debug(ctx, "Creating VM with params", map[string]interface{}{
		"hostname":      params["hostname"],
		"location_id":   params["location_id"],
		"instance_size": params["instance_size"],
		"template":      params["template"],
	})

	result, err := callAPI(ctx, r.client, "POST", "/vm/create", params)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create VM: %v", err))
		return
	}

	responseData, ok := result["response"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid response format from VM create")
		return
	}

	vmID := fmt.Sprintf("%v", responseData["id"])
	plan.ID = types.StringValue(vmID)

	if ip, ok := responseData["ip_address"]; ok {
		plan.IpAddr = types.StringValue(fmt.Sprintf("%v", ip))
	}

	if pw, ok := responseData["password"]; ok {
		plan.RootPassword = types.StringValue(fmt.Sprintf("%v", pw))
	}

	// Fetch VM info to populate all computed fields
	vmInfoResult, err := callAPI(ctx, r.client, "GET", fmt.Sprintf("/vm/info/%s", vmID), nil)
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to get VM info after creation: %v", err))
		// Don't fail, just use what we have
		plan.Status = types.StringValue("creating")
	} else {
		vmData, ok := vmInfoResult["response"].(map[string]interface{})
		if !ok {
			tflog.Warn(ctx, "Invalid response format from VM info")
			plan.Status = types.StringValue("creating")
		} else {
			// Populate all computed fields from the VM info response
			if hostname, ok := vmData["hostname"].(string); ok {
				plan.Hostname = types.StringValue(hostname)
			}

			if status, ok := vmData["status"].(string); ok {
				plan.Status = types.StringValue(status)
			}

			if serverInfo, ok := vmData["server_info"].(map[string]interface{}); ok {
				if ip, ok := serverInfo["ipaddress"].(string); ok {
					plan.IpAddr = types.StringValue(ip)
				}
				if template, ok := serverInfo["template"].(string); ok {
					plan.OS = types.StringValue(template)
				}
				if ram, ok := serverInfo["ram_mb"].(string); ok {
					plan.RAM = types.StringValue(ram)
				}
				if cpus, ok := serverInfo["cpus"].(string); ok {
					plan.CPU = types.StringValue(cpus)
				}
				if space, ok := serverInfo["space_gb"].(string); ok {
					plan.Disk = types.StringValue(space)
				}
				if projUUID, ok := serverInfo["project_uuid"].(string); ok {
					plan.ProjectUUID = types.StringValue(projUUID)
				}
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VmModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()
	if vmID == "" {
		resp.Diagnostics.AddError("Missing VM ID", "VM ID is required.")
		return
	}

	result, err := callAPI(ctx, r.client, "GET", fmt.Sprintf("/vm/info/%s", vmID), nil)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get VM info: %v", err))
		return
	}

	vmData, ok := result["response"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid response format from VM info")
		return
	}

	state.ID = types.StringValue(vmID)

	if hostname, ok := vmData["hostname"].(string); ok {
		state.Hostname = types.StringValue(hostname)
	}

	if status, ok := vmData["status"].(string); ok {
		state.Status = types.StringValue(status)
	}

	if serverInfo, ok := vmData["server_info"].(map[string]interface{}); ok {
		if ip, ok := serverInfo["ipaddress"].(string); ok {
			state.IpAddr = types.StringValue(ip)
		}
		if template, ok := serverInfo["template"].(string); ok {
			state.OS = types.StringValue(template)
		}
		if ram, ok := serverInfo["ram_mb"].(string); ok {
			state.RAM = types.StringValue(ram)
		}
		if cpus, ok := serverInfo["cpus"].(string); ok {
			state.CPU = types.StringValue(cpus)
		}
		if space, ok := serverInfo["space_gb"].(string); ok {
			state.Disk = types.StringValue(space)
		}
		if projUUID, ok := serverInfo["project_uuid"].(string); ok {
			state.ProjectUUID = types.StringValue(projUUID)
		}
		if iso, ok := serverInfo["iso"].(string); ok && iso != "" {
			state.IsoImage = types.StringValue(iso)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VmModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()

	if !plan.Hostname.IsNull() && plan.Hostname.ValueString() != state.Hostname.ValueString() {
		_, err := callAPI(ctx, r.client, "POST", "/vm/hostname", map[string]interface{}{
			"vm_id":    vmID,
			"hostname": plan.Hostname.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update hostname: %v", err))
			return
		}
		state.Hostname = plan.Hostname
	}

	if !plan.InstanceSize.IsNull() && plan.InstanceSize.ValueInt64() != state.InstanceSize.ValueInt64() {
		_, err := callAPI(ctx, r.client, "POST", "/vm/resize", map[string]interface{}{
			"vm_id":         vmID,
			"instance_size": plan.InstanceSize.ValueInt64(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to resize VM: %v", err))
			return
		}
		state.InstanceSize = plan.InstanceSize
	}

	if !plan.Action.IsNull() && plan.Action.ValueString() != "" {
		action := plan.Action.ValueString()
		endpoint := ""
		switch action {
		case "boot":
			endpoint = "/vm/boot"
		case "shutdown":
			endpoint = "/vm/shutdown"
		case "reboot":
			endpoint = "/vm/reboot"
		case "poweroff":
			endpoint = "/vm/poweroff"
		default:
			resp.Diagnostics.AddError("Invalid Action", fmt.Sprintf("Unknown action: %s", action))
			return
		}

		_, err := callAPI(ctx, r.client, "POST", endpoint, map[string]interface{}{
			"vm_id": vmID,
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to %s VM: %v", action, err))
			return
		}
		state.Action = types.StringValue("")
	}

	if !plan.Rescue.IsNull() {
		endpoint := "/vm/rescue/enable"
		if !plan.Rescue.ValueBool() {
			endpoint = "/vm/rescue/disable"
		}
		_, err := callAPI(ctx, r.client, "POST", endpoint, map[string]interface{}{
			"vm_id": vmID,
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to set rescue mode: %v", err))
			return
		}
		state.Rescue = types.BoolNull()
	}

	if !plan.IsoImage.IsNull() {
		currentISO := state.IsoImage.ValueString()
		newISO := plan.IsoImage.ValueString()

		if newISO != currentISO {
			if newISO == "" {
				_, err := callAPI(ctx, r.client, "POST", "/vm/unmountiso", map[string]interface{}{
					"vm_id": vmID,
				})
				if err != nil {
					resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to unmount ISO: %v", err))
					return
				}
			} else {
				_, err := callAPI(ctx, r.client, "POST", "/vm/mountiso", map[string]interface{}{
					"vm_id": vmID,
					"iso":   newISO,
				})
				if err != nil {
					resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to mount ISO: %v", err))
					return
				}
			}
			state.IsoImage = plan.IsoImage
		}
	}

	if !plan.ReinstallTemplate.IsNull() && plan.ReinstallTemplate.ValueString() != "" {
		_, err := callAPI(ctx, r.client, "POST", "/vm/reinstall", map[string]interface{}{
			"vm_id":    vmID,
			"template": plan.ReinstallTemplate.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to reinstall VM: %v", err))
			return
		}
		state.ReinstallTemplate = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VmModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()

	_, err := callAPI(ctx, r.client, "POST", "/vm/destroy", map[string]interface{}{
		"vm_id": vmID,
	})
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to destroy VM: %v", err))
		return
	}
}

type VmModel struct {
	ID                types.String `tfsdk:"id"`
	Hostname          types.String `tfsdk:"hostname"`
	Status            types.String `tfsdk:"status"`
	IpAddr            types.String `tfsdk:"ip_addr"`
	ProjectUUID       types.String `tfsdk:"project_uuid"`
	OS                types.String `tfsdk:"os"`
	RAM               types.String `tfsdk:"ram"`
	CPU               types.String `tfsdk:"cpu"`
	Disk              types.String `tfsdk:"disk"`
	LocationID        types.Int64  `tfsdk:"location_id"`
	InstanceSize      types.Int64  `tfsdk:"instance_size"`
	Template          types.String `tfsdk:"template"`
	SSHKeys           types.List   `tfsdk:"ssh_keys"`
	EnableIPv6        types.Bool   `tfsdk:"enable_ipv6"`
	RootPassword      types.String `tfsdk:"root_password"`
	IsoImage          types.String `tfsdk:"iso_image"`
	ReinstallTemplate types.String `tfsdk:"reinstall_template"`
	Rescue            types.Bool   `tfsdk:"rescue"`
	Action            types.String `tfsdk:"action"`
}
