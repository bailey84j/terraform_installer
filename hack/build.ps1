

GIT_TAG="${BUILD_VERSION:-$(git describe --always --abbrev=40 --dirty)}"

$GOFLAGS= "--mod=vendor"
$LDFLAGS="$($LDFLAGS) -X github.com/bailey84j/terraform_installer/pkg/version.Raw=${GIT_TAG} -X github.com/openshift/installer/pkg/version.Commit=${GIT_COMMIT} -X github.com/openshift/installer/pkg/version.defaultArch=${DEFAULT_ARCH}"
$TAGS="-"
$OUTPUT = "bin/openshift-install"

# shellcheck disable=SC2086
& go build -o "$($OUTPUT)" ./cmd/terraform-install