# Set reverse DNS for a VM's IP
resource "oneprovider_rdns" "example" {
  ip_address = "192.168.1.100"
  domain     = "myserver.example.com"
}

# Update reverse DNS
resource "oneprovider_rdns" "example" {
  ip_address = "192.168.1.100"
  domain     = "new-name.example.com"
}

# Remove reverse DNS (set domain to empty)
resource "oneprovider_rdns" "example" {
  ip_address = "192.168.1.100"
  domain     = ""
}

# Import existing RDNS:
# terraform import oneprovider_rdns.example 192.168.1.100
