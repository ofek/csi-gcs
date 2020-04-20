from invoke import task

from .constants import DRIVER_NAME, GCSFUSE_VERSION, IMAGE_DEV, IMAGE_LATEST, VERSION


def image_name(version):
    return f'{DRIVER_NAME}:v{version}'


@task(
    help={
        'release': 'Build a release image',
        'compress': 'Minimize image size',
        'version': f'The desired version (default: {VERSION})',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    },
    default=True,
)
def build(ctx, release=False, compress=False, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    if release:
        static_image = IMAGE_LATEST
        global_ldflags = '-s -w'
        docker_build_args = '--no-cache'
    else:
        static_image = IMAGE_DEV
        global_ldflags = ''
        version += '-rc'
        docker_build_args = ''

    upx_flags = '--best --ultra-brute' if compress else ''
    image = image_name(version)

    ctx.run(
        f'docker build . --tag {image} '
        f'--build-arg version={version} '
        f'--build-arg global_ldflags="{global_ldflags}" '
        f'--build-arg gcsfuse_version="{gcsfuse}" '
        f'--build-arg upx_flags="{upx_flags}" '
        f'{docker_build_args}',
        echo=True,
    )

    ctx.run(f'docker tag {image} {static_image}', echo=True)


@task(
    help={
        'version': f'The desired version (default: {VERSION})',
    }
)
def push(ctx, version=VERSION):
    ctx.run(f'docker push {image_name(version)}', echo=True)
    ctx.run(f'docker push {IMAGE_LATEST}', echo=True)
