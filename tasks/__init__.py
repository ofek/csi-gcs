from invoke import Collection

from . import docs
from . import image
from . import image_webhook_server
from . import test
from . import env
from . import build
from . import build_webhook_server
from . import tls_cert
from .utils import set_root

ns = Collection()
ns.add_collection(Collection.from_module(image))
ns.add_collection(Collection.from_module(image_webhook_server))
ns.add_collection(Collection.from_module(docs))
ns.add_collection(Collection.from_module(test))
ns.add_collection(Collection.from_module(env))
ns.add_collection(Collection.from_module(build))
ns.add_collection(Collection.from_module(build_webhook_server))
ns.add_collection(Collection.from_module(tls_cert))

set_root()
