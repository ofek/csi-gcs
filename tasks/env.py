from invoke import task

from .constants import DRIVER_NAME, GCSFUSE_VERSION
from .utils import get_root


@task
def create(ctx):
    ctx.run(
        'docker volume create csi-gcs-go',
        echo=True,
    )
    ctx.run(
        f'docker build . '
        f'--tag {DRIVER_NAME}-env '
        f'-f dev-env.Dockerfile '
        f'--build-arg gcsfuse_version="{GCSFUSE_VERSION}"',
        echo=True,
    )


@task
def delete(ctx):
    ctx.run(
        'docker volume rm csi-gcs-go',
        echo=True,
    )
    ctx.run(
        f'docker image rm {DRIVER_NAME}-env',
        echo=True,
    )


@task(
    pre=[create],
    default=True,
)
def run(ctx, command='echo Provide Command'):
    ctx.run(
        f'docker run '
        f'--rm '
        f'-v {get_root()}:/driver '
        f'-v csi-gcs-go:/go '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-v /tmp/csi:/tmp/csi:rw '
        f'-v /var/run/docker.sock:/var/run/docker.sock '
        f'{DRIVER_NAME}-env '
        f'{command}',
        echo=True,
    )
