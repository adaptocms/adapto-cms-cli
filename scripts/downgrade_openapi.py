"""Downgrade a FastAPI OpenAPI 3.1 spec to 3.0 so oapi-codegen (no 3.1 support) can generate the client."""

import json
import sys


def fix(node):
    if isinstance(node, dict):
        for key in ("anyOf", "oneOf"):
            variants = node.get(key)
            if isinstance(variants, list):
                has_null = any(isinstance(v, dict) and v.get("type") == "null" for v in variants)
                non_null = [v for v in variants if not (isinstance(v, dict) and v.get("type") == "null")]
                if has_null:
                    del node[key]
                    node["nullable"] = True
                    if len(non_null) == 1 and isinstance(non_null[0], dict):
                        for k, v in non_null[0].items():
                            node.setdefault(k, v)
                    elif non_null:
                        node[key] = non_null
        t = node.get("type")
        if isinstance(t, list) and "null" in t:
            node["nullable"] = True
            rest = [x for x in t if x != "null"]
            node["type"] = rest[0] if len(rest) == 1 else rest
        for v in list(node.values()):
            fix(v)
    elif isinstance(node, list):
        for v in node:
            fix(v)


spec = json.load(open(sys.argv[1]))
spec["openapi"] = "3.0.3"
fix(spec)
json.dump(spec, open(sys.argv[2], "w"))
print("downgraded 3.1 -> 3.0 ->", sys.argv[2])
