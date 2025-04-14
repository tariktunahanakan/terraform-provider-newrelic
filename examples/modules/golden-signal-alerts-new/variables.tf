variable "service" {
  description = "The service to create alerts for"
  type = object({
    name                       = string
    team                       = string
    threshold_duration         = number
    cpu_threshold              = number
    response_time_threshold    = number
    error_percentage_threshold = number
    throughput_threshold       = number
  })
}

variable "custom_nrql_conditions" {
  description = "List of custom NRQL alert conditions"
  type = list(object({
    name        = string
    description = optional(string)
    query       = string
    threshold   = number
    operator    = string
    duration    = number
    priority    = optional(string)
  }))
  default = []
}
variable "notification_channel_ids" {
  description = "The IDs of notification channels to add to this policy"
  type        = list(string)
}
