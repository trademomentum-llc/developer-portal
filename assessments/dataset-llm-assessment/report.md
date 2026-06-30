# Dataset LLM Workflow Assessment

Workspace root: `/Users/nexus1/Projects`

## Summary

- Unique first-party datasets analyzed: 54
- Duplicated copies collapsed by file hash: 50
- Fit distribution: high=0, medium=38, low=16
- Supported formats were profiled directly where possible; XLSX files were inspected at workbook-metadata level.

## Format Breakdown

- `jsonl_event_log`: 2
- `jsonl_flat`: 9
- `jsonl_nested`: 8
- `jsonl_placeholder_training`: 7
- `markdown_table_disguised_as_csv`: 1
- `sqlite_db`: 1
- `tabular_csv`: 23
- `xlsx_workbook`: 3

## Biggest Datasets

- `71.2MB` omni/multimodal-agent-builder/train-test-validate/ML-Validation/open_images_validation_asr.jsonl
- `1.7MB` omni/ihep-app/ihep/docs/runbooks/PZdataset.xlsx
- `1.4MB` omni/security-workspace/logs/security_fragments.db
- `241.8KB` omni/ihep-app/ihep/training_datasets/risk_prediction/processed/risk_training.jsonl
- `229.2KB` omni/ihep-app/ihep/training_datasets/clinical/processed/clinical_training.jsonl
- `227.3KB` omni/ihep-app/ihep/training_datasets/adherence/processed/adherence_training.jsonl
- `169.0KB` omni/ihep-app/ihep/training_datasets/mental_health/processed/mental_health_training.jsonl
- `137.7KB` omni/ihep-app/ihep/training_datasets/conversational/raw/patient_dialogues.jsonl
- `135.0KB` omni/ihep-app/ihep/training_datasets/social_determinants/raw/sdoh_interventions.jsonl
- `127.8KB` omni/ihep-app/ihep/training_datasets/risk_prediction/raw/risk_stratification_training.jsonl

## Dataset Inventory

