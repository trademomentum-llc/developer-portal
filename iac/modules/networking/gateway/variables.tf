variable "routes" {
  description = "Map of hostname routes to backend services"
  type = map(object({
    hostname = string
    service = object({
      name      = string
      namespace = string
      port      = number
    })
  }))
}
