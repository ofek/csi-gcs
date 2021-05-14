from invoke import task

@task(
    help={
        'namespace': 'Namespace of the webhook service.',
        'service': 'Name of the webhook service.',
    },
    default=True,
)
def build(ctx, namespace="kube-system", service="csi-gcs-webhook-server", output="deploy/overlays/dev/secrets"):
    ctx.run(
      f'openssl req '
      f'-new -newkey rsa:2048 -days 365 -nodes -x509 '
      f'-subj "/CN={service}.{namespace}.svc" '
      f'-keyout {output}/tls.key -out {output}/tls.crt',
      echo=True
    )
