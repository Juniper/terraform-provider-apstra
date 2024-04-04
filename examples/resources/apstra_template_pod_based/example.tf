# This example creates a 5-stage template with two super spine
# planes and two types of pods

resource "apstra_template_pod_based" "example" {
  name                   = "dual-plane"
  super_spine = {
    logical_device_id = "AOS-24x10-2"
    per_plane_count   = 4
    plane_count       = 2
  }
  pod_infos = {
    pod_single = { count = 2 } # pod_single and pod_mlag are the IDs of the 3-stage (rack-based)
    pod_mlag   = { count = 2 } # templates used as pods in this 5-stage (pod-based) template
  }
}
