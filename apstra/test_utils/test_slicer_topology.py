#!/usr/bin/env python3
"""
Script to create a Slicer topology, reserve, deploy, undeploy, unreserve and delete it based on topology_spec.yaml
"""

import sys
import yaml
from pathlib import Path
from systest.cli.topology import topology_util
from systest.web.slicer_client import SlicerClient
from systest.lib.model_json import topology_spec_from_json, get_testbed_deploy_spec_from_json



def main():
    """Main function to create, reserve, deploy, undeploy, unreserve and delete the topology."""
    print("=" * 70)
    print("Slicer Topology Deployment")
    print("=" * 70)

    try:
        # Initialize Slicer client
        slicer_server_url = "https://api.slicer.dc1.apstra.com"
        print(f"\nConnecting to Slicer server: {slicer_server_url}")
        slicer_client = SlicerClient(slicer_server_url=slicer_server_url)
        print("✓ Successfully connected to Slicer server")

        # Load topology specification
        print("\n" + "-" * 70)
        # Get the directory where this script is located
        script_dir = Path(__file__).parent.resolve()
        topology_spec_path = script_dir / "topology_spec.yaml"
        print(f"Using topology spec: {topology_spec_path}")
        topology_spec = topology_util.get_topology_spec(str(topology_spec_path))
        print("✓ Topology specification loaded successfully")

        # # Print topology info
        # if 'topology_spec' in topology_spec:
        #     topo = topology_spec['topology_spec']
        #     print(f"\nTopology Configuration:")
        #     print(f"  - Use OVS: {topo.get('use_ovs', 'N/A')}")
        #     print(f"  - Use Patch Panel: {topo.get('use_patch_panel', 'N/A')}")
        #
        #     if 'duts' in topo and 'duts' in topo['duts']:
        #         print(f"  - Number of DUTs: {len(topo['duts']['duts'])}")
        #         for dut in topo['duts']['duts']:
        #             for dut_name, dut_config in dut.items():
        #                 print(f"    • {dut_name}: {dut_config.get('os_type', 'N/A')} "
        #                       f"({dut_config.get('impl_type', 'N/A')})")

        # if 'deploy_spec' in topology_spec:
        #     deploy = topology_spec['deploy_spec']
        #     print(f"\nDeploy Configuration:")
        #     for key, value in deploy.items():
        #         if isinstance(value, dict):
        #             branch = value.get('branch', 'N/A')
        #             build = value.get('build', 'N/A')
        #             print(f"  - {key}: {branch} ({build})")

        # Create topology
        print("\n" + "-" * 70)
        systest_name = "github_runner"
        owner = "github@apstra.com"
        print("Creating topology...")
        topology = slicer_client.create_systest(name=systest_name,
            topology_spec=topology_spec_from_json(topology_spec),
            owner=owner
        )
        print(f"✓ Topology created successfully")
        print(f"  Topology Name: {topology.name}")

        # Reserve topology
        print("\n" + "-" * 70)
        print("Reserving topology...")
        systest = slicer_client.reserve_resources(systest_name, owner=owner,
                                        duration=10000.0, immediate=True)
        print(f"✓ Topology reserved successfully")

        # Deploy topology
        print("\n" + "-" * 70)
        print("Deploying topology...")
        deployment_result = slicer_client.deploy_testbed(systest_name, get_testbed_deploy_spec_from_json(topology_spec),
                                     timeout=1200, wait=True)
        print(f"✓ Topology deployment initiated successfully")
        print(f"  Status: {deployment_result.deploy_status}")

        print("\n" + "=" * 70)
        print("Topology deployment completed successfully!")

        # Cleanup: Undeploy, Unreserve, and Delete
        print("\n" + "=" * 70)
        print("Starting cleanup process...")
        print("=" * 70)

        # Undeploy topology
        print("\n" + "-" * 70)
        print("Undeploying topology...")
        try:
            slicer_client.undeploy_testbed(systest_name, timeout=600, wait=True)
            print(f"✓ Topology undeployed successfully")
        except Exception as e:
            print(f"⚠ Warning: Failed to undeploy topology: {e}")

        # Unreserve topology
        print("\n" + "-" * 70)
        print("Unreserving topology...")
        try:
            slicer_client.release_resources(systest_name)
            print(f"✓ Topology unreserved successfully")
        except Exception as e:
            print(f"⚠ Warning: Failed to unreserve topology: {e}")

        # Delete topology
        print("\n" + "-" * 70)
        print("Deleting topology...")
        try:
            slicer_client.delete_systest(systest_name)
            print(f"✓ Topology deleted successfully")
        except Exception as e:
            print(f"⚠ Warning: Failed to delete topology: {e}")

        print("\n" + "=" * 70)
        print("Cleanup completed!")
        print("=" * 70)

        return 0

    except FileNotFoundError as e:
        print(f"\n❌ File Error: {e}")
        return 1

    except yaml.YAMLError as e:
        print(f"\n❌ YAML Parsing Error: {e}")
        return 1



if __name__ == "__main__":
    sys.exit(main())

