# Helm Chart for Devsecdb

[Devsecdb](https://secdb.khulnasoft.com) is a Database CI/CD tool for DevOps teams, built for Developers and DBAs.

## TL;DR

```bash
$ helm repo add devsecdb-repo https://khulnasoft.github.io/devsecdb
$ helm repo update
$ helm -n <YOUR_NAMESPACE> \
--set "devsecdb.option.port"={PORT} \
--set "devsecdb.option.externalPg.url"={PGDSN} \
--set "devsecdb.version"={VERSION} \
install <RELEASE_NAME> devsecdb-repo/devsecdb
```

## Prerequisites

- Kubernetes 1.24+
- Helm 3.9.0+

## Installing the Chart

```bash
$ helm -n <YOUR_NAMESPACE> \
--set "devsecdb.option.port"={PORT} \
--set "devsecdb.option.externalPg.url"={PGDSN} \
--set "devsecdb.version"={VERSION} \
--set "devsecdb.option.external-url"={EXTERNAL_URL} \
--set "devsecdb.persistence.enabled"={TRUE/FALSE} \
--set "devsecdb.persistence.storage"={STORAGE_SIZE} \
--set "devsecdb.persistence.storageClass"={STORAGE_CLASS} \
install <RELEASE_NAME> devsecdb-repo/devsecdb
```

For example:

```bash
$ helm -n devsecdb \
--set "devsecdb.option.port"=443 \
--set "devsecdb.option.externalPg.url"="postgresql://devsecdb:devsecdb@database.devsecdb.ap-east-1.rds.amazonaws.com/devsecdb" \
--set "devsecdb.option.external-url"="https://devsecdb.ngrok-free.app" \
--set "devsecdb.version"=2.11.1 \
--set "devsecdb.persistence.enabled"="true" \
--set "devsecdb.persistence.storage"="10Gi" \
--set "devsecdb.persistence.storageClass"="csi-disk" \
install devsecdb-release devsecdb-repo/devsecdb
```

## Uninstalling the Chart

```bash
helm delete --namespace <YOUR_NAMESPACE> <RELEASE_NAME>
```

## Upgrade Devsecdb Version/Configuration

Use `helm upgrade` command to upgrade the devsecdb version or configuration.

```bash
helm -n <YOUR_NAMESPACE> \
--set "devsecdb.option.port"={NEW_PORT} \
--set "devsecdb.option.externalPg.url"={NEW_PGDSN} \
--set "devsecdb.version"={NEW_VERSION} \
--set "devsecdb.option.external-url"={EXTERNAL_URL} \
--set "devsecdb.persistence.enabled"={TRUE/FALSE} \
--set "devsecdb.persistence.storage"={STORAGE_SIZE} \
--set "devsecdb.persistence.storageClass"={STORAGE_CLASS} \
upgrade devsecdb-release devsecdb-repo/devsecdb
```

## Parameters

|                        Parameter                         |                                                                                                                Description                                                                                                                 |                           Default Value                            |
| :------------------------------------------------------: | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: | :----------------------------------------------------------------: |
|                    `devsecdb.version`                    |                                                                                                  The version of Devsecdb to be installed.                                                                                                  |                              "2.11.1"                              |
|              `devsecdb.registryMirrorHost`               |                                                                              The host for the Docker registry mirror. Leave empty for default registry usage.                                                                              |                                 ""                                 |
|                  `devsecdb.option.port`                  |                                                                                                      Port where Devsecdb server runs.                                                                                                      |                                8080                                |
|                  `devsecdb.option.data`                  |                                                                                                  Data directory of Devsecdb data stored.                                                                                                   |                         /var/opt/devsecdb                          |
|              `devsecdb.option.external-url`              |                                              The address for users to visit Devsecdb, visit [our docs](https://www.secdb.khulnasoft.com/docs/get-started/install/external-url/) to get more details.                                               | "<https://www.secdb.khulnasoft.com/docs/get-started/install/external-url>" |
|             `devsecdb.option.disable-sample`             |                                                                                                        Disable the sample instance.                                                                                                        |                               false                                |
|             `devsecdb.option.externalPg.url`             |                                                                                        The PostgreSQL url(DSN) for Devsecdb to store the metadata.                                                                                         |                                 ""                                 |
|     `devsecdb.option.externalPg.existingPgURLSecret`     |                                                                           The name of Secret stores the PostgreSQL url(DSN) for Devsecdb to store the metadata.                                                                            |                                 ""                                 |
|   `devsecdb.option.externalPg.existingPgURLSecretKey`    |                                     The key of Secret stores the PostgreSQL url(DSN) for Devsecdb to store the metadata. Should be used with `devsecdb.option.externalPg.existingPgURLSecret` together.                                      |                                 ""                                 |
|           `devsecdb.option.externalPg.pgHost`            |                                                                                             The PostgreSQL host for Devsecdb metadata storage.                                                                                             |                               "host"                               |
|           `devsecdb.option.externalPg.pgPort`            |                                                                                             The PostgreSQL port for Devsecdb metadata storage.                                                                                             |                               "port"                               |
|         `devsecdb.option.externalPg.pgUsername`          |                                                                                           The PostgreSQL username for Devsecdb metadata storage.                                                                                           |                             "username"                             |
|         `devsecdb.option.externalPg.pgPassword`          |                                                                                           The PostgreSQL password for Devsecdb metadata storage.                                                                                           |                             "password"                             |
|         `devsecdb.option.externalPg.pgDatabase`          |                                                                                             The name of the PostgreSQL database for Devsecdb.                                                                                              |                             "database"                             |
|  `devsecdb.option.externalPg.existingPgPasswordSecret`   |                                                                       The name of Secret that stores the existing PostgreSQL password for Devsecdb metadata storage.                                                                       |                                 ""                                 |
| `devsecdb.option.externalPg.existingPgPasswordSecretKey` |                                                    The key of Secret storing the existing PostgreSQL password. Should be used with `devsecdb.option.externalPg.existingPgPasswordSecret`.                                                    |                                 ""                                 |
|       `devsecdb.option.externalPg.escapePassword`        | Controls whether to escape the password in the connection string. `devsecdb.option.externalPg.existingPgPasswordSecret` or `devsecdb.option.externalPg.pgPassword` should be specified with this value together. **Experimental feature.** |                               false                                |
|              `devsecdb.persistence.enabled`              |                                                                                                  Enable/disable persistence for Devsecdb.                                                                                                  |                               false                                |
|           `devsecdb.persistence.existingClaim`           |                                                                                    Name of the existing PersistentVolumeClaim for Devsecdb persistence.                                                                                    |                                 ""                                 |
|              `devsecdb.persistence.storage`              |                                                                                              Size of the persistent volume for Devsecdb data.                                                                                              |                               "2Gi"                                |
|           `devsecdb.persistence.storageClass`            |                                                                                         Storage class for the persistent volume used by Devsecdb.                                                                                          |                                 ""                                 |
|               `devsecdb.extraSecretMounts`               |                                                                               Additional Devsecdb secret mounts. Defined as an array of volumeMount objects.                                                                               |                                 []                                 |
|                 `devsecdb.extraVolumes`                  |                                                                                    Additional Devsecdb volumes. Defined as an array of volume objects.                                                                                     |                                 []                                 |

**If you enable devsecdb persistence, you should provide storageClass and storage to devsecdb to request a PVC, or provide the already existed PVC by existingClaim.**

## Need Help?

- Contact <support@secdb.khulnasoft.com>
- [Devsecdb Docs](https://secdb.khulnasoft.com/docs)
- [Devsecdb GitHub Issue Page](https://github.com/khulnasoft/devsecdb/issues/new/choose)
