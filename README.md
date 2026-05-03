wordpress-operator
===
A Kubernetes operator for managing WordPress deployments at scale.

## Goals and status

The main goals of the operator are:

1. Easily deploy scalable WordPress sites on top of Kubernetes
2. Allow best practices for en masse upgrades (canary, slow rollout, etc.)
3. Friendly to devops (monitoring, availability, scalability and backup stories solved)

This is a fork of the original [bitpoke/wordpress-operator](https://github.com/bitpoke/wordpress-operator), maintained independently and updated for current Kubernetes versions. Releases are at [MatthewKennedy/wordpress-operator/releases](https://github.com/MatthewKennedy/wordpress-operator/releases).

The minimum supported Kubernetes version is 1.27.

## Components

1. **wordpress-operator** — this project. Image: `public.ecr.aws/w0a8g6c1/wordpress-operator`.
2. **WordPress runtime** — container image (nginx + PHP-FPM) the operator deploys. Image: `ghcr.io/matthewkennedy/wordpress-runtime`. Source: [MatthewKennedy/stack-runtimes](https://github.com/MatthewKennedy/stack-runtimes).

## Deploy

### Install CRDs

Helm installs the CRD by default. To install it manually first:

```shell
kubectl apply --server-side -f https://raw.githubusercontent.com/MatthewKennedy/wordpress-operator/main/config/crd/bases/wordpress.presslabs.org_wordpresses.yaml
```

`--server-side` is required because the generated CRD exceeds the 256 KB `last-applied-configuration` annotation limit.

### Install the operator

The chart is published as an OCI artifact:

```shell
helm install wordpress-operator oci://public.ecr.aws/w0a8g6c1/wordpress-operator \
  --version 2.0.1 \
  --namespace wordpress-operator \
  --create-namespace
```

## Routing

The operator manages the WordPress `Deployment`, `Service`, `Secret`, and `PersistentVolumeClaim`s. Routing and TLS termination are intentionally **not** managed by the operator — point your Gateway API `HTTPRoute`, ingress controller, or reverse proxy at the generated `Service` (port `80` → container `8080`).

## Deploying a WordPress Site

```yaml
apiVersion: wordpress.presslabs.org/v1alpha1
kind: Wordpress
metadata:
  name: mysite
spec:
  replicas: 3
  routes:
    - domain: example.com
  # image: ghcr.io/matthewkennedy/wordpress-runtime
  # tag: latest
  code: # where to find the code
    # contentSubpath: wp-content/
    # by default, code gets an empty dir. Can be one of the following:
    git:
      repository: https://github.com/example.com
      # reference: master
      # env:
      #   - name: SSH_RSA_PRIVATE_KEY
      #     valueFrom:
      #       secretKeyRef:
      #         name: mysite
      #         key: id_rsa

    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {} (default)

  media: # where to find the media files
    # by default, media gets an empty dir. Can be one of the following:
    gcs: # store files using Google Cloud Storage
      bucket: my-wordpress-media
      prefix: mysite/
      env:
        - name: GOOGLE_CREDENTIALS
          valueFrom:
            secretKeyRef:
              name: mysite
              key: google_application_credentials.json
        - name: GOOGLE_PROJECT_ID
          value: development
    # persistentVolumeClaim: {}
    # hostPath: {}
    # emptyDir: {}
  bootstrap: # wordpress install config
    env:
      - name: WORDPRESS_BOOTSTRAP_USER
        valueFrom:
          secretKeyRef:
            name: mysite
            key: USER
      - name: WORDPRESS_BOOTSTRAP_PASSWORD
        valueFrom:
          secretKeyRef:
            name: mysite
            key: PASSWORD
      - name: WORDPRESS_BOOTSTRAP_EMAIL
        valueFrom:
          secretKeyRef:
            name: mysite
            key: EMAIL
      - name: WORDPRESS_BOOTSTRAP_TITLE
        valueFrom:
          secretKeyRef:
            name: mysite
            key: TITLE
  # extra volumes for the WordPress container
  volumes: []
  # extra volume mounts for the WordPress container
  volumeMounts: []
  # extra env variables for the WordPress container
  env:
    - name: DB_HOST
      value: mysite-mysql
    - name: DB_USER
      valueFrom:
        secretKeyRef:
          name: mysite-mysql
          key: USER
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: mysite-mysql
          key: PASSWORD
    - name: DB_NAME
      valueFrom:
        secretKeyRef:
          name: mysite-mysql
          key: DATABASE
  envFrom: []
```

## License

This project is licensed under Apache 2.0 license. Read the [LICENSE](LICENSE) file in the
top distribution directory, for the full license text.
