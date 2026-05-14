# Custom item content schemas

This resource provides representative `content` samples for known Intervals.icu custom item families. icuvisor write tools still validate create/update payloads against readable custom items for the target athlete/item; these samples are guidance, not a validation allow-list, and unknown upstream item types can still pass through when the upstream API supports them.

## Charts, tables, traces, maps, histograms, and heatmaps

Descriptor key: `charts_tables_traces`

Display-oriented custom items define series/traces, formulas or fields, and layout/display options.

Item types: `FITNESS_CHART`, `FITNESS_TABLE`, `TRACE_CHART`, `ACTIVITY_CHART`, `ACTIVITY_HISTOGRAM`, `ACTIVITY_HEATMAP`, `ACTIVITY_MAP`

Representative `content` sample:

```json
{
  "axes": {
    "left": {
      "label": "Load"
    }
  },
  "filters": {
    "sport": "Ride"
  },
  "layout": {
    "height": 240,
    "width": 600
  },
  "series": [
    {
      "color": "blue",
      "field": "ctl",
      "formula": "ctl",
      "label": "Fitness"
    }
  ]
}
```

Inferred paths:

- `content`: object
- `content.axes`: object
- `content.axes.left`: object
- `content.axes.left.label`: string
- `content.filters`: object
- `content.filters.sport`: string
- `content.layout`: object
- `content.layout.height`: number
- `content.layout.width`: number
- `content.series`: array
- `content.series[]`: object
- `content.series[].color`: string
- `content.series[].field`: string
- `content.series[].formula`: string
- `content.series[].label`: string

## Input fields, activity fields, interval fields, and streams

Descriptor key: `fields_streams`

Field and stream items describe custom values, scripts/formulas, units, formats, and visibility.

Item types: `INPUT_FIELD`, `ACTIVITY_FIELD`, `INTERVAL_FIELD`, `ACTIVITY_STREAM`

Representative `content` sample:

```json
{
  "field": "travel_fatigue",
  "format": "0.0",
  "label": "Travel fatigue",
  "script": "return input",
  "type": "number",
  "units": "score",
  "visibility": "PRIVATE"
}
```

Inferred paths:

- `content`: object
- `content.field`: string
- `content.format`: string
- `content.label`: string
- `content.script`: string
- `content.type`: string
- `content.units`: string
- `content.visibility`: string

## Activity panels

Descriptor key: `panels`

Panel items group metrics, labels, and display widgets for activity detail pages.

Item types: `ACTIVITY_PANEL`

Representative `content` sample:

```json
{
  "layout": {
    "columns": 2
  },
  "visibility": "PRIVATE",
  "widgets": [
    {
      "display": "number",
      "field": "ftp",
      "label": "FTP"
    }
  ]
}
```

Inferred paths:

- `content`: object
- `content.layout`: object
- `content.layout.columns`: number
- `content.visibility`: string
- `content.widgets`: array
- `content.widgets[]`: object
- `content.widgets[].display`: string
- `content.widgets[].field`: string
- `content.widgets[].label`: string

## Zones

Descriptor key: `zones`

Zone items define named ranges and display colors for a metric.

Item types: `ZONES`

Representative `content` sample:

```json
{
  "metric": "power",
  "zones": [
    {
      "color": "gray",
      "max": 55,
      "min": 0,
      "name": "Z1"
    },
    {
      "color": "blue",
      "max": 75,
      "min": 56,
      "name": "Z2"
    }
  ]
}
```

Inferred paths:

- `content`: object
- `content.metric`: string
- `content.zones`: array
- `content.zones[]`: object
- `content.zones[].color`: string
- `content.zones[].max`: number
- `content.zones[].min`: number
- `content.zones[].name`: string
