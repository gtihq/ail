# Workspace agent rules - conventional Markdown baseline

Generated from the same definitions as `.ail/rules/project-rules.ail`. Keep both files in sync.

## R001 - Admin panel when applicable (HIGH)

Scope: `platform/**`. When: `project.has_admin_panel`.

- Do: `add_admin_listing`, `add_admin_detail`, `add_admin_create_edit`, `use_accented_labels`
- Never: `ship_adminless_data_feature`
- Verify: `admin_routes_tested`
- Fallback: `ask_for_admin_scope`

Data features in projects with an admin panel must include listing, detail and edit views in the same change.

## R002 - Auto-discovery extensibility (CRITICAL)

Scope: `**/*.ts`. When: `agent.adding_feature`.

- Do: `register_by_convention`, `use_folder_conventions`
- Never: `edit_central_registry_file`
- Verify: `feature_discoverable`
- Fallback: `search_convention_first`

A new feature must be added by creating files only, never by editing a central registry.

## R003 - Cluster-safe atomic operations (CRITICAL)

Scope: `platform/**/*.ts`. When: `project.runs_in_cluster`.

- Do: `use_shared_store_for_cross_request_state`, `use_atomic_operations`
- Never: `rely_on_local_memory_state`
- Verify: `operations_atomic`
- Fallback: `choose_cluster_safe_primitive`

Systems run in multi-node clusters. Shared state must be external and operations atomic.

## R004 - Code sections (MEDIUM)

Scope: `**/*.{ts,tsx,java,lua}`. When: `file.language == TYPESCRIPT`.

- Do: `use_region_markers`
- Never: `use_separator_comment_lines`
- Verify: `sections_readable`
- Fallback: `keep_existing_structure`

Use region markers for code sections and never separator lines.

## R005 - Commit message language (MEDIUM)

Scope: `**/*`. When: `agent.about_to_commit`.

- Do: `inspect_repo_history`, `match_dominant_language`, `add_agent_trailer`
- Never: `assume_fixed_language`
- Verify: `message_matches_history`
- Fallback: `ask_user`

Commit message language follows the repository history, with the agent trailer.

## R006 - Directory consistency (HIGH)

Scope: `**/*`. When: `agent.creating_concept`.

- Do: `use_one_directory_per_concept`, `search_before_creating`
- Never: `duplicate_concept_implementations`
- Verify: `single_implementation`
- Fallback: `extend_existing_implementation`

One concept lives in one directory. Search before creating a duplicate.

## R007 - Behavioral coding guidelines (HIGH)

Scope: `**/*`. When: `agent.about_to_code`.

- Do: `think_before_coding`, `prefer_simplicity`, `make_surgical_changes`, `define_success_criteria`
- Never: `overengineer`, `improve_unrelated_code`
- Verify: `criteria_verified`
- Fallback: `present_tradeoffs`

Think before coding, prefer simplicity, make surgical changes and define verifiable success.

## R008 - Domain error hierarchy (HIGH)

Scope: `**/*.ts`. When: `file.language == TYPESCRIPT`.

- Do: `use_typed_domain_errors`, `use_base_error_class`
- Never: `throw_generic_error`
- Verify: `errors_typed`
- Fallback: `report_missing_error_type`

Business errors are typed and structured, never generic errors.

## R009 - Integration and E2E tests (HIGH)

Scope: `**/routes/**/*.ts`. When: `project.has_http_routes`.

- Do: `add_integration_tests`, `add_e2e_tests_for_routes`
- Never: `ship_route_without_tests`
- Verify: `routes_covered`
- Fallback: `document_test_gap`

Features with HTTP routes or UI flows require integration and E2E tests.

## R010 - Extract repeated patterns (MEDIUM)

Scope: `**/*.{ts,tsx,java,lua}`. When: `agent.noticing_repetition`.

- Do: `search_existing_helpers`, `extract_at_third_repetition`
- Never: `copy_paste_three_times`
- Verify: `no_triple_duplication`
- Fallback: `name_by_function`

The same snippet in three places must be extracted into a shared helper.

