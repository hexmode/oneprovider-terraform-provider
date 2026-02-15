# Manage SSH keys
resource "oneprovider_ssh_key" "example" {
  name  = "my SSH key"
  value = "ssh-rsa AAAA... user@hostname"
}

# Use SSH key in VM
# Note: location_id, instance_size, and template are IDs returned by the OneProvider API.
resource "oneprovider_vm" "with_ssh_key" {
  hostname      = "myserver.example.com"
  location_id   = 6
  instance_size = 108
  template      = "909"

  ssh_keys = [oneprovider_ssh_key.example.id]
}

# Import existing SSH key:
# terraform import oneprovider_ssh_key.example key-uuid-123
