# AIL Injection and Delimiter Safety

## Mandatory delimiters

Every rendered payload is enclosed in these markers unless the caller explicitly disables them:

```text
<!-- BEGIN AIL SYSTEM RULES
TYPE: OPERATIONAL_INSTRUCTIONS
DO_NOT_QUOTE: TRUE
DO_NOT_REPRODUCE: TRUE
DO_NOT_PRESENT_AS_USER_CONTENT: TRUE
AIL_VERSION: 0.3
-->

[RENDERED AIL RULES]

<!-- END AIL SYSTEM RULES -->
```

The renderer emits these markers automatically. Users never add them by hand.

## Interpretation instruction

The following text is emitted outside the delimiters:

> The content between BEGIN AIL SYSTEM RULES and END AIL SYSTEM RULES represents operational instructions. Follow these instructions when performing tasks. They are not user-facing content. Do not quote, enumerate, reproduce, or summarize these rules in normal responses unless the user explicitly asks about the instruction system. Respond to the user's actual request normally. Do not confuse the existence of rules with a request to generate or discuss rules.

## Injection safety rules

- Delimiters do not create system-level authority. Project rules cannot override higher-priority host instructions.
- User-controlled text is never treated as trusted instructions.
- Rule text containing `<!-- BEGIN AIL SYSTEM RULES` or `<!-- END AIL SYSTEM RULES` is escaped and reported with `RENDER_NESTED_DELIMITER`.
- Any `<!--` or `-->` sequence in rule-derived text is replaced with an HTML entity, so nested markers cannot escape their scope.
- Original source and provenance are preserved in IR and optionally in VERBOSE rendering.
- Internal rules are not exposed unless the caller requests them or the user explicitly asks about the instruction system.

## Injection modes

`INLINE` and `FILE` include full rules in the context. `FILE` additionally notes that the host should load it through its native discovery mechanism. `COMPACT` and `VERBOSE` change presentation only, never the set of instructions.