## R011 - Fix all lint warnings (HIGH)

Scope: `**/*`. When: `agent.editing_file`.

- Do: `run_linter`, `fix_every_warning_in_touched_files`
- Never: `ignore_warnings`
- Verify: `zero_diagnostics`
- Fallback: `report_remaining_diagnostics`

Treat every lint warning as an error in files you touched.

## R012 - Frontend DRY components (MEDIUM)

Scope: `**/*.{vue,tsx,jsx,svelte}`. When: `file.language == TYPESCRIPT`.

- Do: `reuse_components_with_variants`, `search_before_creating_component`
- Never: `duplicate_component_markup`
- Verify: `shared_component_used`
- Fallback: `add_variant_prop`

Components sharing most structure are one component with variants.

## R013 - Future-extensible APIs (MEDIUM)

Scope: `**/*.{ts,java,lua}`. When: `agent.designing_api`.

- Do: `design_for_extension`, `add_optional_parameters`
- Never: `break_existing_call_sites`
- Verify: `api_backward_compatible`
- Fallback: `offer_options_object`

Design APIs so future options do not break call sites.

## R014 - Git worktree in monorepo (LOW)

Scope: `**/*`. When: `agent.creating_worktree`.

- Do: `use_project_scoped_worktree`
- Never: `share_worktree_across_projects`
- Verify: `worktree_isolated`
- Fallback: `ask_before_worktree`

Each sub-project has its own git repository and worktree.

## R015 - Hytale UI references (LOW)

Scope: `**/*.{ui,lang}`. When: `file.language == HYTALE`.

- Do: `read_ui_docs_before_editing`
- Never: `guess_ui_syntax`
- Verify: `ui_docs_consulted`
- Fallback: `defer_edit`

Read the Hytale UI documentation before editing UI or language files.

## R016 - Logging discipline (MEDIUM)

Scope: `**/*.{ts,lua}`. When: `agent.logging`.

- Do: `log_decisions`, `log_io`, `log_full_error_objects`
- Never: `log_secrets`, `log_only_error_messages`
- Verify: `logs_have_context`
- Fallback: `use_stable_module_prefix`

Log decisions, effective configuration, I/O and full error objects.

## R017 - Mandatory unit tests (CRITICAL)

Scope: `platform/**/*.ts`. When: `platform.has_test_runner`.

- Do: `add_unit_tests_for_new_logic`
- Never: `skip_tests`
- Verify: `tests_pass`
- Fallback: `document_coverage_gap`

When a test runner exists, new logic requires unit tests.

## R018 - Maximum file size (HIGH)

Scope: `**/*.{ts,tsx,java,lua}`. When: `file.line_count > 700`.

- Do: `split_into_focused_modules`
- Never: `create_god_files`
- Verify: `files_under_limit`
- Fallback: `extract_by_domain`

No source file may exceed 700 lines.

## R019 - Nested API DTOs (HIGH)

Scope: `**/dto/**/*.ts`. When: `agent.designing_api_response`.

- Do: `use_semantic_nesting`, `avoid_flat_root_fields`
- Never: `flatten_aggregate_responses`
- Verify: `dto_nested`
- Fallback: `group_by_concept`

Aggregate API responses use semantic nesting, not flat root fields.

## R020 - Never doubt user actions (MEDIUM)

Scope: `**/*`. When: `user.reports_action`.

- Do: `believe_user_report`, `proceed_from_it`
- Never: `question_completed_action`
- Verify: `no_second_guessing`
- Fallback: `ask_only_if_contradictory`

Believe the user when they say they performed an action.

## R021 - Never touch production (CRITICAL)

Scope: `**/*`. When: `agent.considering_remote`.

- Do: `stay_local`, `ask_permission_for_remote`
- Never: `touch_production`, `use_production_credentials`
- Verify: `no_production_access`
- Fallback: `stop_and_ask`

You are always local. Never access production systems or credentials.

## R022 - No alignment padding (LOW)

Scope: `**/*.{ts,tsx,java,lua}`. When: `agent.formatting_code`.

