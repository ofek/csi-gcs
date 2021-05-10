from invoke import task

from .utils import EnvVars, get_version


@task(
    help={
        'release': 'Build the webhook server binary',
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
            f'-o bin/webhook '
            f'-ldflags "all={global_ldflags}" '
            f'./cmd/webhook',
            echo=True,
        )
