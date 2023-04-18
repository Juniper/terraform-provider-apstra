# Apstra Lab Guide Demo
This directory contains an example project which follows the [Apstra Lab Guide](https://cloudlabs.apstra.com/labguide/Cloudlabs/4.1.2/lab1-junos/lab1-junos-0_intro.html)
currently published with the v4.1.1. Apstra CloudLabs "Juniper Customer Lab".

### Launch a CloudLabs Instance
This demo is tested only against Apstra 4.1.1. The correct CloudLabs template is
*Juniper Customer Lab* with *Apstra Version: AOS_4.1.1_OB* on the CloudLabs
"Experimental" tab. The Terraform plugin works with Apstra 4.1.2, but some of
the baked-in object names (logical devices, interface maps) changed between
revisions of the CloudLabs topologies, so it's smoother sailing with the 4.1.1
revision of the lab topology.

### Install the Provider
Refer to the project's [main README](../README.md) to get the provider installed
on your system.

### Copy the files in this directory to your local system
This might be the easiest way:
```shell
git clone https://github.com/Juniper/terraform-provider-apstra.git
cd terraform-provider-apstra/lab_guide_demo
```

### Configure the Provider
Notes in [0_provider.tf](0_provider.tf) detail configuring the provider to talk
to your CloudLabs instance.

### Work through the files in numerical order
Each terraform configuration file after `0_provider.tf` is 100% commented
out. Work through the files in order, un-commenting one `resource` or
`data`(source) at a time. Compare the results with the lab guide and with the
Apstra web UI.