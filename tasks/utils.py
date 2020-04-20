import os


def set_root():
    os.chdir(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


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
