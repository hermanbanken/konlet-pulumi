package internal

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DefaultCosInstanceTemplate(decl KonletDeclaration) *compute.InstanceTemplateArgs {
	return &compute.InstanceTemplateArgs{
		MachineType: pulumi.String("e2-micro"),
		Disks: &compute.InstanceTemplateDiskArray{&compute.InstanceTemplateDiskArgs{
			DiskSizeGb:  pulumi.Int(10), // Gb
			SourceImage: pulumi.String("https://www.googleapis.com/compute/v1/projects/cos-cloud/global/images/family/cos-stable"),
		}},
		NetworkInterfaces: &compute.InstanceTemplateNetworkInterfaceArray{&compute.InstanceTemplateNetworkInterfaceArgs{
			Network: pulumi.String("default"),
		}},
		ServiceAccount: &compute.InstanceTemplateServiceAccountArgs{
			Scopes: pulumi.ToStringArray([]string{
				"https://www.googleapis.com/auth/devstorage.read_only",
				"https://www.googleapis.com/auth/logging.write",
				"https://www.googleapis.com/auth/service.management.readonly",
				"https://www.googleapis.com/auth/servicecontrol",
				"https://www.googleapis.com/auth/trace.append",
				"https://www.googleapis.com/auth/monitoring.write",
			}),
		},
		Metadata: pulumi.Map{
			"gce-container-declaration": pulumi.String(ContainerDeclaration(decl)),
			// Enables logging (https://github.com/GoogleCloudPlatform/konlet/issues/56#issuecomment-468659630)
			// Reading logs is documented here: https://cloud.google.com/compute/docs/containers/deploying-containers#viewing_container_logs
			"google-logging-enabled": pulumi.String("true"),
			// Do block ssh keys, so no-one can hijack the running containers
			"block-project-ssh-keys": pulumi.String("true"),
		},
		Scheduling: &compute.InstanceTemplateSchedulingArgs{
			AutomaticRestart:  pulumi.Bool(false),
			OnHostMaintenance: pulumi.String("TERMINATE"),
			Preemptible:       pulumi.Bool(true),
		},
	}
}

func ManagedContainer(ctx *pulumi.Context, prov pulumi.ProviderResource, name string, decl KonletDeclaration, zone string, overrideFn func(*compute.InstanceTemplateArgs)) error {
	args := DefaultCosInstanceTemplate(decl)
	overrideFn(args)
	template, err := compute.NewInstanceTemplate(ctx, name, args, pulumi.Provider(prov))
	if err != nil {
		return err
	}

	_, err = compute.NewInstanceGroupManager(ctx, name, &compute.InstanceGroupManagerArgs{
		Name:             pulumi.String(name),
		Zone:             pulumi.String(zone),
		TargetSize:       pulumi.Int(1),
		BaseInstanceName: pulumi.String(name),
		Versions: compute.InstanceGroupManagerVersionArray{
			&compute.InstanceGroupManagerVersionArgs{
				Name:             template.Name,
				InstanceTemplate: template.SelfLink,
			},
		},
		UpdatePolicy: compute.InstanceGroupManagerUpdatePolicyArgs{
			Type:                  pulumi.String("PROACTIVE"),
			MinimalAction:         pulumi.String("REPLACE"),
			MaxUnavailablePercent: pulumi.Int(100),
			MaxSurgePercent:       pulumi.Int(100),
		},
	}, pulumi.Provider(prov), pulumi.DependsOn([]pulumi.Resource{template}))
	if err != nil {
		return err
	}

	return nil
}

// AddNAT configures a NAT router
// https://cloud.google.com/network-connectivity/docs/router/how-to/creating-routers
// > Google ASN: The private ASN (64512-65534, 4200000000-4294967294) for the Cloud
// > Router that you are configuring; it can be any private ASN that you aren't already
// > using as a peer ASN in the same region and network—for example, 65001.
// > Cloud Router requires you to use a private ASN, but your on-premises ASN can be public or private.
func AddNAT(ctx *pulumi.Context, prov pulumi.ProviderResource, name string, region string, asn int) error {
	routerName := fmt.Sprintf("router-%s", name)
	natName := fmt.Sprintf("nat-%s", name)

	router, err := compute.NewRouter(ctx, routerName, &compute.RouterArgs{
		Region:  pulumi.String(region),
		Network: pulumi.String("default"),
		Bgp: &compute.RouterBgpArgs{
			Asn: pulumi.Int(asn),
		},
	})
	if err != nil {
		return err
	}
	_, err = compute.NewRouterNat(ctx, natName, &compute.RouterNatArgs{
		Router:                        router.Name,
		Region:                        router.Region,
		NatIpAllocateOption:           pulumi.String("AUTO_ONLY"),
		SourceSubnetworkIpRangesToNat: pulumi.String("ALL_SUBNETWORKS_ALL_IP_RANGES"),
		LogConfig: &compute.RouterNatLogConfigArgs{
			Enable: pulumi.Bool(true),
			Filter: pulumi.String("ERRORS_ONLY"),
		},
	})
	if err != nil {
		return err
	}
	return nil
}
