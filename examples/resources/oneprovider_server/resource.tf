# Manage a dedicated server
# Note: oneprovider_server is intended to manage an *existing* server. Import it first.
resource "oneprovider_server" "example" {
  # Update hostname
  hostname = "myserver.example.com"
}

# Import existing server:
# terraform import oneprovider_server.example 12345
