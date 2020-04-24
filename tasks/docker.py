import os
from invoke import Exit, task

from .constants import DRIVER_NAME, GCSFUSE_VERSION

@task
def build_environment(ctx):
  ctx.run(
    f'docker volume create csi-gcs-go',
    echo=True,
  )
  ctx.run(
    f'docker build . '
    f'--tag {DRIVER_NAME}/env '
    f'-f dev-env.Dockerfile '
    f'--build-arg gcsfuse_version="{GCSFUSE_VERSION}"',
    echo=True,
  )

@task(
  pre=[build_environment],
  default=True,
)
def run(ctx, command="echo Provide Command"):
  pwd = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

  ctx.run(
    f'docker run '
    f'--rm '
    f'-v {pwd}:/driver '
    f'-v csi-gcs-go:/go '
    f'-p 8765:8765 '
    f'--cap-add SYS_ADMIN '
    f'--device /dev/fuse '
    f'--privileged '
    f'-v /tmp/csi:/tmp/csi:rw '
    f'-v /var/run/docker.sock:/var/run/docker.sock '
    f'{DRIVER_NAME}/env '
    f'{command}',
    echo=True,
  )
