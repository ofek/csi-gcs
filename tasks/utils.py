import os
import subprocess
from .constants import DRIVER_NAME

def get_root():
    return os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def set_root():
    os.chdir(get_root())

def get_latest_commit_hash(ctx):
    result = ctx.run('git rev-parse HEAD', hide=True)
    return result.stdout.strip()

def get_git_user(ctx):
    user = os.getenv('GH_USER')
    if user is not None:
        return user

    result = ctx.run('git config --get user.name', hide=True)
    return result.stdout.strip()

def get_git_email(ctx):
    email = os.getenv('GH_EMAIL')
    if email is not None:
        return email

    result = ctx.run('git config --get user.email', hide=True)
    return result.stdout.strip()

def create_file(f):
    with open(f, 'a'):
        os.utime(f, None)

def get_version():
    version = subprocess.run(['git', 'describe', '--long', '--tags', '--match=v*', '--dirty'], stdout=subprocess.PIPE, cwd=get_root())
    if version.returncode == 0:
        return version.stdout.decode('utf-8').strip()

    current_ref = subprocess.run(['git', 'rev-list', '-n1', 'HEAD'], stdout=subprocess.PIPE, cwd=get_root())
    return current_ref.stdout.decode('utf-8').strip()

def image_name(version=False):
    if not version:
        version = get_version()
    return f'{DRIVER_NAME}:{version}'

def image_tags():
    last_tag = subprocess.run(['git', 'describe', '--tags', '--match=v*', '--abbrev=0'], stdout=subprocess.PIPE, cwd=get_root())

    if last_tag.returncode != 0:
        return ['dev']

    current_ref = subprocess.run(['git', 'rev-list', '-n1', 'HEAD'], stdout=subprocess.PIPE, cwd=get_root())
    last_tag_ref = subprocess.run(['git', 'rev-list', '-n1', last_tag.stdout.decode('utf-8').strip()], stdout=subprocess.PIPE, cwd=get_root())

    if last_tag_ref.stdout.decode('utf-8').strip() == current_ref.stdout.decode('utf-8').strip():
        return [last_tag.stdout.decode('utf-8').strip(), 'latest']

    return ['dev']


class EnvVars(dict):
    def __init__(self, env_vars=None, ignore=None):
        super(EnvVars, self).__init__(os.environ)
        self.old_env = dict(self)

        if env_vars is not None:
            self.update(env_vars)

        if ignore is not None:
            for env_var in ignore:
                self.pop(env_var, None)

    def __enter__(self):
        os.environ.clear()
        os.environ.update(self)

    def __exit__(self, exc_type, exc_value, traceback):
        os.environ.clear()
        os.environ.update(self.old_env)
