# This example "acknowledges" a managed device after checking that the
# discovered serial number (device_key) matches the supplied value. The
# only managed device checked is the one managed by the specified System
# Agent.

resource "apstra_managed_device_ack" "spine1" {
  agent_id = "d4689206-fe4b-4e39-9a96-3a398f319660"
  device_key = "525400BB9B5F"
}
