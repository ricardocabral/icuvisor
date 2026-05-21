import copy
import json
import tempfile
import unittest
from pathlib import Path

from kr5_benchmark import (
    ANALYZER_DISABLED_MODE,
    ANALYZER_ENABLED_MODE,
    BenchmarkError,
    TokenCounter,
    ToolCall,
    load_fixture,
    load_prompt_set,
    raw_stream_pull_count,
    summarize,
    validate_measurement,
)

ROOT = Path(__file__).resolve().parents[2]
PROMPT_SET = ROOT / "scripts/benchmark/prompts/kr5_shared_prompts.json"
FIXTURE_DIR = ROOT / "scripts/benchmark/testdata/fixtures"
ANALYZER_FIXTURE = FIXTURE_DIR / "icuvisor-analyzer-family.json"
CORE_FIXTURE = FIXTURE_DIR / "icuvisor-core.json"


class KR5BenchmarkAnalyzerTest(unittest.TestCase):
    def setUp(self):
        self.prompt_set = load_prompt_set(PROMPT_SET)
        self.analyzer_measurement = load_fixture(ANALYZER_FIXTURE)

    def test_scoped_analyzer_prompts_do_not_break_legacy_fixtures(self):
        validate_measurement(self.prompt_set, load_fixture(CORE_FIXTURE))

    def test_v2_analyzer_fixture_invariants(self):
        result = summarize(
            self.prompt_set,
            [self.analyzer_measurement],
            TokenCounter(),
            generated_at="2026-05-20T00:00:00Z",
        )
        self.assertEqual(result["schema_version"], "kr5-benchmark-result-v2")
        server = result["servers"][0]
        summaries = server["mode_summaries"]
        self.assertIn(ANALYZER_ENABLED_MODE, summaries)
        self.assertIn(ANALYZER_DISABLED_MODE, summaries)
        self.assertEqual(summaries[ANALYZER_ENABLED_MODE]["raw_stream_pull_count"], 0)
        self.assertEqual(summaries[ANALYZER_DISABLED_MODE]["raw_stream_pull_count"], 3)
        self.assertLess(
            summaries[ANALYZER_ENABLED_MODE]["response_tokens"],
            summaries[ANALYZER_DISABLED_MODE]["response_tokens"],
        )

    def test_disabled_mode_rejects_analyzer_tool_catalog_exposure(self):
        measurement = copy.deepcopy(self.analyzer_measurement)
        analyzer_tool = next(
            tool
            for tool in measurement.mode_catalogs[ANALYZER_ENABLED_MODE]
            if tool["name"] == "analyze_trend"
        )
        measurement.mode_catalogs[ANALYZER_DISABLED_MODE].append(analyzer_tool)
        with self.assertRaisesRegex(BenchmarkError, "disabled analyzer mode exposes"):
            validate_measurement(self.prompt_set, measurement)

    def test_disabled_mode_rejects_analyzer_tool_calls(self):
        measurement = copy.deepcopy(self.analyzer_measurement)
        call = next(
            item
            for item in measurement.calls
            if item.prompt_id == "KR5-A01" and item.mode == ANALYZER_DISABLED_MODE
        )
        call.tool = "analyze_trend"
        with self.assertRaisesRegex(BenchmarkError, "unlisted tool|disabled analyzer mode calls"):
            validate_measurement(self.prompt_set, measurement)

    def test_analyzer_modes_reject_non_analyzer_catalog_payload_drift(self):
        measurement = copy.deepcopy(self.analyzer_measurement)
        tool = next(
            item
            for item in measurement.mode_catalogs[ANALYZER_ENABLED_MODE]
            if item["name"] == "get_activities"
        )
        tool["description"] = tool["description"] + " drift"
        with self.assertRaisesRegex(BenchmarkError, "catalog payload differs"):
            validate_measurement(self.prompt_set, measurement)

    def test_expected_source_tool_usage_is_count_validated(self):
        measurement = copy.deepcopy(self.analyzer_measurement)
        call = next(
            item
            for item in measurement.calls
            if item.prompt_id == "KR5-A01" and item.tool == "analyze_trend"
        )
        call.source_tool_usage = [{"tool": "get_activities", "count": 2}]
        with self.assertRaisesRegex(BenchmarkError, "source_tool_usage"):
            validate_measurement(self.prompt_set, measurement)

    def test_raw_stream_pull_count_counts_only_top_level_stream_calls(self):
        calls = [
            ToolCall("P", "i", "get_activity_streams", {}, result={}),
            ToolCall("P", "i", "icu_get_activity_streams", {}, result={}),
            ToolCall("P", "i", "unavailable:get_activity_streams", {}, result={"isError": True}),
            ToolCall(
                "P",
                "i",
                "get_activity_histogram",
                {},
                result={},
                source_tool_usage=[{"tool": "get_activity_streams", "count": 1}],
            ),
        ]
        self.assertEqual(raw_stream_pull_count(calls), 2)


if __name__ == "__main__":
    unittest.main()
