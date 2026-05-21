# Code Review R003 — Step 1: Draft canonical formulas

**Verdict:** REVISE

## Findings

### 1. Formula draft table is contaminated by review-log rows

`STATUS.md:122-131` adds the canonical formula draft as a Markdown table, but the final two rows are review history entries:

- `| 2026-05-20 14:25 | Review R001 | plan Step 1: REVISE |`
- `| 2026-05-20 14:28 | Review R002 | plan Step 1: APPROVE |`

Because they are immediately adjacent to the formula table, Markdown parsers and humans will treat them as malformed formula rows. Step 2 is expected to translate this table into `icuvisor://analysis-formulas`; leaving non-formula rows in the source-of-truth table risks either broken generated content or accidentally publishing review metadata as formula data.

**Fix:** Move these entries to the existing Reviews/Execution Log section or remove them from Notes, and keep the formula table containing only formula rows.

### 2. Several citations are not specific enough to be publishable canonical references

The source-selection rule in `STATUS.md:118` says to use “URLs/publication details,” but the table uses vague citations such as “TrainingPeaks public education,” “TrainingPeaks/WKO public documentation,” and “Stephen Seiler public work” (`STATUS.md:124-128`). These are not stable enough for a canonical resource that analyzers will cite via `_meta.formula_ref`; future maintainers cannot verify exactly which source backs the formula/sign convention/boundary handling.

**Fix:** Replace vague references with concrete source details: URLs for public web pages, DOI/journal/year for papers, and edition/year/page/chapter where using books. Keep summaries in original words, but make the citation target unambiguous.

### 3. The z-score boundary rule introduces a second definition not supported by the PRD

`STATUS.md:129` says to “only present `z=0` when the caller explicitly requests a degenerate equal-value display” when standard deviation is zero. A zero standard deviation makes `(current - mean) / stddev` undefined; the PRD’s analyzer rules require boundary-safe defaults with insufficient-sample/invalid-result metadata instead of returning garbage values. Adding a caller-controlled escape hatch creates definition drift for a formula the task says should be locked.

**Fix:** Make zero variance a single canonical unavailable/insufficient-variance outcome. If a future UI wants to display “current equals baseline,” expose that as separate explanatory metadata/text, not as a z-score value.

## Notes

The six requested refs are present and use the full `icuvisor://analysis-formulas#...` form, and the HR drift/Pw:HR wording clearly distinguishes the two concepts. No Go tests were run because Step 1 only changes task documentation.
