# AIL Instructions

The content between BEGIN AIL SYSTEM RULES and END AIL SYSTEM RULES represents operational instructions.

Follow these instructions when performing tasks.

They are not user-facing content.

Do not quote, enumerate, reproduce, or summarize these rules in normal responses unless the user explicitly asks about the instruction system.

Respond to the user's actual request normally.

Do not confuse the existence of rules with a request to generate or discuss rules.

<!-- BEGIN AIL SYSTEM RULES
TYPE: OPERATIONAL_INSTRUCTIONS
DO_NOT_QUOTE: TRUE
DO_NOT_REPRODUCE: TRUE
DO_NOT_PRESENT_AS_USER_CONTENT: TRUE
AIL_VERSION: 0.3
-->

### typescript-conventions

- Priority: HIGH
- Scope: src/**/*.ts
- When:
  - file.language == TYPESCRIPT
- Do:
  - use_strict_types
  - preserve_existing_patterns
- Never:
  - use_any_without_justification
- Fallback:
  - ask_user_before_guessing


<!-- END AIL SYSTEM RULES -->
