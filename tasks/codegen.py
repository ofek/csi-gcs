from invoke import task

from .utils import get_root


@task(
    default=True,
)
def codegen(ctx):
    mount_dir = '/go/src/github.com/ofek/csi-gcs'
    ctx.run(
        f'docker run '
        f'--rm '
        f'-v "{get_root()}:{mount_dir}" '
        f'-w {mount_dir} '
        f'golang:1.18.2-alpine3.15 '
        f'./hack/update-codegen.sh',
        echo=True,
    )
