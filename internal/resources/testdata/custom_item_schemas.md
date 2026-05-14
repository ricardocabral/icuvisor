# Custom item content schemas

This resource provides representative `content` samples for known Intervals.icu custom item families. icuvisor write tools still validate create/update payloads against readable custom items for the target athlete/item; these samples are guidance, not a validation allow-list, and unknown upstream item types can still pass through when the upstream API supports them.

## Charts, tables, traces, maps, histograms, and heatmaps

Descriptor key: `charts_tables_traces`

Display-oriented custom items define series/traces, formulas or fields, and layout/display options.

### `FITNESS_CHART`

Fitness page chart with series, axes, filters, and layout.

Representative `content` sample:

```json
{
  "axes": {
    "left": {
      "label": "Value"
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
      "label": "Primary"
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

### `FITNESS_TABLE`

Fitness page table with columns and filters.

Representative `content` sample:

```json
{
  "columns": [
    {
      "field": "ctl",
      "format": "0",
      "label": "Fitness"
    }
  ],
  "filters": {
    "sport": "Ride"
  },
  "sort": {
    "direction": "desc",
    "field": "date"
  }
}
```

Inferred paths:

- `content`: object
- `content.columns`: array
- `content.columns[]`: object
- `content.columns[].field`: string
- `content.columns[].format`: string
- `content.columns[].label`: string
- `content.filters`: object
- `content.filters.sport`: string
- `content.sort`: object
- `content.sort.direction`: string
- `content.sort.field`: string

### `TRACE_CHART`

Trace chart with stream or field series.

Representative `content` sample:

```json
{
  "axes": {
    "left": {
      "label": "Value"
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
      "field": "heart_rate",
      "formula": "heart_rate",
      "label": "Primary"
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

### `ACTIVITY_CHART`

Activity detail chart with per-activity series.

Representative `content` sample:

```json
{
  "axes": {
    "left": {
      "label": "Value"
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
      "field": "power",
      "formula": "power",
      "label": "Primary"
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

### `ACTIVITY_HISTOGRAM`

Activity histogram with bucketed metric bins.

Representative `content` sample:

```json
{
  "bucket_size": 25,
  "colors": [
    "blue",
    "green",
    "orange"
  ],
  "field": "power",
  "units": "W"
}
```

Inferred paths:

- `content`: object
- `content.bucket_size`: number
- `content.colors`: array
- `content.colors[]`: string
- `content.field`: string
- `content.units`: string

### `ACTIVITY_HEATMAP`

Activity heatmap with x/y fields and color scale.

Representative `content` sample:

```json
{
  "color": {
    "field": "time",
    "scale": "viridis"
  },
  "x_field": "cadence",
  "y_field": "power"
}
```

Inferred paths:

- `content`: object
- `content.color`: object
- `content.color.field`: string
- `content.color.scale`: string
- `content.x_field`: string
- `content.y_field`: string

### `ACTIVITY_MAP`

Activity map overlay with path and metric coloring.

Representative `content` sample:

```json
{
  "markers": [
    {
      "field": "lap",
      "label": "Lap"
    }
  ],
  "overlay": {
    "field": "power",
    "palette": "thermal"
  },
  "path": {
    "source": "latlng"
  }
}
```

Inferred paths:

- `content`: object
- `content.markers`: array
- `content.markers[]`: object
- `content.markers[].field`: string
- `content.markers[].label`: string
- `content.overlay`: object
- `content.overlay.field`: string
- `content.overlay.palette`: string
- `content.path`: object
- `content.path.source`: string

## Input fields, activity fields, interval fields, and streams

Descriptor key: `fields_streams`

Field and stream items describe custom values, scripts/formulas, units, formats, and visibility.

### `INPUT_FIELD`

Athlete-entered custom field definition.

Representative `content` sample:

```json
{
  "field": "travel_fatigue",
  "format": "0.0",
  "label": "Custom field",
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

### `ACTIVITY_FIELD`

Computed or entered activity-level custom field.

Representative `content` sample:

```json
{
  "field": "fueling_score",
  "format": "0.0",
  "label": "Custom field",
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

### `INTERVAL_FIELD`

Interval-level custom field definition.

Representative `content` sample:

```json
{
  "field": "interval_note",
  "format": "0.0",
  "label": "Custom field",
  "script": "return input",
  "type": "text",
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

### `ACTIVITY_STREAM`

Custom activity stream definition with samples or script metadata.

Representative `content` sample:

```json
{
  "samples": [
    20.1,
    20.4,
    20.8
  ],
  "script": "return samples",
  "stream": "custom_temperature",
  "units": "C"
}
```

Inferred paths:

- `content`: object
- `content.samples`: array
- `content.samples[]`: number
- `content.script`: string
- `content.stream`: string
- `content.units`: string

## Activity panels

Descriptor key: `panels`

Panel items group metrics, labels, and display widgets for activity detail pages.

### `ACTIVITY_PANEL`

Activity detail panel with widgets and layout.

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

### `ZONES`

Named metric zone ranges and colors.

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
