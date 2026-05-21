# R001 Plan Review — Step 1: Map nutrition fields from upstream fixtures

**Verdict:** Approved with required clarifications before coding Step 2.

I did not find a separate detailed implementation plan beyond the Step 1 checklist in `PROMPT.md`/`STATUS.md`; this review treats those bullets as the plan.

## Findings

1. **Step 1 is directionally correct, but must leave an explicit mapping artifact.**  
   Current code already decodes wellness nutrition fields in `internal/intervals/wellness.go`:
   - `kcalConsumed` → likely intake calories
   - `carbohydrates` → macro grams
   - `protein` → macro grams
   - `fatTotal` → macro grams

   Activity decoding in `internal/intervals/activities.go` currently exposes only `calories` for activity energy, already shaped as `calories_burned` in activity rows. I did not see activity macro fields in existing fixtures. Step 1 should record this mapping/gap in `STATUS.md` Discoveries or a concise table before later shaping changes.

2. **Do not invent activity macros or total calories if fixtures do not prove them.**  
   The plan includes `calories_total` as an example, but the implementation should only choose that key if an upstream fixture demonstrates total-calorie semantics. For the fields visible today, the safer mapping appears to be `kcalConsumed` → `calories_intake` (or a still more explicit kcal-suffixed variant if the team chooses one), and activity `calories` → `calories_burned`.

3. **Fixture work needs to distinguish observed upstream data from synthetic regression data.**  
   Existing checked-in activity/wellness fixtures do not currently contain nutrition macros. If Step 1 adds fixture values, make clear whether they are copied from observed upstream payload shape/test captures or intentionally minimal synthetic fixtures for decoder coverage. Avoid presenting synthetic activity macro fields as upstream-supported.

4. **Plan Step 2 dependency: wellness currently clones raw fields first.**  
   `wellnessRow` starts from `cloneJSONMap(row.Raw)`, so legacy upstream names like `kcalConsumed`, `carbohydrates`, `protein`, and `fatTotal` will continue to leak into terse responses unless the shaping step explicitly removes or translates them. This is not a Step 1 blocker, but Step 1's chosen key map should call it out so Step 2 does not accidentally emit both old ambiguous keys and new disambiguated keys.

## Recommended Step 1 acceptance criteria

Before marking Step 1 complete, ensure the status/discovery notes contain a table like:

| Surface | Upstream JSON key | Current typed field | Chosen public key | Semantics | Fixture evidence |
|---|---|---|---|---|---|
| wellness | `kcalConsumed` | `KcalConsumed` | `calories_intake` | consumed/intake kcal | fixture path or documented gap |
| wellness | `carbohydrates` | `Carbohydrates` | `carbs_g` | grams carbohydrate | fixture path or documented gap |
| wellness | `protein` | `Protein` | `protein_g` | grams protein | fixture path or documented gap |
| wellness | `fatTotal` | `FatTotal` | `fat_g` | grams total fat | fixture path or documented gap |
| activity | `calories` | `Calories` | `calories_burned` | active/exercise calories | fixture path |
| activity | macros/total calories | absent unless proven | none | gap, do not emit | documented gap |

With that explicit mapping and gap documentation, the Step 1 plan is safe to proceed.
