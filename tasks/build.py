from invoke import task

from .utils import EnvVars, get_version


@task(
    help={
        'release': 'Build a release image',
    },
    default=True,
)
def build(ctx, release=False):
    if release:
        global_ldflags = '-s -w'
    else:
        global_ldflags = ''

    with EnvVars({'CGO_ENABLED': '0', 'GOOS': 'linux', 'GOARCH': 'amd64'}):
        ctx.run(
            f'go build '
            f'-o bin/driver '
            f'-ldflags "all={global_ldflags}" '
            f'-ldflags "-X github.com/ofek/csi-gcs/pkg/driver.driverVersion={get_version()} {global_ldflags}" '
            f'./cmd',
            echo=True,
        )
