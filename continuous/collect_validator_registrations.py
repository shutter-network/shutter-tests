"""
requirements:
    eth_abi
    requests
"""

import sys
import os
import binascii
import textwrap
import json
from typing import Any, Dict, Iterator, List
from time import sleep
import eth_abi

import requests

API_KEY = os.environ.get("GNOSISSCAN_API_KEY")
if API_KEY is None:
    sys.exit("You need to export GNOSISSCAN_API_KEY")

DEPOSIT_CONTRACT = "0x0B98057eA310F4d31F2a452B414647007d1645d9"

BASE_URL = "https://api.gnosisscan.io/api"

DEPOSIT_EVENT = "0x649bbc62d0e31342afea4e5cd82d4049e7e1ee912fc0889aa790803be39038c5"


def get_params(
    from_block: int, to_block: int, page: int, offset: int
) -> Dict[str, Any]:
    return dict(
        module="logs",
        action="getLogs",
        address=DEPOSIT_CONTRACT,
        fromBlock=from_block,
        toBlock=to_block,
        page=page,
        offset=offset,
        apikey=API_KEY,
    )


def query(
    session: requests.Session,
    params: Dict[str, Any],
) -> Dict[str, Any]:
    response = session.get(BASE_URL, params=params)
    return response.json()


def slice_and_dice(from_block: int, to_block: int) -> Iterator[Dict[str, Any]]:
    block_range = 1000
    for start in range(from_block, to_block, block_range):
        yield dict(from_block=start, to_block=start + block_range)


def paginate(block_range: Dict[str, Any]) -> Iterator[Dict[str, Any]]:
    max_results = 10000
    page_size = 100
    for p in range(1, max_results // page_size + 1):
        yield get_params(**block_range, page=p, offset=page_size), page_size


def extract(results: List[Dict[str, Any]]) -> Iterator[Any]:
    for result in results:
        try:
            _ = result.keys()
        except:
            print("break", result)
            break
        topic0 = result["topics"][0]
        if topic0 != DEPOSIT_EVENT:
            continue
        block = int(result["blockNumber"], 16)
        yield (result["data"], block)


def parse_deposit_event(data: str):
    slice_len = len(data[2:]) // 4
    decoded = eth_abi.decode(['bytes', 'bytes', 'bytes', 'bytes', 'bytes'], bytes.fromhex(data[2:]))#.encode("utf-8"))
    pubkey, withdrawal_credentials, amount, signature, index = decoded
    print(binascii.hexlify(pubkey), binascii.hexlify(withdrawal_credentials))


if __name__ == "__main__":
    session = requests.Session()
    count = 10
    for block_range in slice_and_dice(29029018, 38504792):
        for params, page_size in paginate(block_range):
            result = query(session, params)
            if result["status"] == "0":
                break
            for x, y in extract(result["result"]):
                parse_deposit_event(x)
            if len(result["result"]) < page_size:
                break
            sleep(0.2)
