# Project agent rules - conventional Markdown baseline

This file expresses the same behavioral requirements as `.ail/rules/project-rules.ail` without AIL syntax. Keep both files in sync.

## R001 - Symbols and comments (CRITICAL)

Applies to Go source files.

- Do: `use_ascii_hyphen`, `preserve_existing_comments`
- Never: `emit_typographic_dash`, `rewrite_adjacent_comments`
- Fallback: `ask_user_before_inserting_symbol`

Use the ASCII hyphen in every file. Preserve comments that already exist. Do not rewrite comments that are unrelated to the change.

## R002 - File size (HIGH)

Applies to Go source files.

- Do: `keep_files_under_700_lines`, `extract_focused_submodules`
- Never: `create_god_files`
- Fallback: `split_by_domain_concern`

No source file should exceed 700 lines. When a file approaches the limit, extract focused submodules and split by domain concern.

## R003 - Verification (HIGH)

Applies when a linter is configured for the project.

- Do: `run_all_tests_after_edits`, `fix_every_diagnostic_in_touched_files`
- Never: `ignore_warnings`
- Verify: `tests_pass`, `vet_clean`
- Fallback: `report_remaining_diagnostics`

After editing a file, run the full test suite and fix every error and warning in the files you touched. Do not dismiss warnings as minor.

## R004 - Environment boundary (CRITICAL)

Applies before any command is run.

- Do: `stay_local`, `ask_permission_before_remote_operations`
- Never: `touch_production`, `use_production_credentials`
- Fallback: `stop_and_ask_user`

You are always in a local development environment. Never access production systems or credentials. Ask before any remote operation.

## R005 - Approach selection (HIGH)

Applies when choosing an implementation approach.

- Do: `state_assumptions_explicitly`, `prefer_simplicity`, `make_surgical_changes`
- Never: `add_speculative_abstractions`, `improve_unrelated_code`
- Fallback: `present_tradeoffs_and_wait`

State assumptions before coding. Prefer the smallest change that solves the problem. Touch only what the request requires.

## R006 - Success criteria (HIGH)

Applies when defining a goal.

- Do: `define_verifiable_criteria`, `iterate_until_verified`
- Never: `claim_success_without_evidence`
- Fallback: `report_exact_coverage`

Turn tasks into checks you can run. Report exactly what was covered when something blocks completion.

## R007 - Directory consistency (MEDIUM)

Applies when a concept already exists.

- Do: `use_one_directory_per_concept`, `search_before_creating`
- Never: `duplicate_implementations`
- Fallback: `extend_existing_implementation`

One concept lives in one directory. Search by name and purpose before creating a new type, package or helper.

## R008 - User-facing language (MEDIUM)

Applies when the user writes in Brazilian Portuguese.

- Do: `use_proper_accents`, `avoid_technical_jargon_in_user_copy`
- Never: `output_unaccented_portuguese`
- Fallback: `ask_user_for_preferred_wording`

Use correct accents in Portuguese strings and responses. Keep user-facing copy free of technical jargon.

## R009 - Development exposure (MEDIUM)

Applies when a service runs in development mode.

- Do: `expose_over_tailscale`
- Never: `bind_only_to_loopback_without_tailscale`
- Fallback: `implement_tailscale_access_before_reporting_done`

Any service started in development mode must also be reachable over Tailscale.

## R010 - Logging (LOW)

Applies when adding log statements.

- Do: `log_decisions_and_io`, `log_full_error_objects`
- Never: `log_secrets`, `log_only_error_messages`
- Fallback: `use_stable_module_prefix`

Log decisions, effective configuration, external I/O and full error objects. Never log secrets. Use a stable prefix per module.
