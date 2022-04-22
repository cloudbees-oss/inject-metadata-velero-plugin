A [restore plugin for Velero](https://velero.io/docs/v1.8/custom-plugins/#plugin-kinds)
which mutates all restored `StatefulSet`s,
adding (or replacing) the environment variable `RESTORED_FROM_BACKUP` in each container.
The value will be the identifier of the Velero restore object.

This is useful for informing a stateful application that it has been freshly restored from backup.
It might then take special steps to recover from partly written files, etc.
CloudBees CI [honors this variable](https://docs.cloudbees.com/docs/admin-resources/latest/pipelines/controlling-builds#_restarting_builds_after_a_restore).

Usage:

```yaml
initContainers:
# your existing plugins, then add:
- name: inject-metadata-velero-plugin
  image: ghcr.io/cloudbees-oss/inject-metadata-velero-plugin:main
  imagePullPolicy: Always
  volumeMounts:
  - mountPath: /target
    name: plugins
```

Binaries are [signed with Sigstore](https://docs.sigstore.dev/cosign/openid_signing).

[Kubernetes feature #48180](https://github.com/kubernetes/kubernetes/issues/48180), if implemented, would offer an alternate approach.