- Do: `use_normal_spacing`
- Never: `add_alignment_padding`
- Verify: `formatting_clean`
- Fallback: `follow_formatter`

Do not use alignment spaces in assignments or comments.

## R023 - No catch any (HIGH)

Scope: `**/*.ts`. When: `file.language == TYPESCRIPT`.

- Do: `use_unknown_catch_variable`
- Never: `annotate_catch_as_any`
- Verify: `catches_typed`
- Fallback: `use_type_guard`

Catch variables use unknown, not any.

## R024 - No typographic dash (CRITICAL)

Scope: `**/*`. When: `agent.writing_text`.

- Do: `use_ascii_hyphen`
- Never: `emit_typographic_dash`
- Verify: `ascii_only`
- Fallback: `ask_before_symbol`

Never use the typographic dash in code or content.

## R025 - No file protocol dependencies (HIGH)

Scope: `**/package.json`. When: `agent.adding_dependency`.

- Do: `use_workspaces_or_git`, `use_published_package`
- Never: `use_file_protocol_outside_repo`
- Verify: `dependencies_resolvable`
- Fallback: `ask_for_package_source`

Never use file: or link: for packages outside the sub-project git repo.

## R026 - No global pollution (MEDIUM)

Scope: `**/*.{ts,tsx,java,lua}`. When: `agent.creating_helper`.

- Do: `use_namespaces_or_classes`, `use_object_literals`
- Never: `add_globals`
- Verify: `no_global_leaks`
- Fallback: `wrap_in_module`

Use namespaces, classes or object literals instead of globals.

## R027 - No inline type imports (MEDIUM)

Scope: `**/*.{ts,tsx}`. When: `file.language == TYPESCRIPT`.

- Do: `use_import_type_at_top`
- Never: `use_inline_type_import`
- Verify: `imports_canonical`
- Fallback: `move_to_top`

Use import type at the top of the file.

## R028 - No local file imports (HIGH)

Scope: `**/package.json`. When: `agent.adding_dependency`.

- Do: `use_workspaces_or_published_package`
- Never: `use_local_file_dependency`
- Verify: `dependencies_portable`
- Fallback: `ask_for_constraint`

Do not use file: or link: local dependencies.

## R029 - No native browser dialogs (MEDIUM)

Scope: `**/*.{vue,tsx,jsx,svelte}`. When: `file.language == TYPESCRIPT`.

- Do: `use_project_dialogs`
- Never: `use_alert_confirm_prompt`
- Verify: `dialogs_consistent`
- Fallback: `reuse_existing_dialog`

Use project dialogs instead of native browser dialogs.

## R030 - No non-route files in routes (HIGH)

Scope: `**/routes/**/*.ts`. When: `file.in_routes_folder`.

- Do: `keep_only_route_handlers`
- Never: `place_helpers_in_routes`
- Verify: `routes_clean`
- Fallback: `move_to_shared_folder`

Route folders contain only route handler files.

## R031 - No re-export on refactor (HIGH)

Scope: `**/*.{ts,tsx,java,lua}`. When: `agent.refactoring`.

- Do: `update_all_call_sites`
- Never: `create_compatibility_reexport`
- Verify: `no_dead_exports`
- Fallback: `search_all_references`

Update every use when refactoring; never add compatibility re-exports.

## R032 - No technical user copy (MEDIUM)

Scope: `**/*`. When: `agent.writing_user_copy`.

- Do: `use_plain_language`
- Never: `use_jargon_in_user_copy`
- Verify: `copy_readable`
- Fallback: `rewrite_for_user`

Keep user-facing copy free of technical jargon.

## R033 - No test artifacts in src (HIGH)

Scope: `platform/**/*.ts`. When: `platform.has_test_runner`.

- Do: `keep_tests_in_tests_folder`
- Never: `place_fake_or_mock_in_src`
- Verify: `src_clean`
- Fallback: `move_to_tests`

Test artifacts belong in the tests folder, never in src.

## R034 - No unsafe type assertions (HIGH)

