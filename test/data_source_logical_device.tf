data "apstra_logical_device" "by_id" {
  id = "AOS-8x10-1"
}

output "apsta_logical_device_by_id" {
  value = data.apstra_logical_device.by_id
}