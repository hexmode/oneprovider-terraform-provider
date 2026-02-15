// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUnitSSHKeyModel(t *testing.T) {
	model := SSHKeyModel{
		ID:    types.StringValue("test-uuid"),
		Name:  types.StringValue("test-key"),
		Value: types.StringValue("ssh-rsa AAAAB3..."),
	}

	if model.ID.ValueString() != "test-uuid" {
		t.Errorf("Expected ID to be 'test-uuid', got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "test-key" {
		t.Errorf("Expected Name to be 'test-key', got %s", model.Name.ValueString())
	}
	if model.Value.ValueString() != "ssh-rsa AAAAB3..." {
		t.Errorf("Expected Value to be 'ssh-rsa AAAAB3...', got %s", model.Value.ValueString())
	}
}

func TestUnitRdnsModel(t *testing.T) {
	model := RdnsModel{
		IpAddress: types.StringValue("192.168.1.1"),
		Domain:    types.StringValue("example.com"),
		Ttl:       types.Int64Value(3600),
	}

	if model.IpAddress.ValueString() != "192.168.1.1" {
		t.Errorf("Expected IP to be '192.168.1.1', got %s", model.IpAddress.ValueString())
	}
	if model.Domain.ValueString() != "example.com" {
		t.Errorf("Expected Domain to be 'example.com', got %s", model.Domain.ValueString())
	}
	if model.Ttl.ValueInt64() != 3600 {
		t.Errorf("Expected TTL to be 3600, got %d", model.Ttl.ValueInt64())
	}
}

func TestUnitImageModel(t *testing.T) {
	model := ImageModel{
		ID:        types.StringValue("img-123"),
		Name:      types.StringValue("My Image"),
		VmID:      types.StringValue("vm-456"),
		Size:      types.Int64Value(20),
		OsName:    types.StringValue("almalinux"),
		OsDisplay: types.StringValue("AlmaLinux 8.6 64bits"),
	}

	if model.ID.ValueString() != "img-123" {
		t.Errorf("Expected ID to be 'img-123', got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "My Image" {
		t.Errorf("Expected Name to be 'My Image', got %s", model.Name.ValueString())
	}
	if model.Size.ValueInt64() != 20 {
		t.Errorf("Expected Size to be 20, got %d", model.Size.ValueInt64())
	}
}

func TestUnitVmModel(t *testing.T) {
	model := VmModel{
		ID:           types.StringValue("vm-123"),
		Hostname:     types.StringValue("test-vm"),
		Status:       types.StringValue("Active"),
		IpAddr:       types.StringValue("10.0.0.1"),
		InstanceSize: types.Int64Value(1024),
		Action:       types.StringValue("reboot"),
	}

	if model.ID.ValueString() != "vm-123" {
		t.Errorf("Expected ID to be 'vm-123', got %s", model.ID.ValueString())
	}
	if model.Hostname.ValueString() != "test-vm" {
		t.Errorf("Expected Hostname to be 'test-vm', got %s", model.Hostname.ValueString())
	}
	if model.Status.ValueString() != "Active" {
		t.Errorf("Expected Status to be 'Active', got %s", model.Status.ValueString())
	}
	if model.Action.ValueString() != "reboot" {
		t.Errorf("Expected Action to be 'reboot', got %s", model.Action.ValueString())
	}
}
