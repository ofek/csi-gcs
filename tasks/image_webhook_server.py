from invoke import task

from .utils import image_name_webhook_server, image_tags

@task(
    help={
        'release': 'Build a webhook server release image',
    },
    default=True,
)
def build(ctx, release=False):
    if release:
        global_ldflags = '-s -w'
        docker_build_args = '--no-cache'
    else:
        global_ldflags = ''
        docker_build_args = ''

    image = image_name_webhook_server()

    ctx.run(
        f'docker build . --tag {image} '
        f'-f webhook.Dockerfile '
        f'--build-arg global_ldflags="{global_ldflags}" '
        f'{docker_build_args}',
        echo=True,
    )

    for tag in image_tags():
        ctx.run(f'docker tag {image} {image_name_webhook_server(tag)}', echo=True)

@task
def deploy(ctx):
    ctx.run(f'docker push {image_name_webhook_server()}', echo=True)
    for tag in image_tags():
        if tag != 'dev':
            ctx.run(f'docker tag {image_name_webhook_server()} {image_name_webhook_server(tag)}', echo=True)
            ctx.run(f'docker push {image_name_webhook_server(tag)}', echo=True)
