package main

import (
	"fmt"
	"pulumi/internal"

	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v5/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func stackExample(ctx *pulumi.Context) (err error) {
	project := config.New(ctx, "gcp").Get("project")
	region := config.New(ctx, "gcp").Get("region")
	zone := config.New(ctx, "gcp").Get("zone")

	var prov *gcp.Provider
	if prov, err = gcp.NewProvider(ctx, "proj", &gcp.ProviderArgs{Project: pulumi.String(project)}); err != nil {
		return err
	}

	image := config.New(ctx, "docker").Get("image")
	serviceAccountName := "myapp"
	someEnv := "FOOBAR"

	// Create a VM that does 1 thing: run our Docker image
	err = internal.ManagedContainer(ctx, prov, "mycontainer", exampleDecl(image, someEnv), zone, exampleInstanceArgs(project, serviceAccountName))
	if err != nil {
		return err
	}
	// Setup NAT: the instance has no IP and without NAT no outgoing traffic is possible
	err = internal.AddNAT(ctx, prov, "myapp", region, 64514)
	if err != nil {
		return err
	}
	return nil
}

// exampleInstanceArgs sets some overrides on top of internal.DefaultCosInstanceTemplate
func exampleInstanceArgs(project, serviceAccountName string) func(ita *compute.InstanceTemplateArgs) {
	return func(ita *compute.InstanceTemplateArgs) {
		sa := ita.ServiceAccount.(*compute.InstanceTemplateServiceAccountArgs)
		sa.Email = pulumi.String(fmt.Sprintf("%s@%s.iam.gserviceaccount.com", serviceAccountName, project))

		scopes := sa.Scopes.(pulumi.StringArray)
		scopes = append(scopes, pulumi.String("https://www.googleapis.com/auth/cloud-platform")) // SecretManager
		sa.Scopes = scopes
	}
}

func exampleDecl(image string, someEnv string) internal.KonletDeclaration {
	decl := internal.KonletDeclaration{
		Spec: internal.KonletDeclSpec{
			Containers: []internal.KonletDeclContainer{{
				Image:           image,
				Name:            "container",
				SecurityContext: internal.KonletSecurityContext{Privileged: false},
				Stdin:           false,
				Tty:             false,
				VolumeMounts:    []struct{}{},
				Env: []internal.KonletDeclEnv{
					// Some application configuration
					{Name: "MY_CUSTOM_ENV", Value: someEnv},
				},
			}},
			RestartPolicy: "Always",
			Volumes:       []struct{}{},
		},
	}

	// Optional: load some secret from Secret Manager
	// using gcp-get-secret (https://github.com/binxio/gcp-get-secret)
	if 1 < 0 {
		decl.Spec.Containers[0].Env = append(decl.Spec.Containers[0].Env,
			// Downloads the content from secret "my-secret" into "/env.sh"
			internal.KonletDeclEnv{Name: "MY_SECRET", Value: "gcp:///my-secret?destination=/env.sh"},
		)
		// Run gcp-get-secret to provision the secrets in /env.sh
		decl.Spec.Containers[0].Command = []string{"/bin/gcp-get-secret"}
		// TODO tweak "/bin/server" to match your CMD
		decl.Spec.Containers[0].Args = []string{"-use-default-credentials", "-verbose", "/bin/sh", "-c", "source /env.sh && /bin/server"}
	}

	return decl
}
