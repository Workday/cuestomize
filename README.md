# CUEstomize
CUEstomize is a tool that allows you to generate Kubernetes manifests to be consumed by Kustomize using CUE-Lang.

It does so by implementing a Kustomize KRM function that reads CUE files from a directory and combines them in a standard way with the Kustomize input.

This repo is generic over the CUE model, and provides a Docker image that embeds the KRM function binary.
This allows it to be used with any CUE model following a compatible convention for Input and Outputs naming.

## Usage
CUEstomize is generic, and out-of-the-box it does not provide any CUE model.
To use it, you need to write your own CUE model and bake an image that embeds the CUE model and the KRM function binary (you can use cuestomize image as a base image).

## Local Development
Build image locally.

Run:
```bash
# kustomize build <PATH_TO_KUSTOMIZATION_FILE> --enable-alpha-plugins
# Example:
kustomize build . --enable-alpha-plugins
```
