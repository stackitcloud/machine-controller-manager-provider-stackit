# Local Development Guide

This guide describes how to run the STACKIT MCM provider against a real Gardener shoot on STACKIT. A local kind cluster is not suitable because the provider creates real STACKIT VMs that must join a real Kubernetes cluster.

## Prerequisites

Before you begin, ensure you have the following:

- A Gardener installation on STACKIT
- A shoot cluster created in that Gardener
- Access to the seed and shoot via `gardenctl`
- A local Go toolchain

## Overview

You will run the provider and MCM locally while pointing them at real clusters:

- `$TARGET_KUBECONFIG` points to the cluster where you want machines to join (the shoot).
- `$CONTROL_KUBECONFIG` points to the cluster that stores Machine objects (the seed).
- `$CONTROL_NAMESPACE` is where MCM watches Machine objects (usually `shoot--projectname--shootname` in the seed).

MachineClass objects and Secrets are assumed to already exist. Running the provider locally results in faster feedback loops when developing new features and enables using the go debugger.

## 1. Get kubeconfigs with gardenctl

Assume:

- Shoot name: `foobar`
- Seed name: `foobar-seed`

Export kubeconfigs using `gardenctl`:

```bash
# Target (shoot) kubeconfig
gardenctl kubeconfig --raw --shoot foobar > /tmp/target.kubeconfig

# Control (seed) kubeconfig
gardenctl kubeconfig --raw --seed foobar-seed > /tmp/control.kubeconfig
```

Set these environment variables on every terminal used:

```bash
export TARGET_KUBECONFIG=/tmp/target.kubeconfig
export CONTROL_KUBECONFIG=/tmp/control.kubeconfig
export CONTROL_NAMESPACE=shoot--testing--foobar
```

## 2. Scale down the in-cluster MCM

Scale the existing MCM in the seed to 0 so your local controller can take over:

```bash
kubectl --kubeconfig "$CONTROL_KUBECONFIG" -n "$CONTROL_NAMESPACE" scale deployment/machine-controller-manager --replicas=0
```

Since Gardener periodically scales the deployment back up, you can use a watch command in a separate terminal to continuously scale it down:

```bash
watch -n 5 "kubectl --kubeconfig '$CONTROL_KUBECONFIG' -n '$CONTROL_NAMESPACE' scale deployment/machine-controller-manager --replicas=0"
```

This will check and scale down the deployment every 5 seconds.  
Make sure to set the environment varialbes.

## 3. Run the provider (driver)

On another terminal in your provider repo:

If you are running against QA, export these environment variables before starting:

```bash
export STACKIT_TOKEN_BASEURL="https://service-account.api.qa.stackit.cloud/token"
export STACKIT_IAAS_ENDPOINT="https://iaas.api.qa.stackit.cloud"
```

```bash
make start
```

Make sure to set the environment varialbes.

## 4. Run MCM locally

On another terminal in the Gardener MCM repository:

```bash
git clone git@github.com:gardener/machine-controller-manager.git
cd machine-controller-manager
```

Run MCM:

```bash
make start
```

Make sure to set the environment varialbes.

## Notes

- This workflow assumes MachineClass and Secret objects already exist and are valid.
- The local controllers should reconcile existing resources and provision STACKIT VMs that join the shoot.
- Re-enable the in-cluster MCM after testing by scaling it back up.