| Path | Kind | Size | Shape | Fit | Workflow | Notes |
| --- | --- | ---: | --- | --- | --- | --- |
| omni/PixelAgent/data/adb_commands.jsonl | jsonl_nested | 76.1KB | 119 rows x ~7 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. |
| omni/PixelAgent/data/calibration.jsonl | jsonl_flat | 8.7KB | 38 rows x ~11 flattened cols | medium | THE KNOCKOUT EDA | - |
| omni/PixelAgent/data/hypotheses.jsonl | jsonl_nested | 3.2KB | 3 rows x ~29 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. |
| omni/PixelAgent/data/state_log.jsonl | jsonl_event_log | 2.0KB | 7 rows x ~33 flattened cols | medium | THE TIME TRAVELER | Event log with timestamps; requires time-series/event aggregation before classic EDA. |
| omni/PixelAgent/data/state_log_v2.jsonl | jsonl_event_log | 2.5KB | 7 rows x ~30 flattened cols | medium | THE TIME TRAVELER | Event log with timestamps; requires time-series/event aggregation before classic EDA. |
| omni/PixelAgent/data/telemetry_history.jsonl | jsonl_nested | 10.9KB | 26 rows x ~12 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. |
| omni/cali/jeff/tri-mar/Trident_Markets_Operating_Budget.xlsx | xlsx_workbook | 10.7KB | 1 sheets | medium | THE KNOCKOUT EDA | Workbook metadata parsed via ZIP/XML; cell-level profiling not run. |
| omni/distributed-ai-cluster/proposals/doe-hep/potential-co-pis.csv | tabular_csv | 5.9KB | 14 rows x 11 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Top missingness: Phone=71%. Contains direct contact fields; do not send blindly to third-party LLMs. |
| omni/eight-layer-pqc/docs/00637599.csv | tabular_csv | 1.7KB | 5 rows x 4 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Duplicate copies: 1 |
| omni/eight-layer-pqc/docs/102737b2.csv | tabular_csv | 1.3KB | 5 rows x 4 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Duplicate copies: 1 |
| omni/eight-layer-pqc/docs/53919eaa.csv | tabular_csv | 1.7KB | 6 rows x 4 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Duplicate copies: 1 |
| omni/eight-layer-pqc/docs/5ef3c051.csv | tabular_csv | 1.5KB | 5 rows x 4 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Duplicate copies: 1 |
| omni/eight-layer-pqc/docs/cae830c0.csv | tabular_csv | 1.5KB | 6 rows x 4 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. Duplicate copies: 1 |
| omni/health-insights/docs/income-generation-for-ihep-sustainability-v1-evals.csv | tabular_csv | 96.1KB | 7 rows x 3 cols | medium | THE DATA JANITOR | Contains long-form text blobs; classic numeric EDA is only a partial fit. Very small sample size; inferential workflows are weak. |
| omni/ihep-app/ihep/docs/compliance/IHEP_Compliance_Mapping.xlsx | xlsx_workbook | 7.0KB | 1 sheets | medium | THE KNOCKOUT EDA | Workbook metadata parsed via ZIP/XML; cell-level profiling not run. Duplicate copies: 2 |
| omni/ihep-app/ihep/docs/runbooks/PZdataset.xlsx | xlsx_workbook | 1.7MB | 1 sheets | medium | THE KNOCKOUT EDA | Workbook metadata parsed via ZIP/XML; cell-level profiling not run. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/adherence/processed/adherence_training.jsonl | jsonl_placeholder_training | 227.3KB | 400 rows x ~9 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/adherence/raw/medication_adherence_patterns.jsonl | jsonl_flat | 90.0KB | 26 rows x ~10 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/clinical/processed/clinical_training.jsonl | jsonl_placeholder_training | 229.2KB | 400 rows x ~9 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/clinical/raw/chronic_disease_management.jsonl | jsonl_flat | 94.5KB | 24 rows x ~10 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/clinical/raw/hiv_care_continuum.jsonl | jsonl_flat | 83.0KB | 36 rows x ~10 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/conversational/raw/patient_dialogues.jsonl | jsonl_nested | 137.7KB | 24 rows x ~14 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/evaluation/raw/bias_eval.jsonl | jsonl_nested | 9.8KB | 24 rows x ~8 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. Duplicate rows observed in sample. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/evaluation/raw/clinical_eval.jsonl | jsonl_nested | 11.4KB | 30 rows x ~6 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. Duplicate rows observed in sample. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/evaluation/raw/safety_eval.jsonl | jsonl_nested | 10.5KB | 24 rows x ~8 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. Duplicate rows observed in sample. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/financial_health/raw/financial_navigation.jsonl | jsonl_flat | 59.8KB | 10 rows x ~10 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/mental_health/processed/mental_health_training.jsonl | jsonl_placeholder_training | 169.0KB | 300 rows x ~10 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/mental_health/raw/mental_health_assessments.jsonl | jsonl_flat | 120.6KB | 24 rows x ~17 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/processed_for_training/train.jsonl | jsonl_placeholder_training | 490.0B | 10 rows x ~1 flattened cols | low | THE DATA JANITOR | Text rows are synthetic placeholders rather than meaningful training content. Validation split is byte-identical to the training file. Duplicate copies: 5 |
| omni/ihep-app/ihep/training_datasets/risk_prediction/processed/risk_training.jsonl | jsonl_placeholder_training | 241.8KB | 200 rows x ~22 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Repeated patient IDs detected in sample/full file (100 repeated IDs). Contains patient-profile features; treat as sensitive even if synthetic. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/risk_prediction/raw/risk_stratification_training.jsonl | jsonl_flat | 127.8KB | 16 rows x ~11 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/social_determinants/processed/sdoh_training.jsonl | jsonl_placeholder_training | 15.2KB | 24 rows x ~9 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/social_determinants/raw/sdoh_interventions.jsonl | jsonl_flat | 135.0KB | 22 rows x ~10 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/wearables/processed/wearables_training.jsonl | jsonl_placeholder_training | 125.2KB | 200 rows x ~8 flattened cols | low | THE DATA JANITOR | Outputs are placeholders rather than labels or gold responses. All sampled outputs are identical. Duplicate copies: 2 |
| omni/ihep-app/ihep/training_datasets/wearables/raw/wearable_data_interpretation.jsonl | jsonl_flat | 53.7KB | 10 rows x ~11 flattened cols | medium | THE KNOCKOUT EDA | Duplicate copies: 2 |
| omni/jarmacz.com/aesthetic_errors_analysis.csv | tabular_csv | 2.6KB | 10 rows x 5 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_0fd1056b-2f0e-4482-8062-06d17b804e2d_closure_ledger.csv | tabular_csv | 484.0B | 2 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_29df166a-9386-40b5-8326-9c73901d292f_closure_ledger.csv | tabular_csv | 289.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_3cd019bc-a7e4-49fe-9e61-3ef61a92c5d8_closure_ledger.csv | tabular_csv | 275.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_52a3d7d1-bbfe-4a61-af73-8ce9a8a8178e_closure_ledger.csv | tabular_csv | 289.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_5571af03-3faf-49c9-90cc-4d48bff1f789_closure_ledger.csv | tabular_csv | 289.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_55d05e02-1e92-4475-aa90-3b43e7b7ffcb_closure_ledger.csv | tabular_csv | 484.0B | 2 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_5636504e-c1a9-4b56-9626-4042366de181_closure_ledger.csv | tabular_csv | 275.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_a8086a5d-eee8-4ef1-87cb-e75b47dcab64_closure_ledger.csv | tabular_csv | 484.0B | 2 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_adcc1803-9ada-4c0a-8167-054a84806bd1_closure_ledger.csv | tabular_csv | 289.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_b0c51747-b0ac-4db7-a368-6f672608c469_closure_ledger.csv | tabular_csv | 289.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_bdde8e40-8320-4d2b-99ec-0b7c86bc7b49_closure_ledger.csv | tabular_csv | 484.0B | 2 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_d277871d-13f1-4951-8a3f-e04a5393302e_closure_ledger.csv | tabular_csv | 275.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_df557fb6-f6e9-4641-90d3-582c5677284a_closure_ledger.csv | tabular_csv | 484.0B | 2 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_e876a125-b25e-49dd-86dd-b45cb399776f_closure_ledger.csv | tabular_csv | 275.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/agent_fed4022a-2c6f-45f0-89f7-0e51a58a5079_closure_ledger.csv | tabular_csv | 275.0B | 1 rows x 7 cols | medium | THE DATA JANITOR | Mostly categorical/text fields; plotting and correlation steps are limited. Very small sample size; inferential workflows are weak. |
| omni/multimodal-agent-builder/curated_hf_datasets.csv | markdown_table_disguised_as_csv | 2.3KB | - | low | THE DATA JANITOR | File is a Markdown table saved with .csv extension. |
| omni/multimodal-agent-builder/train-test-validate/ML-Validation/open_images_validation_asr.jsonl | jsonl_nested | 71.2MB | 41691 rows x ~4 flattened cols | low | THE DATA JANITOR | Nested list-heavy schema; standard tabular primitives do not apply directly. |
| omni/security-workspace/logs/security_fragments.db | sqlite_db | 1.4MB | 3 tables | medium | THE KNOCKOUT EDA | Table sizes: audit_log=3687, fragments=3692, synergies=0 Requires table selection and joins before any LLM-style workflow. |
