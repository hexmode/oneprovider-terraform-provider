// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccConfigurationDataSource tests reading the configuration data source.
func TestAccConfigurationDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Just verify the data source was read successfully
					resource.TestCheckResourceAttrSet("data.oneprovider_configuration.test", "id"),
				),
			},
		},
	})
}

func testAccConfigurationDataSourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

data "oneprovider_configuration" "test" {}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

// TestAccSSHKeyResource tests SSH key CRUD operations.
func TestAccSSHKeyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyResourceConfig("test-key-" + strconv.FormatInt(time.Now().Unix(), 10)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_ssh_key.test", "id"),
				),
			},
		},
	})
}

func testAccSSHKeyResourceConfig(name string) string {
	// Try to read a real SSH key for testing
	sshKey := os.Getenv("ONEPROVIDER_TEST_SSH_KEY")
	if sshKey == "" {
		// Fallback: try to read from default SSH location
		if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_ed25519.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		} else if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_ecdsa.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		} else if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		}
	}
	// Fallback to a placeholder if no key found (will fail at API level)
	if sshKey == "" {
		sshKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC+randomkey123 test@example.com"
	}

	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_ssh_key" "test" {
  name  = "%s"
  value = "%s"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"), name, sshKey)
}

// TestAccVMResource tests VM creation and basic operations.
func TestAccVMResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("oneprovider_vm.test", "id", func(value string) error {
						visualCheck("VM created with ID: " + value + " - check OneProvider panel")
						return nil
					}),
					resource.TestCheckResourceAttr("oneprovider_vm.test", "hostname", "tf-test-vm"),
					resource.TestCheckResourceAttrSet("oneprovider_vm.test", "ip_addr"),
					resource.TestCheckResourceAttrSet("oneprovider_vm.test", "status"),
				),
			},
			{
				ResourceName:            "oneprovider_vm.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"root_password", "action", "rescue", "iso_image", "reinstall_template"},
			},
			// Test hostname update
			{
				Config: testAccVMResourceConfigUpdatedHostname(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oneprovider_vm.test", "hostname", "tf-test-vm-updated"),
				),
			},
		},
	})
}

func testAccVMResourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "test" {
  hostname      = "tf-test-vm"
  location_id   = 6
  instance_size = 108
  template      = "909"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

func testAccVMResourceConfigUpdatedHostname() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "test" {
  hostname      = "tf-test-vm-updated"
  location_id   = 6
  instance_size = 108
  template      = "909"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

// TestAccVMResourceWithSSHKey tests VM creation with SSH key.
func TestAccVMResourceWithSSHKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMResourceWithSSHKeyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_vm.test", "id"),
					resource.TestCheckResourceAttrSet("oneprovider_vm.test", "ip_addr"),
				),
			},
		},
	})
}

func testAccVMResourceWithSSHKeyConfig() string {
	// Try to read a real SSH key for testing
	sshKey := os.Getenv("ONEPROVIDER_TEST_SSH_KEY")
	if sshKey == "" {
		// Fallback: try to read from default SSH location
		if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_ed25519.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		} else if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_ecdsa.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		} else if data, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa.pub"); err == nil {
			sshKey = strings.TrimSpace(string(data))
		}
	}
	// Fallback to a placeholder if no key found (will fail at API level)
	if sshKey == "" {
		sshKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC+randomkey123 test@example.com"
	}

	// Use unique key name to avoid conflicts
	keyName := "tf-test-vm-key-" + strconv.FormatInt(time.Now().Unix(), 10)

	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_ssh_key" "test" {
  name  = "%s"
  value = "%s"
}

resource "oneprovider_vm" "test" {
  hostname      = "tf-test-vm-with-key"
  location_id   = 6
  instance_size = 108
  template      = "909"
  ssh_keys      = [oneprovider_ssh_key.test.id]
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"), keyName, sshKey)
}

