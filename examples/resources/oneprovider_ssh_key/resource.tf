# Manage SSH keys
resource "oneprovider_ssh_key" "example" {
  name  = "my SSH key"
  value = "ssh-rsa AAAA... user@hostname"
}

# Use SSH key in VM
resource "oneprovider_vm" "with_ssh_key" {
  hostname      = "myserver.example.com"
  location_id   = 1
  instance_size = 1
  template      = "almalinux-8.6"

  ssh_keys = [oneprovider_ssh_key.example.id]
}

# Import existing SSH key:
# terraform import oneprovider_ssh_key.example key-uuid-123
