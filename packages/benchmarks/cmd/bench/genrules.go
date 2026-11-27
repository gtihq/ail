package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ruleDefinition is the single source of truth for both instruction formats.
type ruleDefinition struct {
	id       string
	title    string
	priority string
	scope    string
	when     string
	do       string
	never    string
	verify   string
	fallback string
	prose    string
}

// workspaceRules mirrors the 45 rules in the Projetos workspace .agents/memories folder.
var workspaceRules = []ruleDefinition{
	{"R001", "Admin panel when applicable", "HIGH", "platform/**", "project.has_admin_panel", "add_admin_listing add_admin_detail add_admin_create_edit use_accented_labels", "ship_adminless_data_feature", "admin_routes_tested", "ask_for_admin_scope", "Data features in projects with an admin panel must include listing, detail and edit views in the same change."},
	{"R002", "Auto-discovery extensibility", "CRITICAL", "**/*.ts", "agent.adding_feature", "register_by_convention use_folder_conventions", "edit_central_registry_file", "feature_discoverable", "search_convention_first", "A new feature must be added by creating files only, never by editing a central registry."},
	{"R003", "Cluster-safe atomic operations", "CRITICAL", "platform/**/*.ts", "project.runs_in_cluster", "use_shared_store_for_cross_request_state use_atomic_operations", "rely_on_local_memory_state", "operations_atomic", "choose_cluster_safe_primitive", "Systems run in multi-node clusters. Shared state must be external and operations atomic."},
	{"R004", "Code sections", "MEDIUM", "**/*.{ts,tsx,java,lua}", "file.language == TYPESCRIPT", "use_region_markers", "use_separator_comment_lines", "sections_readable", "keep_existing_structure", "Use region markers for code sections and never separator lines."},
	{"R005", "Commit message language", "MEDIUM", "**/*", "agent.about_to_commit", "inspect_repo_history match_dominant_language add_agent_trailer", "assume_fixed_language", "message_matches_history", "ask_user", "Commit message language follows the repository history, with the agent trailer."},
	{"R006", "Directory consistency", "HIGH", "**/*", "agent.creating_concept", "use_one_directory_per_concept search_before_creating", "duplicate_concept_implementations", "single_implementation", "extend_existing_implementation", "One concept lives in one directory. Search before creating a duplicate."},
	{"R007", "Behavioral coding guidelines", "HIGH", "**/*", "agent.about_to_code", "think_before_coding prefer_simplicity make_surgical_changes define_success_criteria", "overengineer improve_unrelated_code", "criteria_verified", "present_tradeoffs", "Think before coding, prefer simplicity, make surgical changes and define verifiable success."},
	{"R008", "Domain error hierarchy", "HIGH", "**/*.ts", "file.language == TYPESCRIPT", "use_typed_domain_errors use_base_error_class", "throw_generic_error", "errors_typed", "report_missing_error_type", "Business errors are typed and structured, never generic errors."},
	{"R009", "Integration and E2E tests", "HIGH", "**/routes/**/*.ts", "project.has_http_routes", "add_integration_tests add_e2e_tests_for_routes", "ship_route_without_tests", "routes_covered", "document_test_gap", "Features with HTTP routes or UI flows require integration and E2E tests."},
	{"R010", "Extract repeated patterns", "MEDIUM", "**/*.{ts,tsx,java,lua}", "agent.noticing_repetition", "search_existing_helpers extract_at_third_repetition", "copy_paste_three_times", "no_triple_duplication", "name_by_function", "The same snippet in three places must be extracted into a shared helper."},
	{"R011", "Fix all lint warnings", "HIGH", "**/*", "agent.editing_file", "run_linter fix_every_warning_in_touched_files", "ignore_warnings", "zero_diagnostics", "report_remaining_diagnostics", "Treat every lint warning as an error in files you touched."},
	{"R012", "Frontend DRY components", "MEDIUM", "**/*.{vue,tsx,jsx,svelte}", "file.language == TYPESCRIPT", "reuse_components_with_variants search_before_creating_component", "duplicate_component_markup", "shared_component_used", "add_variant_prop", "Components sharing most structure are one component with variants."},
	{"R013", "Future-extensible APIs", "MEDIUM", "**/*.{ts,java,lua}", "agent.designing_api", "design_for_extension add_optional_parameters", "break_existing_call_sites", "api_backward_compatible", "offer_options_object", "Design APIs so future options do not break call sites."},
	{"R014", "Git worktree in monorepo", "LOW", "**/*", "agent.creating_worktree", "use_project_scoped_worktree", "share_worktree_across_projects", "worktree_isolated", "ask_before_worktree", "Each sub-project has its own git repository and worktree."},
	{"R015", "Hytale UI references", "LOW", "**/*.{ui,lang}", "file.language == HYTALE", "read_ui_docs_before_editing", "guess_ui_syntax", "ui_docs_consulted", "defer_edit", "Read the Hytale UI documentation before editing UI or language files."},
	{"R016", "Logging discipline", "MEDIUM", "**/*.{ts,lua}", "agent.logging", "log_decisions log_io log_full_error_objects", "log_secrets log_only_error_messages", "logs_have_context", "use_stable_module_prefix", "Log decisions, effective configuration, I/O and full error objects."},
	{"R017", "Mandatory unit tests", "CRITICAL", "platform/**/*.ts", "platform.has_test_runner", "add_unit_tests_for_new_logic", "skip_tests", "tests_pass", "document_coverage_gap", "When a test runner exists, new logic requires unit tests."},
	{"R018", "Maximum file size", "HIGH", "**/*.{ts,tsx,java,lua}", "file.line_count > 700", "split_into_focused_modules", "create_god_files", "files_under_limit", "extract_by_domain", "No source file may exceed 700 lines."},
	{"R019", "Nested API DTOs", "HIGH", "**/dto/**/*.ts", "agent.designing_api_response", "use_semantic_nesting avoid_flat_root_fields", "flatten_aggregate_responses", "dto_nested", "group_by_concept", "Aggregate API responses use semantic nesting, not flat root fields."},
	{"R020", "Never doubt user actions", "MEDIUM", "**/*", "user.reports_action", "believe_user_report proceed_from_it", "question_completed_action", "no_second_guessing", "ask_only_if_contradictory", "Believe the user when they say they performed an action."},
	{"R021", "Never touch production", "CRITICAL", "**/*", "agent.considering_remote", "stay_local ask_permission_for_remote", "touch_production use_production_credentials", "no_production_access", "stop_and_ask", "You are always local. Never access production systems or credentials."},
	{"R022", "No alignment padding", "LOW", "**/*.{ts,tsx,java,lua}", "agent.formatting_code", "use_normal_spacing", "add_alignment_padding", "formatting_clean", "follow_formatter", "Do not use alignment spaces in assignments or comments."},
	{"R023", "No catch any", "HIGH", "**/*.ts", "file.language == TYPESCRIPT", "use_unknown_catch_variable", "annotate_catch_as_any", "catches_typed", "use_type_guard", "Catch variables use unknown, not any."},
	{"R024", "No typographic dash", "CRITICAL", "**/*", "agent.writing_text", "use_ascii_hyphen", "emit_typographic_dash", "ascii_only", "ask_before_symbol", "Never use the typographic dash in code or content."},
	{"R025", "No file protocol dependencies", "HIGH", "**/package.json", "agent.adding_dependency", "use_workspaces_or_git use_published_package", "use_file_protocol_outside_repo", "dependencies_resolvable", "ask_for_package_source", "Never use file: or link: for packages outside the sub-project git repo."},
	{"R026", "No global pollution", "MEDIUM", "**/*.{ts,tsx,java,lua}", "agent.creating_helper", "use_namespaces_or_classes use_object_literals", "add_globals", "no_global_leaks", "wrap_in_module", "Use namespaces, classes or object literals instead of globals."},
	{"R027", "No inline type imports", "MEDIUM", "**/*.{ts,tsx}", "file.language == TYPESCRIPT", "use_import_type_at_top", "use_inline_type_import", "imports_canonical", "move_to_top", "Use import type at the top of the file."},
	{"R028", "No local file imports", "HIGH", "**/package.json", "agent.adding_dependency", "use_workspaces_or_published_package", "use_local_file_dependency", "dependencies_portable", "ask_for_constraint", "Do not use file: or link: local dependencies."},
	{"R029", "No native browser dialogs", "MEDIUM", "**/*.{vue,tsx,jsx,svelte}", "file.language == TYPESCRIPT", "use_project_dialogs", "use_alert_confirm_prompt", "dialogs_consistent", "reuse_existing_dialog", "Use project dialogs instead of native browser dialogs."},
	{"R030", "No non-route files in routes", "HIGH", "**/routes/**/*.ts", "file.in_routes_folder", "keep_only_route_handlers", "place_helpers_in_routes", "routes_clean", "move_to_shared_folder", "Route folders contain only route handler files."},
	{"R031", "No re-export on refactor", "HIGH", "**/*.{ts,tsx,java,lua}", "agent.refactoring", "update_all_call_sites", "create_compatibility_reexport", "no_dead_exports", "search_all_references", "Update every use when refactoring; never add compatibility re-exports."},
	{"R032", "No technical user copy", "MEDIUM", "**/*", "agent.writing_user_copy", "use_plain_language", "use_jargon_in_user_copy", "copy_readable", "rewrite_for_user", "Keep user-facing copy free of technical jargon."},
	{"R033", "No test artifacts in src", "HIGH", "platform/**/*.ts", "platform.has_test_runner", "keep_tests_in_tests_folder", "place_fake_or_mock_in_src", "src_clean", "move_to_tests", "Test artifacts belong in the tests folder, never in src."},
	{"R034", "No unsafe type assertions", "HIGH", "**/*.{ts,tsx,vue}", "file.language == TYPESCRIPT", "use_type_guards use_augmentation", "use_as_any_cast", "assertions_safe", "use_data_helper", "Avoid escape-hatch assertions such as as any."},
	{"R035", "No Vue this as any", "MEDIUM", "**/*.vue", "file.language == VUE", "use_data_or_augmentation", "cast_this_as_any", "vue_typed", "declare_component_data", "Do not cast this as any in Vue components."},
	{"R036", "Object literal one prop per line", "LOW", "**/*.{ts,tsx,js,jsx,vue}", "agent.writing_object", "one_property_per_line", "pack_multiple_properties_per_line", "objects_readable", "follow_formatter", "Object literals with two or more properties place each on its own line."},
	{"R037", "Preserve existing comments", "HIGH", "**/*", "agent.editing_file", "keep_existing_comments", "delete_or_rewrite_comments", "comments_preserved", "restore_if_removed", "Preserve existing code comments during refactors."},
	{"R038", "Portuguese accentuation", "MEDIUM", "**/*", "user.language == PT_BR", "use_proper_accents", "output_unaccented_portuguese", "accents_correct", "ask_for_wording", "Brazilian Portuguese strings require proper accents."},
	{"R039", "Read project docs", "HIGH", "**/*", "agent.entering_subproject", "read_project_docs_before_editing", "guess_conventions", "docs_consulted", "defer_edit", "Read the sub-project documentation before editing its code."},
	{"R040", "Respect user changes", "CRITICAL", "**/*", "agent.editing_file", "preserve_recent_user_edits", "revert_user_changes", "user_edits_intact", "ask_before_overwrite", "Never revert or override recent user edits."},
	{"R041", "Route file naming", "MEDIUM", "**/routes/**/*.ts", "agent.creating_route", "use_nested_folders avoid_hyphenated_filenames", "use_hyphens_in_route_filenames", "route_files_nested", "rename_to_folder", "File-based routing uses nested folders instead of hyphens in filenames."},
	{"R042", "Route path conventions", "MEDIUM", "**/routes/**/*.ts", "agent.creating_route", "use_versioning use_plural_names use_kebab_case", "use_inconsistent_paths", "paths_consistent", "follow_sibling_routes", "Route paths use versioning, pluralization and kebab-case for proper names."},
	{"R043", "Sequelize index naming", "LOW", "**/database/**/*.ts", "file.in_database_folder", "use_table_prefix omit_idx_and_uq", "use_idx_prefix", "index_names_consistent", "follow_existing_indexes", "Sequelize index names use reduced table prefixes without idx_ or uq_."},
	{"R044", "Sequelize ORM only", "CRITICAL", "**/database/**/*.ts", "file.in_database_folder", "use_sequelize_orm use_transactions", "write_raw_sql", "no_raw_sql", "use_model_query", "Application code uses Sequelize ORM, never raw SQL."},
	{"R045", "TSDoc required", "MEDIUM", "**/*.{ts,tsx,java}", "agent.exporting_function", "add_tsdoc_to_exports", "export_undocumented_function", "exports_documented", "write_doc_comment", "Exported functions and methods require TSDoc."},
}

