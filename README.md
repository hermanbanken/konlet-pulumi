# Konlet: cheapest way to continuously run Docker image on GCP
This Pulumi setup deploys a Docker image to GCP. It does so in the cheapest way possible: using e2-micro instances and [preemptible](https://cloud.google.com/compute/docs/instances/preemptible) instances ([60-91% discount](https://cloud.google.com/compute/docs/instances/spot#pricing)). This means you can have a running instance 24/7 for around 3$/month.

The setup uses the [Konlet feature](https://github.com/GoogleCloudPlatform/konlet) of the host VM OS (cos-cloud).

### Alternatives
In principle, this is much like [running the following commands](https://cloud.google.com/compute/docs/containers/deploying-containers). However, those commands can not easily be automated. This repository provides a template for how to automate them using Pulumi.

```bash
export DOCKER_IMAGE=gcr.io/cloud-marketplace/google/nginx1:15

# Single instance
gcloud compute instances create-with-container VM_NAME \
    --container-image $DOCKER_IMAGE

# or with a MIG (https://cloud.google.com/compute/docs/containers/deploying-containers#managedinstancegroupcontainer) template
gcloud compute instance-templates create-with-container nginx-template \
    --container-image $DOCKER_IMAGE
gcloud compute instance-groups managed create example-group \
    --base-instance-name nginx-vm --size 1 --template nginx-template
```
