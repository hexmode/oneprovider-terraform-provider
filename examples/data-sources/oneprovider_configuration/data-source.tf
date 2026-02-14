# Get available configurations, sizes, templates, and locations
data "oneprovider_configuration" "all" {}

output "available_locations" {
  value = data.oneprovider_configuration.all.locations
}

output "available_templates" {
  value = data.oneprovider_configuration.all.templates
}
