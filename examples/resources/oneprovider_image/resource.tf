# Create an image from an existing VM
resource "oneprovider_image" "example" {
  name  = "my-backup-image"
  vm_id = "12345"
}

# Use an image when creating a VM
# Note: for VM creation, template accepts either an OS template ID (string) or an image UUID.
resource "oneprovider_vm" "from_image" {
  hostname      = "restored.example.com"
  location_id   = 6
  instance_size = 108
  template      = oneprovider_image.example.id
}

# Import existing image:
# terraform import oneprovider_image.example image-id-123