func words(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// genrulesCommand generates both instruction representations from one definition list.
func genrulesCommand(args []string) {
	flags := flag.NewFlagSet("genrules", flag.ExitOnError)
	flags.Parse(args)

	var ail strings.Builder
	ail.WriteString("META { id: \"workspace-rules\" version: \"0.3\" priority: HIGH }\n\n")
	for _, rule := range workspaceRules {
		fmt.Fprintf(&ail, "RULE %s {\n", rule.id)
		fmt.Fprintf(&ail, "  priority: %s\n", rule.priority)
		fmt.Fprintf(&ail, "  scope: %q\n", rule.scope)
		fmt.Fprintf(&ail, "  when { %s }\n", words(rule.when))
		fmt.Fprintf(&ail, "  do { %s }\n", words(rule.do))
		fmt.Fprintf(&ail, "  never { %s }\n", words(rule.never))
		fmt.Fprintf(&ail, "  verify { %s }\n", words(rule.verify))
		fmt.Fprintf(&ail, "  fallback { %s }\n", words(rule.fallback))
		ail.WriteString("}\n\n")
	}

	var md strings.Builder
	md.WriteString("# Workspace agent rules - conventional Markdown baseline\n\n")
	md.WriteString("Generated from the same definitions as `.ail/rules/project-rules.ail`. Keep both files in sync.\n\n")
	for _, rule := range workspaceRules {
		fmt.Fprintf(&md, "## %s - %s (%s)\n\n", rule.id, rule.title, rule.priority)
		fmt.Fprintf(&md, "Scope: `%s`. When: `%s`.\n\n", rule.scope, words(rule.when))
		fmt.Fprintf(&md, "- Do: `%s`\n", strings.Join(strings.Fields(rule.do), "`, `"))
		fmt.Fprintf(&md, "- Never: `%s`\n", strings.Join(strings.Fields(rule.never), "`, `"))
		fmt.Fprintf(&md, "- Verify: `%s`\n", strings.Join(strings.Fields(rule.verify), "`, `"))
		fmt.Fprintf(&md, "- Fallback: `%s`\n\n", strings.Join(strings.Fields(rule.fallback), "`, `"))
		md.WriteString(rule.prose + "\n\n")
	}

	if err := os.WriteFile(filepath.Join(".ail", "rules", "project-rules.ail"), []byte(ail.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join("packages", "benchmarks", "MARKDOWN_BASELINE.md"), []byte(md.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("generated %d rules in both formats\n", len(workspaceRules))
}
