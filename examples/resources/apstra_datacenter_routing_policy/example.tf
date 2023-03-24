# This example creates a routing policy within the blueprint named "production"

data "apstra_datacenter_blueprint" "prod" {
  name = "production"
}

resource "apstra_datacenter_routing_policy" "just_pull_every_available_lever" {
  blueprint_id  = data.apstra_datacenter_blueprint.prod.id
  name          = "nope"
  description   = "Nothing good can come from this"
  import_policy = "defult_only" // "default_only" is the default. other options: "all" "extra_only"
  extra_imports = [
    { prefix = "10.0.0.0/8",                                 action = "deny"   },
    { prefix = "11.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "deny"   },
    { prefix = "12.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "deny"   },
    { prefix = "13.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "deny"   },
    { prefix = "14.0.0.0/8", ge_mask = 9,                    action = "deny"   },
    { prefix = "15.0.0.0/8", ge_mask = 32,                   action = "deny"   },
    { prefix = "16.0.0.0/8",                 le_mask = 9,    action = "deny"   },
    { prefix = "17.0.0.0/8",                 le_mask = 32,   action = "deny"   },
    { prefix = "20.0.0.0/8",                                 action = "permit" },
    { prefix = "21.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "permit" },
    { prefix = "22.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "permit" },
    { prefix = "23.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "permit" },
    { prefix = "24.0.0.0/8", ge_mask = 9,                    action = "permit" },
    { prefix = "25.0.0.0/8", ge_mask = 32,                   action = "permit" },
    { prefix = "26.0.0.0/8",                 le_mask = 9,    action = "permit" },
    { prefix = "27.0.0.0/8",                 le_mask = 32,   action = "permit" },
    { prefix = "30.0.0.0/8",                                                   }, // default action is "permit"
    { prefix = "31.0.0.0/8", ge_mask = 31,   le_mask = 32,                     }, // default action is "permit"
    { prefix = "32.0.0.0/8", ge_mask = 9,    le_mask = 10,                     }, // default action is "permit"
    { prefix = "33.0.0.0/8", ge_mask = 9,    le_mask = 32,                     }, // default action is "permit"
    { prefix = "34.0.0.0/8", ge_mask = 9,                                      }, // default action is "permit"
    { prefix = "35.0.0.0/8", ge_mask = 32,                                     }, // default action is "permit"
    { prefix = "36.0.0.0/8",                 le_mask = 9,                      }, // default action is "permit"
    { prefix = "37.0.0.0/8",                 le_mask = 32,                     }, // default action is "permit"
  ]
  extra_exports = [
    { prefix = "40.0.0.0/8",                                 action = "deny"   },
    { prefix = "41.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "deny"   },
    { prefix = "42.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "deny"   },
    { prefix = "43.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "deny"   },
    { prefix = "44.0.0.0/8", ge_mask = 9,                    action = "deny"   },
    { prefix = "45.0.0.0/8", ge_mask = 32,                   action = "deny"   },
    { prefix = "46.0.0.0/8",                 le_mask = 9,    action = "deny"   },
    { prefix = "47.0.0.0/8",                 le_mask = 32,   action = "deny"   },
    { prefix = "50.0.0.0/8",                                 action = "permit" },
    { prefix = "51.0.0.0/8", ge_mask = 31,   le_mask = 32,   action = "permit" },
    { prefix = "52.0.0.0/8", ge_mask = 9,    le_mask = 10,   action = "permit" },
    { prefix = "53.0.0.0/8", ge_mask = 9,    le_mask = 32,   action = "permit" },
    { prefix = "54.0.0.0/8", ge_mask = 9,                    action = "permit" },
    { prefix = "55.0.0.0/8", ge_mask = 32,                   action = "permit" },
    { prefix = "56.0.0.0/8",                 le_mask = 9,    action = "permit" },
    { prefix = "57.0.0.0/8",                 le_mask = 32,   action = "permit" },
    { prefix = "60.0.0.0/8",                                                   }, // default action is "permit"
    { prefix = "61.0.0.0/8", ge_mask = 31,   le_mask = 32,                     }, // default action is "permit"
    { prefix = "62.0.0.0/8", ge_mask = 9,    le_mask = 10,                     }, // default action is "permit"
    { prefix = "63.0.0.0/8", ge_mask = 9,    le_mask = 32,                     }, // default action is "permit"
    { prefix = "64.0.0.0/8", ge_mask = 9,                                      }, // default action is "permit"
    { prefix = "65.0.0.0/8", ge_mask = 32,                                     }, // default action is "permit"
    { prefix = "66.0.0.0/8",                 le_mask = 9,                      }, // default action is "permit"
    { prefix = "67.0.0.0/8",                 le_mask = 32,                     }, // default action is "permit"
  ]
  export_policy = {
#    export_spine_leaf_links       = false // default value is "false"
#    export_spine_superspine_links = false // default value is "false"
#    export_l3_edge_server_links   = false // default value is "false"
    export_l2_edge_subnets        = true
    export_loopbacks              = false // but it's okay if you type it in
    export_static_routes          = null  // you can also use "null" to get "false"
  }
  aggregate_prefixes = [
    "0.0.0.0/0",  // any prefix is okay here...
    "0.0.0.0/1",
    "0.0.0.0/2",
    "0.0.0.0/3",  // but it has to land on a "zero address"...
    "0.0.0.0/4",
    "0.0.0.0/5",
    "0.0.0.0/6",  // so, "1.0.0.0/6" would not be okay...
    "0.0.0.0/7",
    "0.0.0.0/8",
    "1.0.0.0/9",  // but "1.0.0.0/9" is fine...
    "0.0.0.0/10",
    "0.0.0.0/11",
    "0.0.0.0/12", // Burma-Shave
    "0.0.0.0/13",
    "0.0.0.0/14",
    "0.0.0.0/15",
    "0.0.0.0/16",
    "0.0.0.0/17",
    "0.0.0.0/18",
    "0.0.0.0/19",
    "0.0.0.0/20",
    "0.0.0.0/21",
    "0.0.0.0/22",
    "0.0.0.0/23",
    "0.0.0.0/24",
    "0.0.0.0/25",
    "0.0.0.0/26",
    "0.0.0.0/27",
    "0.0.0.0/28",
    "0.0.0.0/29",
    "0.0.0.0/30",
    "0.0.0.0/31",
    "0.0.0.0/32",
    "255.255.255.255/32"
  ]
  expect_default_ipv4 = true
#  expect_default_ipv6 = false // default value is "false"
}
