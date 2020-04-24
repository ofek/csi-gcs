from invoke import Collection

from . import docs
from . import image
from . import test
from . import docker
from . import build
from .utils import set_root

ns = Collection()
ns.add_collection(Collection.from_module(image))
ns.add_collection(Collection.from_module(docs))
ns.add_collection(Collection.from_module(test))
ns.add_collection(Collection.from_module(docker))
ns.add_collection(Collection.from_module(build))

set_root()
