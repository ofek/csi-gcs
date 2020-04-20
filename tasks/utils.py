import os


def set_root():
    os.chdir(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def get_latest_commit_hash(ctx):
    result = ctx.run('git rev-parse HEAD', hide=True)
    return result.stdout.strip()


def create_file(f):
    with open(f, 'a'):
        os.utime(f, None)
