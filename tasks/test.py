
from invoke import task

from .constants import DRIVER_NAME, GCSFUSE_VERSION, IMAGE_DEV, IMAGE_LATEST, VERSION
from .image import image_name

@task(
    help={
        'version': f'The desired version (default: {VERSION})',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    },
    default=True,
)
def all(ctx, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    unit(ctx, version, gcsfuse)
    sanity(ctx, version, gcsfuse)

def test_image_name(version=VERSION):
    image = image_name(version)
    image += '-builder'
    return image

def build(ctx, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    global_ldflags = ''
    version += '-rc'

    image = test_image_name(version)

    ctx.run(
        f'docker build . --tag {image} '
        f'--build-arg version={version} '
        f'--build-arg global_ldflags="{global_ldflags}" '
        f'--build-arg gcsfuse_version="{gcsfuse}" '
        f'--target test',
        echo=True,
    )

@task(
    help={
        'version': f'The desired version (default: {VERSION})',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    }
)
def sanity(ctx, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    build(ctx, version, gcsfuse)

    version += '-rc'

    image = test_image_name(version)

    ctx.run(
        f'docker run '
        f'--rm '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-t {image} '
        f'go test ./test',
        echo=True
    )

@task(
    help={
        'version': f'The desired version (default: {VERSION})',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    }
)
def unit(ctx, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    build(ctx, version, gcsfuse)

    version += '-rc'

    image = test_image_name(version)

    ctx.run(
        f'docker run '
        f'--rm '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-t {image} '
        f'go test ./pkg/driver',
        echo=True
    )

    ctx.run(
        f'docker run '
        f'--rm '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-t {image} '
        f'go test ./pkg/flags',
        echo=True
    )

    ctx.run(
        f'docker run '
        f'--rm '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-t {image} '
        f'go test ./pkg/util',
        echo=True
    )