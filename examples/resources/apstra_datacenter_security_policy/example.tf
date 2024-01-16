# The following example creates a pair of policies which allow only HTTPS
# and ICMP traffic from clients on network "iBs6gKOBpZgBzXYaMzI"to servers
# on network "bIo0YmuXqFB4_BmAHAE".

resource "apstra_datacenter_security_policy" "client_traffic" {
  blueprint_id                     = "a52fb4ff-b352-46a3-9141-820b40972133"
  name                             = "http client traffic out"
  description                      = "Allow outbound to TCP/443 and ICMP traffic. Block all others."
  source_application_point_id      = "iBs6gKOBpZgBzXYaMzI"
  destination_application_point_id = "bIo0YmuXqFB4_BmAHAE"
  rules = [
    {
      name         = "permit_tcp_443"
      description  = "Permit client traffic on TCP/443 to servers"
      action       = "permit"
      protocol     = "tcp"
      source_ports = null // null == "any"
      destination_ports = [
        {
          from_port = 443
          to_port   = 443
        }
      ]
    },
    {
      name        = "permit_icmp"
      description = "Permit ICMP"
      action      = "permit_log"
      protocol    = "icmp"
    },
    {
      name        = "deny_all_ip"
      action      = "deny_log"
      protocol    = "ip"
    },
  ]
}

resource "apstra_datacenter_security_policy" "server_traffic" {
  blueprint_id                     = "a52fb4ff-b352-46a3-9141-820b40972133"
  name                             = "http server traffic replies"
  description                      = "Allow inbound established flows from TCP/443 and ICMP traffic. Block all others."
  source_application_point_id      = "bIo0YmuXqFB4_BmAHAE"
  destination_application_point_id = "iBs6gKOBpZgBzXYaMzI"
  rules = [
    {
      name        = "permit_tcp_443"
      description = "Permit server replies on TCP/443"
      action      = "permit"
      protocol    = "tcp"
      established = true
      source_ports = [
        {
          from_port = 443
          to_port   = 443
        }
      ]
      destination_ports = null // null == "any"
    },
    {
      name        = "permit_icmp"
      description = "Permit ICMP"
      action      = "permit_log"
      protocol    = "icmp"
    },
    {
      name        = "deny_all_ip"
      action      = "deny_log"
      protocol    = "ip"
    },
  ]
}
