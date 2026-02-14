# Create an image from an existing VM
resource "oneprovider_image" "example" {
  name   = "my-backup-image"
  vm_id  = "12345"
}

# List available images to get ID
data "oneprovider_configuration" "all" {}

# Use image when creating VM
resource "oneprovider_vm" "from_image" {
  hostname      = "restored.example.com"
  location_id   = 1
  instance_size = 1
  template      = "my-backup-image"
}

# Import existing image:
# terraform import oneprovider_image.example image-id-123
