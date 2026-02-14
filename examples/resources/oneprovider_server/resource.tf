# Manage a dedicated server
resource "oneprovider_server" "example" {
  id = "12345"

  # Update hostname
  hostname = "myserver.example.com"
}

# Import existing server:
# terraform import oneprovider_server.example 12345
