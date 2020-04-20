import os
import shutil
import webbrowser
from tempfile import TemporaryDirectory

from invoke import Exit, task

from .utils import create_file, get_git_email, get_git_user, get_latest_commit_hash


def insert_verbosity_flag(command, verbosity):
    # One level is no tox flag
    if verbosity:
        verbosity -= 1
    # By default hide deps stage and success text
    else:
        verbosity -= 2

    if verbosity < 0:
        command.insert(1, f"-{'q' * abs(verbosity)}")
    elif verbosity > 0:
        command.insert(1, f"-{'v' * abs(verbosity)}")


@task(
    help={'verbose': 'Increase verbosity (can be used additively)'},
    incrementable=['verbose'],
)
def build(ctx, verbose=False):
    """Build documentation"""
    command = ['tox', '-e', 'docs', '--', 'build', '--clean']
    insert_verbosity_flag(command, verbose)

    print('Building documentation...')
    ctx.run(' '.join(command))


@task(
    default=True,
    pre=[build],
    help={
        'no-open': 'Do not open the documentation in a web browser',
        'verbose': 'Increase verbosity (can be used additively)',
    },
    incrementable=['verbose'],
)
def serve(ctx, no_open=False, verbose=False):
    """Serve and view documentation in a web browser"""
    address = 'localhost:8765'

    command = ['tox', '-e', 'docs', '--', 'serve', '--livereload', '--dev-addr', address]
    insert_verbosity_flag(command, verbose)

    if not no_open:
        webbrowser.open_new_tab(f'http://{address}')

    ctx.run(' '.join(command))


@task
def publish(ctx):
    """Publish documentation on GitHub Pages"""
    github_token = os.getenv('GITHUB_TOKEN')
    if not github_token:
        raise Exit('No `GITHUB_TOKEN` has been set')

    site_dir = os.path.abspath('site')
    if not os.path.isdir(site_dir):
        raise Exit('Site directory does not exist, build docs by running `inv docs.build`')

    print('Reading current Git configuration...')
    git_user = get_git_user(ctx)
    git_email = get_git_email(ctx)
    latest_commit_hash = get_latest_commit_hash(ctx)

    if 'GITHUB_ACTIONS' in os.environ:
        remote = f'https://{os.getenv("GITHUB_ACTOR")}:{github_token}@github.com/ofek/csi-gcs.git'
    else:
        remote = f'https://{github_token}@github.com/ofek/csi-gcs.git'

    print('Copying site to a temporary directory...')
    with TemporaryDirectory() as d:
        temp_repo_dir = shutil.copytree(site_dir, os.path.join(os.path.realpath(d), 'site'))

        # https://help.github.com/en/github/working-with-github-pages/about-github-pages#static-site-generators
        # https://github.com/mkdocs/mkdocs/pull/2060
        print('Writing .nojekyll at the root...')
        create_file(os.path.join(temp_repo_dir, '.nojekyll'))

        origin = os.getcwd()
        try:
            os.chdir(temp_repo_dir)
            print('Configuring the temporary Git repository...')
            ctx.run('git init', hide=True)
            ctx.run(f'git config user.name "{git_user}"', hide=True)
            ctx.run(f'git config user.email "{git_email}"', hide=True)
            ctx.run(f'git remote add upstream {remote}', hide=True)

            print('Discovering remote...')
            ctx.run('git fetch --depth 1 upstream', hide=True)

            upstream_check = ctx.run(f'git ls-remote --heads {remote} gh-pages', hide=True)
            if upstream_check.stdout.strip():
                ctx.run('git reset upstream/gh-pages', hide=True)

            print('Committing site contents to branch gh-pages...')
            ctx.run('git add --all', hide=True)
            ctx.run(f'git commit --allow-empty -m "build docs at {latest_commit_hash}"', hide=True)
            ctx.run('git push upstream HEAD:gh-pages', hide=True)
        finally:
            os.chdir(origin)
