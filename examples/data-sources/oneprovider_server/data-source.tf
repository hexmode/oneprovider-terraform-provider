# Get information about a specific server
data "oneprovider_server" "example" {
  id = "12345"
}

output "server_ip" {
  value = data.oneprovider_server.example.ip_addr
}
