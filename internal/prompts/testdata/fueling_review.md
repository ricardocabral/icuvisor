Prompt: Fueling review
Scope: start_date=2026-05-01, end_date=2026-05-14, race_date=2026-06-07, race_name=A Race.
Resources: icuvisor://athlete-profile.
Tools: get_athlete_profile, resolve_calendar_dates, get_activities, get_activity_details, get_wellness_data, get_training_summary, get_events.
Do:
- Read profile first for athlete-local timezone and units. For a default or relative window, call resolve_calendar_dates with offsets -14 and -1; use returned athlete-local dates, not UTC or client-time arithmetic.
- For activity_id, read terse get_activity_details once. For a date range, read terse paginated get_activities with include_unnamed:true for duration, load, carbs_ingested_g, and carbs_used_g; do not fetch details for every row.
- Fetch every activity page needed before calling a range complete; otherwise state covered count/window as partial, count unavailable or Strava-blocked rows separately, and never reveal next_page_token. Preserve activity timezones and label current-day _meta.as_of data partial.
- Read get_wellness_data only when daily nutrition evidence is useful, with fields kcalConsumed, carbohydrates, protein, and fatTotal plus only explicitly requested custom codes. Keep its returned calories_intake, carbs_g, protein_g, and fat_g as daily fields; do not broaden into health fields when nutrition provenance or freshness is unavailable.
- Keep carbs_ingested_g as athlete-logged during-activity intake, carbs_used_g as an upstream used/burned estimate, and calories_burned/load as context only. Preserve custom-field codes and unknown meanings; never substitute carbs_used_g, wellness totals, calories, or load for intake.
- Use moving_time_seconds only: logged carbs/hour = carbs_ingested_g / (moving_time_seconds / 3600), labelled g/h. Calculate only a non-negative numeric ingested value with positive duration; zero yields 0 g/h, while absent/negative intake, invalid duration, and unavailable rows are counted exclusions. Aggregate only eligible grams and those same durations, then state eligible/total-session coverage.
- When race_date is supplied, call get_events with oldest/newest equal to that athlete-local date and limit:100; race_name only disambiguates that result. Mark _meta.truncated race context partial, and call a complete no-match unconfirmed rather than inventing an event.
- Return separate Sourced activity evidence, Sourced daily-wellness evidence, Sourced race/calendar context, Labelled calculations, Coverage and data gaps, and General educational guidance sections. Surface _meta.stale, _meta.missing_fields, field semantics/provenance, and availability warnings.
Guardrails:
- This workflow is read-only: do not call write/delete tools, include_full, streams, or raw payloads.
- Do not treat missing data as zero or inadequate fueling, extrapolate a rate, invent intake or targets, or create a food/product library.
- Do not diagnose health conditions or eating disorders, prescribe individualized nutrition, infer deficits, recommend carbohydrate/calorie/sodium/fluid/sweat-rate targets, or claim product/performance effects.
- Keep general material visibly separate as conditional educational guidance; refer individualized or medical nutrition requests to a qualified sports dietitian or clinician.
Return: source-labelled logged-fueling evidence, transparent g/h calculations where valid, covered-session counts and exclusions, calendar context when confirmed, and clearly separated educational guidance.