Scope: `**/*.{ts,tsx,vue}`. When: `file.language == TYPESCRIPT`.

- Do: `use_type_guards`, `use_augmentation`
- Never: `use_as_any_cast`
- Verify: `assertions_safe`
- Fallback: `use_data_helper`

Avoid escape-hatch assertions such as as any.

## R035 - No Vue this as any (MEDIUM)

Scope: `**/*.vue`. When: `file.language == VUE`.

- Do: `use_data_or_augmentation`
- Never: `cast_this_as_any`
- Verify: `vue_typed`
- Fallback: `declare_component_data`

Do not cast this as any in Vue components.

## R036 - Object literal one prop per line (LOW)

Scope: `**/*.{ts,tsx,js,jsx,vue}`. When: `agent.writing_object`.

- Do: `one_property_per_line`
- Never: `pack_multiple_properties_per_line`
- Verify: `objects_readable`
- Fallback: `follow_formatter`

Object literals with two or more properties place each on its own line.

## R037 - Preserve existing comments (HIGH)

Scope: `**/*`. When: `agent.editing_file`.

- Do: `keep_existing_comments`
- Never: `delete_or_rewrite_comments`
- Verify: `comments_preserved`
- Fallback: `restore_if_removed`

Preserve existing code comments during refactors.

## R038 - Portuguese accentuation (MEDIUM)

Scope: `**/*`. When: `user.language == PT_BR`.

- Do: `use_proper_accents`
- Never: `output_unaccented_portuguese`
- Verify: `accents_correct`
- Fallback: `ask_for_wording`

Brazilian Portuguese strings require proper accents.

## R039 - Read project docs (HIGH)

Scope: `**/*`. When: `agent.entering_subproject`.

- Do: `read_project_docs_before_editing`
- Never: `guess_conventions`
- Verify: `docs_consulted`
- Fallback: `defer_edit`

Read the sub-project documentation before editing its code.

## R040 - Respect user changes (CRITICAL)

Scope: `**/*`. When: `agent.editing_file`.

- Do: `preserve_recent_user_edits`
- Never: `revert_user_changes`
- Verify: `user_edits_intact`
- Fallback: `ask_before_overwrite`

Never revert or override recent user edits.

## R041 - Route file naming (MEDIUM)

Scope: `**/routes/**/*.ts`. When: `agent.creating_route`.

- Do: `use_nested_folders`, `avoid_hyphenated_filenames`
- Never: `use_hyphens_in_route_filenames`
- Verify: `route_files_nested`
- Fallback: `rename_to_folder`

File-based routing uses nested folders instead of hyphens in filenames.

## R042 - Route path conventions (MEDIUM)

Scope: `**/routes/**/*.ts`. When: `agent.creating_route`.

- Do: `use_versioning`, `use_plural_names`, `use_kebab_case`
- Never: `use_inconsistent_paths`
- Verify: `paths_consistent`
- Fallback: `follow_sibling_routes`

Route paths use versioning, pluralization and kebab-case for proper names.

## R043 - Sequelize index naming (LOW)

Scope: `**/database/**/*.ts`. When: `file.in_database_folder`.

- Do: `use_table_prefix`, `omit_idx_and_uq`
- Never: `use_idx_prefix`
- Verify: `index_names_consistent`
- Fallback: `follow_existing_indexes`

Sequelize index names use reduced table prefixes without idx_ or uq_.

## R044 - Sequelize ORM only (CRITICAL)

Scope: `**/database/**/*.ts`. When: `file.in_database_folder`.

- Do: `use_sequelize_orm`, `use_transactions`
- Never: `write_raw_sql`
- Verify: `no_raw_sql`
- Fallback: `use_model_query`

Application code uses Sequelize ORM, never raw SQL.

## R045 - TSDoc required (MEDIUM)

Scope: `**/*.{ts,tsx,java}`. When: `agent.exporting_function`.

- Do: `add_tsdoc_to_exports`
- Never: `export_undocumented_function`
- Verify: `exports_documented`
- Fallback: `write_doc_comment`

Exported functions and methods require TSDoc.

