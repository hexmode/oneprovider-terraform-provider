# Create a VM
resource "oneprovider_vm" "example" {
  hostname      = "myvm.example.com"
  location_id   = 1
  instance_size = 1
  template      = "almalinux-8.6"

  # Optional settings
  enable_ipv6 = false
  ssh_keys     = ["sshkey-uuid-123"]
}

# VM operations (via Update)
resource "oneprovider_vm" "example" {
  # ... existing config ...

  # Trigger VM action (boot, shutdown, reboot, poweroff)
  # Note: This will trigger the action on every apply
  # action = "reboot"

  # Enable rescue mode
  # rescue = true

  # Mount an ISO
  # iso_image = "CentOS-7-x86_64.iso"

  # Unmount ISO (set to empty)
  # iso_image = ""

  # Reinstall with different template
  # reinstall_template = "centos-7-x86_64"
}

# Import existing VM:
# terraform import oneprovider_vm.example 12345
