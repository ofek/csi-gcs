
from invoke import task

@task
def sanity(ctx):
    ctx.run(f'go test ./test', echo=True)

@task
def unit_driver(ctx):
    ctx.run(f'go test ./pkg/driver', echo=True)

@task
def unit_flags(ctx):
    ctx.run(f'go test ./pkg/flags', echo=True)

@task
def unit_util(ctx):
    ctx.run(f'go test ./pkg/util', echo=True)

@task(pre=[unit_flags, unit_driver, unit_util])
def unit(ctx): pass

@task(
    pre=[unit, sanity],
    default=True,
)
def all(ctx): pass