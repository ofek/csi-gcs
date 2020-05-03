import re
import subprocess

MARKER = '<STABLE_VERSION>'
VERSION = None


def get_latest_tag():
    result = subprocess.run(['git', 'tag'], capture_output=True, check=True)
    lines = result.stdout.decode('utf-8').splitlines()

    pattern = re.compile(r'^v(\d+).(\d+).(\d+)$')
    tags = []
    for line in lines:
        if match := pattern.search(line):
            tags.append(tuple(map(int, match.groups())))

    if not tags:
        raise Exception('No tags found')

    tags.sort()
    latest = '.'.join(map(str, tags[-1]))

    return f"v{latest}"


def patch(lines):
    """This injects the latest stable version based on tags."""
    global VERSION
    if VERSION is None:
        VERSION = get_latest_tag()

    for i, line in enumerate(lines):
        lines[i] = line.replace(MARKER, VERSION)
