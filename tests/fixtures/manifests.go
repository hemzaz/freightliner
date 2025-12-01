package fixtures

// TestManifests provides sample Docker manifest data for testing

const (
	// SimpleManifestV2 is a basic Docker Image Manifest V2, Schema 2
	SimpleManifestV2 = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "size": 1469,
    "digest": "sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b"
  },
  "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 2479,
      "digest": "sha256:f1b5933fe4b5f49bbe8258745cf396afe07e625bdab3168e364daf7c956b6b81"
    }
  ]
}`

	// MultiLayerManifest is a manifest with multiple layers
	MultiLayerManifest = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "size": 2500,
    "digest": "sha256:abc123def456789012345678901234567890123456789012345678901234567"
  },
  "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 5242880,
      "digest": "sha256:layer1111111111111111111111111111111111111111111111111111111111"
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 10485760,
      "digest": "sha256:layer2222222222222222222222222222222222222222222222222222222222"
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 20971520,
      "digest": "sha256:layer3333333333333333333333333333333333333333333333333333333333"
    }
  ]
}`

	// OCIManifest is an OCI Image Manifest
	OCIManifest = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 7023,
    "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  },
  "layers": [
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 32654,
      "digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
    },
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 16724,
      "digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
    }
  ]
}`

	// ManifestList is a Docker Manifest List (multi-arch)
	ManifestList = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1234,
      "digest": "sha256:amd64manifest111111111111111111111111111111111111111111111111",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 1235,
      "digest": "sha256:arm64manifest222222222222222222222222222222222222222222222222",
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    }
  ]
}`
)

// TestImageConfigs provides sample Docker config files
const (
	// SimpleConfig is a basic Docker image config
	SimpleConfig = `{
  "architecture": "amd64",
  "config": {
    "Hostname": "",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/bin/sh"
    ],
    "Image": "sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": null
  },
  "container": "abc123",
  "container_config": {
    "Hostname": "abc123",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/bin/sh",
      "-c",
      "#(nop) ",
      "CMD [\"/bin/sh\"]"
    ],
    "Image": "sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": {}
  },
  "created": "2023-01-01T00:00:00.000000000Z",
  "docker_version": "20.10.12",
  "history": [
    {
      "created": "2023-01-01T00:00:00.000000000Z",
      "created_by": "/bin/sh -c #(nop) ADD file:abc123 in / "
    },
    {
      "created": "2023-01-01T00:00:01.000000000Z",
      "created_by": "/bin/sh -c #(nop)  CMD [\"/bin/sh\"]",
      "empty_layer": true
    }
  ],
  "os": "linux",
  "rootfs": {
    "type": "layers",
    "diff_ids": [
      "sha256:f1b5933fe4b5f49bbe8258745cf396afe07e625bdab3168e364daf7c956b6b81"
    ]
  }
}`
)

// TestRegistryResponses provides sample registry API responses
const (
	// TagListResponse is a sample response from /v2/{name}/tags/list
	TagListResponse = `{
  "name": "test/app",
  "tags": [
    "v1.0.0",
    "v1.1.0",
    "v1.2.0",
    "latest",
    "dev"
  ]
}`

	// RepositoryListResponse is a sample response from /v2/_catalog
	RepositoryListResponse = `{
  "repositories": [
    "test/app",
    "test/service-a",
    "test/service-b",
    "prod/app",
    "prod/service"
  ]
}`

	// ErrorResponse is a sample error response from registry
	ErrorResponse = `{
  "errors": [
    {
      "code": "MANIFEST_UNKNOWN",
      "message": "manifest unknown",
      "detail": {
        "Tag": "nonexistent"
      }
    }
  ]
}`

	// UnauthorizedResponse is a sample 401 response
	UnauthorizedResponse = `{
  "errors": [
    {
      "code": "UNAUTHORIZED",
      "message": "authentication required",
      "detail": null
    }
  ]
}`
)
