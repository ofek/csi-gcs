from invoke import Collection

from . import docs
from . import image
from . import test
from . import env
from . import build
from . import codegen
from .utils import set_root

ns = Collection()
ns.add_collection(Collection.from_module(image))
ns.add_collection(Collection.from_module(docs))
ns.add_collection(Collection.from_module(test))
ns.add_collection(Collection.from_module(env))
ns.add_collection(Collection.from_module(build))
ns.add_collection(Collection.from_module(codegen))

set_root()
