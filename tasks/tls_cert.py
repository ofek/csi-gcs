from invoke import task

@task(
    help={
        'namespace': 'Namespace of the webhook service.',
        'service': 'Name of the webhook service.',
    },
    default=True,
)
def build(ctx, namespace="kube-system", service="csi-gcs-webhook-server", output="deploy/overlays/dev/secrets"):
    domain_name = f'{service}.{namespace}.svc'
    ctx.run(
      'openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 '
      f'-subj "/CN={domain_name}" '
      '-reqexts SAN '
      '-extensions SAN '
      f'-config <(cat /etc/ssl/openssl.cnf <(echo "\n[ SAN ]\nsubjectAltName=DNS:{domain_name}")) '
      f'-keyout {output}/tls.key -out {output}/tls.crt',
      echo=True
    )
