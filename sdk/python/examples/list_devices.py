import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from _env import bootstrap, load_env

bootstrap()
load_env()

from wlte_openapi import WlteClient

client = WlteClient(
    client_id=os.environ["WLTE_CLIENT_ID"],
    client_secret=os.environ["WLTE_CLIENT_SECRET"],
    base_url=os.environ.get("WLTE_BASE_URL"),
)

print(client.devices.list(page=1, page_size=50))
