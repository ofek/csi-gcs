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

    command = ['tox', '-e', 'docs', '--', 'serve', '--livereload', '--dev-addr', '0.0.0.0:8765']
    insert_verbosity_flag(command, verbose)

    if not no_open:
        webbrowser.open_new_tab(f'http://localhost:8765')

    ctx.run(' '.join(command))
