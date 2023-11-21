# This example returns the result of a graph query that identifies any
# leaf devices in the blueprint

locals {
  blueprint_id = "abc-123"
}

data "apstra_datacenter_graph_query" "query_leafs" {
  blueprint_id = local.blueprint_id
  query = "node('system', role='leaf', name='target')"
}

locals {
  query_leafs_result = data.apstra_datacenter_graph_query.query_leafs.result
}
