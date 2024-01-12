# details of a security policy can be retrieved by ID or by name

# by id
data "apstra_datacenter_security_policy" "by_id" {
  blueprint_id = "a52fb4ff-b352-46a3-9141-820b40972133"
  id           = "K8g10S8W3oZIkRkBn0w"
}

# by name
data "apstra_datacenter_security_policy" "by_name" {
  blueprint_id = "a52fb4ff-b352-46a3-9141-820b40972133"
  name         = "my_policy"
}
