from invoke import task

from .constants import GCSFUSE_VERSION
from .utils import get_version, image_name, image_tags


@task(
    help={
        'release': 'Build a release image',
        'compress': 'Minimize image size',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    },
    default=True,
)
def build(ctx, release=False, compress=False, gcsfuse=GCSFUSE_VERSION):
    if release:
        global_ldflags = '-s -w'
        docker_build_args = '--no-cache'
    else:
        global_ldflags = ''
        docker_build_args = ''

    upx_flags = '--best --ultra-brute' if compress else ''
    image = image_name(None if release else 'dev')

    ctx.run(
        f'docker build . --tag "{image}" '
        f'--build-arg version="{get_version()}" '
        f'--build-arg global_ldflags="{global_ldflags}" '
        f'--build-arg gcsfuse_version="{gcsfuse}" '
        f'--build-arg upx_flags="{upx_flags}" '
        f'{docker_build_args}',
        echo=True,
    )

    if release:
        for tag in image_tags():
            ctx.run(f'docker tag {image} {image_name(tag)}', echo=True)


@task
def deploy(ctx):
    ctx.run(f'docker push {image_name()}', echo=True)
    for tag in image_tags():
        if tag != 'dev':
            ctx.run(f'docker tag {image_name()} {image_name(tag)}', echo=True)
            ctx.run(f'docker push {image_name(tag)}', echo=True)