// TestAccRdnsResource tests reverse DNS operations.
func TestAccRdnsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// First create a VM to get an IP address
				Config: testAccVMResourceForRdnsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_vm.rdns_test", "ip_addr"),
				),
			},
			{
				// Then set RDNS using the VM's IP
				Config: testAccRdnsResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_rdns.test", "domain"),
				),
			},
			{
				ResourceName:            "oneprovider_rdns.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vm_id"},
			},
			{
				// Update RDNS
				Config: testAccRdnsResourceConfigUpdated(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("oneprovider_rdns.test", "domain", "updated.example.com"),
				),
			},
		},
	})
}

func testAccVMResourceForRdnsConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "rdns_test" {
  hostname      = "tf-test-vm-rdns"
  location_id   = 6
  instance_size = 108
  template      = "909"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

func testAccRdnsResourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "rdns_test" {
  hostname      = "tf-test-vm-rdns"
  location_id   = 6
  instance_size = 108
  template      = "909"
}

resource "oneprovider_rdns" "test" {
  ip_address = oneprovider_vm.rdns_test.ip_addr
  domain     = "original.example.com"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

func testAccRdnsResourceConfigUpdated() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "rdns_test" {
  hostname      = "tf-test-vm-rdns"
  location_id   = 6
  instance_size = 108
  template      = "909"
}

resource "oneprovider_rdns" "test" {
  ip_address = oneprovider_vm.rdns_test.ip_addr
  domain     = "updated.example.com"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

// TestAccImageResource tests image creation from VM.
func TestAccImageResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// First create a VM to create image from
				Config: testAccVMResourceForImageConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_vm.image_test", "id"),
				),
			},
			{
				// Create image from VM
				Config: testAccImageResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_image.test", "id"),
					resource.TestCheckResourceAttr("oneprovider_image.test", "name", "tf-test-image"),
				),
			},
		},
	})
}

func testAccVMResourceForImageConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "image_test" {
  hostname      = "tf-test-vm-image"
  location_id   = 6
  instance_size = 108
  template      = "909"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

func testAccImageResourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "image_test" {
  hostname      = "tf-test-vm-image"
  location_id   = 6
  instance_size = 108
  template      = "909"
}

resource "oneprovider_image" "test" {
  name  = "tf-test-image"
  vm_id = oneprovider_vm.image_test.id
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

// TestAccServerDataSource tests reading server information.
func TestAccServerDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// First create a VM to get an ID
				Config: testAccVMResourceForServerDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oneprovider_vm.server_test", "id"),
				),
			},
			{
				// Then read it via data source
				Config: testAccServerDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.oneprovider_server.test", "id"),
					resource.TestCheckResourceAttrSet("data.oneprovider_server.test", "ip_addr"),
					resource.TestCheckResourceAttrSet("data.oneprovider_server.test", "hostname"),
				),
			},
		},
	})
}

func testAccVMResourceForServerDataSourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "server_test" {
  hostname      = "tf-test-vm-ds"
  location_id   = 6
  instance_size = 108
  template      = "909"
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

func testAccServerDataSourceConfig() string {
	return fmt.Sprintf(`
provider "oneprovider" {
  api_key    = "%s"
  client_key = "%s"
}

resource "oneprovider_vm" "server_test" {
  hostname      = "tf-test-vm-ds"
  location_id   = 6
  instance_size = 108
  template      = "909"
}

data "oneprovider_server" "test" {
  id = oneprovider_vm.server_test.id
}
`, os.Getenv("ONEPROVIDER_API_KEY"), os.Getenv("ONEPROVIDER_CLIENT_KEY"))
}

// testAccPreCheck checks if acceptance tests can run.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ONEPROVIDER_API_KEY"); v == "" {
		t.Fatal("ONEPROVIDER_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("ONEPROVIDER_CLIENT_KEY"); v == "" {
		t.Fatal("ONEPROVIDER_CLIENT_KEY must be set for acceptance tests")
	}
}

// visualCheck pauses for manual verification when VISUAL_CHECK env var is set.
// Pass a message describing what to verify.
func visualCheck(message string) {
	if os.Getenv("VISUAL_CHECK") != "" {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("VISUAL CHECK:", message)
		fmt.Println("Press Enter to continue...")
		_, _ = fmt.Scanln()
		fmt.Println(strings.Repeat("=", 60) + "\n")
	}
}
