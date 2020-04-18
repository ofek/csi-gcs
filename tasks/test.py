
from invoke import task

from .constants import DRIVER_NAME, GCSFUSE_VERSION, IMAGE_DEV, IMAGE_LATEST, VERSION
from .image import image_name

@task(
    help={
        'version': f'The desired version (default: {VERSION})',
        'gcsfuse': f'The version or commit hash of gcsfuse (default: {GCSFUSE_VERSION})',
    }
)
def sanity(ctx, version=VERSION, gcsfuse=GCSFUSE_VERSION):
    global_ldflags = ''
    version += '-rc'

    image = image_name(version)
    image += '-builder'

    ctx.run(
        f'docker build . --tag {image} '
        f'--build-arg version={version} '
        f'--build-arg global_ldflags="{global_ldflags}" '
        f'--build-arg gcsfuse_version="{gcsfuse}" '
        f'--target test',
        echo=True,
    )

    ctx.run(
        f'docker run '
        f'--rm '
        f'--cap-add SYS_ADMIN '
        f'--device /dev/fuse '
        f'--privileged '
        f'-t {image} '
        f'go test',
        echo=True
    )