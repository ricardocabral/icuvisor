#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
	cat <<'USAGE'
Probe Intervals.icu custom-item list -> by-ID behavior against a real account.

Required environment:
  INTERVALS_ICU_API_KEY       Intervals.icu API key (never printed)
  INTERVALS_ICU_ATHLETE_ID    Athlete ID to probe (never printed)
  ICUVISOR_CUSTOM_ITEM_LIVE_PROBE_APPROVED=1
                              Explicit operator approval gate for live account access

Optional environment:
  INTERVALS_ICU_API_BASE_URL  Defaults to https://intervals.icu/api/v1
  ICUVISOR_PROBE_TIMEOUT_SECONDS
                              Per-request timeout, defaults to 20 seconds

Output is redacted by default: the probe prints only structural assertions and
counts. It never prints item IDs, names, content payloads, athlete IDs, or the API
key. Missing credentials or missing approval exit successfully with SKIP.
USAGE
	exit 0
fi

if [[ -z "${INTERVALS_ICU_API_KEY:-}" || -z "${INTERVALS_ICU_ATHLETE_ID:-}" ]]; then
	echo "SKIP custom-item by-ID live probe: set INTERVALS_ICU_API_KEY and INTERVALS_ICU_ATHLETE_ID to run; output is redacted."
	exit 0
fi

if [[ "${ICUVISOR_CUSTOM_ITEM_LIVE_PROBE_APPROVED:-}" != "1" ]]; then
	echo "SKIP custom-item by-ID live probe: set ICUVISOR_CUSTOM_ITEM_LIVE_PROBE_APPROVED=1 after explicit operator approval."
	exit 0
fi

if ! command -v python3 >/dev/null 2>&1; then
	echo "FAIL custom-item by-ID live probe: python3 is required." >&2
	exit 1
fi

python3 <<'PY'
import base64
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request


class ProbeFailure(Exception):
    pass


def env_required(name):
    value = os.environ.get(name, "").strip()
    if not value:
        raise ProbeFailure(f"missing required environment variable {name}")
    return value


def request_json(label, url, auth_header, timeout):
    request = urllib.request.Request(
        url,
        headers={
            "Authorization": auth_header,
            "Accept": "application/json",
            "User-Agent": "icuvisor-custom-item-by-id-live-probe",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:  # noqa: S310 - operator-approved live probe.
            body = response.read(2_000_000)
    except urllib.error.HTTPError as exc:
        raise ProbeFailure(f"{label} endpoint returned HTTP {exc.code}") from None
    except urllib.error.URLError:
        raise ProbeFailure(f"{label} endpoint request failed before an HTTP response") from None
    except TimeoutError:
        raise ProbeFailure(f"{label} endpoint timed out") from None

    try:
        return json.loads(body.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError):
        raise ProbeFailure(f"{label} endpoint did not return valid JSON") from None


def main():
    api_key = env_required("INTERVALS_ICU_API_KEY")
    athlete_id = env_required("INTERVALS_ICU_ATHLETE_ID")
    base_url = os.environ.get("INTERVALS_ICU_API_BASE_URL", "https://intervals.icu/api/v1").strip().rstrip("/")
    timeout_raw = os.environ.get("ICUVISOR_PROBE_TIMEOUT_SECONDS", "20").strip()
    try:
        timeout = float(timeout_raw)
    except ValueError:
        raise ProbeFailure("ICUVISOR_PROBE_TIMEOUT_SECONDS must be a number") from None
    if timeout <= 0:
        raise ProbeFailure("ICUVISOR_PROBE_TIMEOUT_SECONDS must be positive")

    credentials = f"API_KEY:{api_key}".encode("utf-8")
    auth_header = "Basic " + base64.b64encode(credentials).decode("ascii")
    athlete_path = urllib.parse.quote(athlete_id, safe="")
    list_url = f"{base_url}/athlete/{athlete_path}/custom-item"

    items = request_json("list", list_url, auth_header, timeout)
    if not isinstance(items, list):
        raise ProbeFailure("list endpoint returned a non-array payload")

    item_id = None
    for item in items:
        if isinstance(item, dict) and item.get("id") not in (None, ""):
            item_id = str(item["id"]).strip()
            break
    if not item_id:
        print("SKIP custom-item by-ID live probe: list endpoint returned no items with IDs; no real values printed.")
        return 0

    detail_url = f"{list_url}/{urllib.parse.quote(item_id, safe='')}"
    detail = request_json("detail", detail_url, auth_header, timeout)
    if not isinstance(detail, dict):
        raise ProbeFailure("detail endpoint returned a non-object payload")
    if str(detail.get("id", "")).strip() != item_id:
        raise ProbeFailure("detail endpoint returned an object for a different item ID")
    if "content" not in detail:
        raise ProbeFailure("detail endpoint omitted the content field")

    print(
        "PASS custom-item by-ID live probe: "
        f"list_count={len(items)} detail_key_count={len(detail)} content_present=true; "
        "real item identifiers, names, and content redacted"
    )
    return 0


try:
    raise SystemExit(main())
except ProbeFailure as exc:
    print(f"FAIL custom-item by-ID live probe: {exc}", file=sys.stderr)
    raise SystemExit(1)
PY
