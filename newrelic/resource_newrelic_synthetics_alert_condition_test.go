//go:build integration || ALERTS
// +build integration ALERTS

package newrelic

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccNewRelicSyntheticsAlertCondition_Basic(t *testing.T) {
	resourceName := "newrelic_synthetics_alert_condition.foo"
	rName := generateNameForIntegrationTestResource()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckEnvVars(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNewRelicSyntheticsAlertConditionDestroy,
		Steps: []resource.TestStep{
			// Test: Create
			{
				Config: testAccNewRelicSyntheticsAlertConditionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicSyntheticsAlertConditionExists(resourceName),
				),
			},
			// Test: No diff on re-apply
			{
				Config:             testAccNewRelicSyntheticsAlertConditionConfig(rName),
				ExpectNonEmptyPlan: false,
			},

			// Test: Update
			{
				Config: testAccCheckNewRelicSyntheticsAlertConditionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicSyntheticsAlertConditionExists(resourceName),
				),
			},
			// Test: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"monitor_id"},
			},
		},
	})
}

func TestAccNewRelicSyntheticsAlertCondition_MissingPolicy(t *testing.T) {
	rName := generateNameForIntegrationTestResource()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckEnvVars(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNewRelicSyntheticsAlertConditionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNewRelicSyntheticsAlertConditionConfig(rName),
			},
			{
				PreConfig: testAccDeleteNewRelicAlertPolicy(fmt.Sprintf("tf-test-%s", rName)),
				Config:    testAccNewRelicSyntheticsAlertConditionConfig(rName),
				Check:     testAccCheckNewRelicSyntheticsAlertConditionExists("newrelic_synthetics_alert_condition.foo"),
			},
		},
	})
}

func testAccCheckNewRelicSyntheticsAlertConditionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderConfig).NewClient
	for _, r := range s.RootModule().Resources {
		if r.Type != "newrelic_synthetics_alert_condition" {
			continue
		}

		ids, err := parseIDs(r.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		_, err = client.Alerts.GetSyntheticsCondition(policyID, id)
		if err == nil {
			return fmt.Errorf("synthetics alert condition still exists")
		}

	}
	return nil
}

func testAccCheckNewRelicSyntheticsAlertConditionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no alert condition ID is set")
		}

		client := testAccProvider.Meta().(*ProviderConfig).NewClient

		ids, err := parseIDs(rs.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		found, err := client.Alerts.GetSyntheticsCondition(policyID, id)
		if err != nil {
			return err
		}

		if found.ID != id {
			return fmt.Errorf("alert condition not found: %v - %v", id, found)
		}

		return nil
	}
}

func testAccNewRelicSyntheticsAlertConditionConfig(name string) string {
	return fmt.Sprintf(`
resource "newrelic_synthetics_monitor" "bar" {
	name      = "tf-test-synthetic-%[1]s"
	type      = "SIMPLE"
	period    = "EVERY_10_MINUTES"
	status    = "DISABLED"
	locations_public = ["US_EAST_1"]
	uri       = "https://google.com"
}

resource "newrelic_alert_policy" "foo" {
	name = "tf-test-%[1]s"
}

resource "newrelic_synthetics_alert_condition" "foo" {
	policy_id   = newrelic_alert_policy.foo.id
	name        = "tf-test-%[1]s"
	monitor_id  = newrelic_synthetics_monitor.bar.id
	runbook_url = "www.example.com"
	enabled     = "true"
}
`, name)
}

func testAccCheckNewRelicSyntheticsAlertConditionUpdated(name string) string {
	return fmt.Sprintf(`
resource "newrelic_synthetics_monitor" "bar" {
	name      = "tf-test-synthetic-%[1]s"
	type      = "SIMPLE"
	period    = "EVERY_15_MINUTES"
	status    = "DISABLED"
	locations_public = ["US_EAST_1"]
	uri       = "https://google.com"
}

resource "newrelic_alert_policy" "foo" {
	name = "tf-test-%[1]s"
}

resource "newrelic_synthetics_alert_condition" "foo" {
	policy_id   = newrelic_alert_policy.foo.id
	name        = "tf-test-%[1]s"
	monitor_id  = newrelic_synthetics_monitor.bar.id
	runbook_url = "www.example-updated.com"
	enabled     = "false"
}
`, name)
}

func TestAccNewRelicSyntheticsAlertConditionCheckMonitorIdEmpty(t *testing.T) {
	rName := generateNameForIntegrationTestResource()
	expectedMsg, _ := regexp.Compile("expected \"monitor_id\" to not be an empty string, got")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheckEnvVars(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccNewRelicSyntheticsAlertConditionConfigCheckMonitorIdEmpty(rName),
				ExpectError: expectedMsg,
			},
		},
	})
}

func testAccNewRelicSyntheticsAlertConditionConfigCheckMonitorIdEmpty(name string) string {
	return fmt.Sprintf(`
resource "newrelic_alert_policy" "foo" {
	name = "tf-test-%[1]s"
}

resource "newrelic_synthetics_alert_condition" "foo" {
	policy_id   = newrelic_alert_policy.foo.id
	name        = "tf-test-%[1]s"
	monitor_id  = ""
	runbook_url = "www.example.com"
	enabled     = "true"
}
`, name)
}
