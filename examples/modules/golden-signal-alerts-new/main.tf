

resource "newrelic_alert_policy" "golden_signal_policy" {
  name = "Golden Signals - ${var.service.name}"
}

resource "newrelic_nrql_alert_condition" "custom_conditions" {
  for_each = { for idx, cond in var.custom_nrql_conditions : idx => cond }

  policy_id = newrelic_alert_policy.golden_signal_policy.id
  name      = each.value.name
  enabled   = true

  nrql {
    query = each.value.query
  }

  critical {
    threshold             = each.value.threshold
    operator              = each.value.operator
    threshold_duration    = each.value.duration
    threshold_occurrences = "all"
  }

  type                         = "static"
  violation_time_limit_seconds = 259200
}

resource "newrelic_nrql_alert_condition" "response_time_web" {
  policy_id   = newrelic_alert_policy.golden_signal_policy.id
  name        = "High Response Time (web)"
  fill_option = "static"
  fill_value  = 0

  nrql {
    query = <<-EOT
      SELECT filter(average(newrelic.timeslice.value), WHERE metricTimesliceName = 'HttpDispatcher') OR 0
      FROM Metric
      WHERE label.team = '${var.service.team}'
      AND metricTimesliceName IN ('HttpDispatcher', 'Agent/MetricsReported/count')
      FACET appName
    EOT
  }

  critical {
    operator              = "above"
    threshold             = var.service.response_time_threshold
    threshold_duration    = var.service.threshold_duration
    threshold_occurrences = "all"
  }
}

resource "newrelic_nrql_alert_condition" "throughput_web" {
  policy_id   = newrelic_alert_policy.golden_signal_policy.id
  name        = "Low Throughput (web)"
  fill_option = "static"
  fill_value  = 0

  nrql {
    query = <<-EOT
      SELECT filter(count(newrelic.timeslice.value), WHERE metricTimesliceName = 'HttpDispatcher') OR 0
      FROM Metric
      WHERE label.team = '${var.service.team}'
      AND metricTimesliceName IN ('HttpDispatcher', 'Agent/MetricsReported/count')
      FACET appName
    EOT
  }

  critical {
    operator              = "below"
    threshold             = var.service.throughput_threshold
    threshold_duration    = var.service.threshold_duration
    threshold_occurrences = "all"
  }
}

resource "newrelic_nrql_alert_condition" "error_percentage" {
  policy_id   = newrelic_alert_policy.golden_signal_policy.id
  name        = "High Error Percentage"
  fill_option = "static"
  fill_value  = 0

  nrql {
    query = <<-EOT
      SELECT ((filter(count(newrelic.timeslice.value), where metricTimesliceName = 'Errors/all')
            / filter(count(newrelic.timeslice.value), WHERE metricTimesliceName IN ('HttpDispatcher', 'OtherTransaction/all'))) OR 0) * 100
      FROM Metric
      WHERE label.team = '${var.service.team}'
      AND metricTimesliceName IN ('Errors/all', 'HttpDispatcher', 'OtherTransaction/all', 'Agent/MetricsReported/count')
      FACET appName
    EOT
  }

  critical {
    operator              = "above"
    threshold             = var.service.error_percentage_threshold
    threshold_duration    = var.service.threshold_duration
    threshold_occurrences = "all"
  }
}

resource "newrelic_nrql_alert_condition" "high_cpu" {
  policy_id   = newrelic_alert_policy.golden_signal_policy.id
  name        = "High CPU usage"
  fill_option = "static"
  fill_value  = 0

  nrql {
    query = <<-EOT
      SELECT average(cpuPercent)
      FROM SystemSample
      WHERE label.team = '${var.service.team}'
      FACET entityId
    EOT
  }

  critical {
    operator              = "above"
    threshold             = var.service.cpu_threshold
    threshold_duration    = var.service.threshold_duration
    threshold_occurrences = "all"
  }
}

resource "newrelic_workflow" "golden_signal_workflow" {
  name                  = "Golden Signals Workflow ${var.service.name}"
  muting_rules_handling = "NOTIFY_ALL_ISSUES"

  issues_filter {
    name = " Golden signal policy Ids filter"
    type = "FILTER"

    predicate {
      attribute = "labels.policyIds"
      operator  = "EXACTLY_MATCHES"
      values    = [newrelic_alert_policy.golden_signal_policy.id]
    }
  }
  dynamic "destination" {
    for_each = var.notification_channel_ids
    content {
      channel_id = destination.value
    }
  }
}