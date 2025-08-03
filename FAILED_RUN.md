SAST Scanning	Run Gosec SAST	﻿2025-08-03T18:58:05.5037687Z ##[group]Run echo "🔒 Running Gosec static application security testing"
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5038247Z [36;1mecho "🔒 Running Gosec static application security testing"[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5038578Z [36;1m[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5038759Z [36;1m# SECURITY: Install gosec[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5039160Z [36;1mgo install github.com/securecodewarrior/github-action-gosec/cmd/gosec@latest[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5039789Z [36;1m[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5039980Z [36;1m# SECURITY: Run comprehensive scan[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5040412Z [36;1mgosec -severity HIGH -confidence medium -fmt sarif -out gosec-results.sarif ./...[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5040813Z [36;1m[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5040987Z [36;1m# SECURITY: Check results[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5041249Z [36;1mif [[ -f "gosec-results.sarif" ]]; then[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5041586Z [36;1m  echo "✅ Gosec scan completed - results saved to SARIF"[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5041884Z [36;1melse[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5042102Z [36;1m  echo "❌ Gosec scan failed to generate results"[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5042451Z [36;1m  exit 1[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5042655Z [36;1mfi[0m
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5074724Z shell: /usr/bin/bash -e {0}
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5074970Z env:
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5075153Z   SEVERITY_THRESHOLD: HIGH
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5075378Z   SCAN_TIMEOUT: 600
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5075579Z   MAX_VULNERABILITIES_HIGH: 0
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5075808Z   MAX_VULNERABILITIES_CRITICAL: 0
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5076037Z ##[endgroup]
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.5143836Z 🔒 Running Gosec static application security testing
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.8211488Z go: github.com/securecodewarrior/github-action-gosec/cmd/gosec@latest: module github.com/securecodewarrior/github-action-gosec/cmd/gosec: git ls-remote -q origin in /home/runner/go/pkg/mod/cache/vcs/7505be6eec20f0b603a234d24c6495998ee4105c45b4ad50d74c42a44cc45adb: exit status 128:
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.8214700Z 	fatal: could not read Username for 'https://github.com': terminal prompts disabled
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.8215812Z Confirm the import path was entered correctly.
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.8216918Z If this is a private repository, see https://golang.org/doc/faq#git_https for additional information.
SAST Scanning	Run Gosec SAST	2025-08-03T18:58:05.8233565Z ##[error]Process completed with exit code 1.
Secret Scanning	Run TruffleHog secret scan	﻿2025-08-03T18:57:42.9411936Z ##[group]Run trufflesecurity/trufflehog@main
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9412297Z with:
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9412524Z   path: ./
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9412731Z   base: master
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9412947Z   head: HEAD
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9413187Z   extra_args: --debug --only-verified --fail
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9413478Z   version: latest
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9413688Z env:
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9413908Z   SEVERITY_THRESHOLD: HIGH
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9414152Z   SCAN_TIMEOUT: 600
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9414408Z   MAX_VULNERABILITIES_HIGH: 0
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9414673Z   MAX_VULNERABILITIES_CRITICAL: 0
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9414933Z ##[endgroup]
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9551090Z ##[group]Run ##########################################
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9551576Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9551945Z [36;1m## ADVANCED USAGE                       ##[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9552291Z [36;1m## Scan by BASE & HEAD user inputs      ##[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9552620Z [36;1m## If BASE == HEAD, exit with error     ##[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9552988Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9553326Z [36;1m# Check if jq is installed, if not, install it[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9553660Z [36;1mif ! command -v jq &> /dev/null[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9553955Z [36;1mthen[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9554212Z [36;1m  echo "jq could not be found, installing..."[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9554574Z [36;1m  apt-get -y update && apt-get install -y jq[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9554899Z [36;1mfi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9555135Z [36;1m[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9555437Z [36;1mgit status >/dev/null  # make sure we are in a git repository[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9555821Z [36;1mif [ -n "$BASE" ] || [ -n "$HEAD" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9556124Z [36;1m  if [ -n "$BASE" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9556467Z [36;1m    base_commit=$(git rev-parse "$BASE" 2>/dev/null) || true[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9556794Z [36;1m  else[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9557021Z [36;1m    base_commit=""[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9557268Z [36;1m  fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9557509Z [36;1m  if [ -n "$HEAD" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9557829Z [36;1m    head_commit=$(git rev-parse "$HEAD" 2>/dev/null) || true[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9558157Z [36;1m  else[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9558369Z [36;1m    head_commit=""[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9558612Z [36;1m  fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9558856Z [36;1m  if [ "$base_commit" == "$head_commit" ] ; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9559803Z [36;1m    echo "::error::BASE and HEAD commits are the same. TruffleHog won't scan anything. Please see documentation (https://github.com/trufflesecurity/trufflehog#octocat-trufflehog-github-action)."[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9560537Z [36;1m    exit 1[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9560751Z [36;1m  fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9560982Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9561288Z [36;1m## Scan commits based on event type     ##[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9561601Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9561878Z [36;1melse[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9562097Z [36;1m  if [ "push" == "push" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9562432Z [36;1m    COMMIT_LENGTH=$(printenv COMMIT_IDS | jq length)[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9562767Z [36;1m    if [ $COMMIT_LENGTH == "0" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9563067Z [36;1m      echo "No commits to scan"[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9563334Z [36;1m      exit 0[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9563553Z [36;1m    fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9563814Z [36;1m    HEAD=24a2a293709dd2cfbac6fd7fdab281d5e725eb7f[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9564310Z [36;1m    if [ c758a654427abb7f970de6f731ab4d6a31246d14 == "0000000000000000000000000000000000000000" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9564761Z [36;1m      BASE=""[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9565200Z [36;1m    else[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9565457Z [36;1m      BASE=c758a654427abb7f970de6f731ab4d6a31246d14[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9565754Z [36;1m    fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9566058Z [36;1m  elif [ "push" == "workflow_dispatch" ] || [ "push" == "schedule" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9566601Z [36;1m    BASE=""[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9566851Z [36;1m    HEAD=""[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9567141Z [36;1m  elif [ "push" == "pull_request" ]; then[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9567425Z [36;1m    BASE=[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9567641Z [36;1m    HEAD=[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9567849Z [36;1m  fi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9568048Z [36;1mfi[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9568277Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9568570Z [36;1m##          Run TruffleHog              ##[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9568858Z [36;1m##########################################[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9569339Z [36;1mdocker run --rm -v .:/tmp -w /tmp \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9569693Z [36;1mghcr.io/trufflesecurity/trufflehog:${VERSION} \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9570034Z [36;1mgit file:///tmp/ \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9570284Z [36;1m--since-commit \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9570523Z [36;1m${BASE:-''} \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9570747Z [36;1m--branch \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9570986Z [36;1m${HEAD:-''} \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9571203Z [36;1m--fail \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9571419Z [36;1m--no-update \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9571650Z [36;1m--github-actions \[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9571894Z [36;1m${ARGS:-''}[0m
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9602787Z shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9603171Z env:
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9603391Z   SEVERITY_THRESHOLD: HIGH
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9603639Z   SCAN_TIMEOUT: 600
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9603883Z   MAX_VULNERABILITIES_HIGH: 0
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9604167Z   MAX_VULNERABILITIES_CRITICAL: 0
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9604427Z   BASE: master
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9604644Z   HEAD: HEAD
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9604876Z   ARGS: --debug --only-verified --fail
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9605246Z   COMMIT_IDS: [
Secret Scanning	Run TruffleHog secret scan	  "24a2a293709dd2cfbac6fd7fdab281d5e725eb7f"
Secret Scanning	Run TruffleHog secret scan	]
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9605577Z   VERSION: latest
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:42.9605798Z ##[endgroup]
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:43.0913651Z ##[error]BASE and HEAD commits are the same. TruffleHog won't scan anything. Please see documentation (https://github.com/trufflesecurity/trufflehog#octocat-trufflehog-github-action).
Secret Scanning	Run TruffleHog secret scan	2025-08-03T18:57:43.0919536Z ##[error]Process completed with exit code 1.
Infrastructure Security Scanning	Run Checkov IaC scan	﻿2025-08-03T18:58:01.2216699Z ##[group]Run bridgecrewio/checkov-action@master
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2217014Z with:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2217179Z   directory: .
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2217438Z   framework: terraform,kubernetes,dockerfile,github_actions
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2217908Z   output_format: sarif
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2218125Z   output_file_path: checkov-results.sarif
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2218399Z   check: CKV_DOCKER_*,CKV_K8S_*,CKV_TF_*,CKV_GHA_*
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2218651Z   soft_fail: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2218822Z   log_level: WARNING
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219000Z   container_user: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219166Z env:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219342Z   SEVERITY_THRESHOLD: HIGH
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219537Z   SCAN_TIMEOUT: 600
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219711Z   MAX_VULNERABILITIES_HIGH: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2219934Z   MAX_VULNERABILITIES_CRITICAL: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2220142Z ##[endgroup]
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.2365527Z ##[command]/usr/bin/docker run --name ghcriobridgecrewiocheckov32457_6a42db --label 11ff2e --workdir /github/workspace --rm -e "SEVERITY_THRESHOLD" -e "SCAN_TIMEOUT" -e "MAX_VULNERABILITIES_HIGH" -e "MAX_VULNERABILITIES_CRITICAL" -e "INPUT_DIRECTORY" -e "INPUT_FRAMEWORK" -e "INPUT_OUTPUT_FORMAT" -e "INPUT_OUTPUT_FILE_PATH" -e "INPUT_CHECK" -e "INPUT_SOFT_FAIL" -e "INPUT_FILE" -e "INPUT_SKIP_CHECK" -e "INPUT_COMPACT" -e "INPUT_QUIET" -e "INPUT_API-KEY" -e "INPUT_OUTPUT_BC_IDS" -e "INPUT_USE_ENFORCEMENT_RULES" -e "INPUT_SKIP_RESULTS_UPLOAD" -e "INPUT_SKIP_FRAMEWORK" -e "INPUT_EXTERNAL_CHECKS_DIRS" -e "INPUT_EXTERNAL_CHECKS_REPOS" -e "INPUT_DOWNLOAD_EXTERNAL_MODULES" -e "INPUT_ENABLE_SECRETS_SCAN_ALL_FILES" -e "INPUT_LOG_LEVEL" -e "INPUT_CONFIG_FILE" -e "INPUT_BASELINE" -e "INPUT_SOFT_FAIL_ON" -e "INPUT_HARD_FAIL_ON" -e "INPUT_CONTAINER_USER" -e "INPUT_DOCKER_IMAGE" -e "INPUT_DOCKERFILE_PATH" -e "INPUT_VAR_FILE" -e "INPUT_GITHUB_PAT" -e "INPUT_TFC_TOKEN" -e "INPUT_TF_REGISTRY_TOKEN" -e "INPUT_CKV_VALIDATE_SECRETS" -e "INPUT_VCS_BASE_URL" -e "INPUT_VCS_USERNAME" -e "INPUT_VCS_TOKEN" -e "INPUT_BITBUCKET_TOKEN" -e "INPUT_BITBUCKET_APP_PASSWORD" -e "INPUT_BITBUCKET_USERNAME" -e "INPUT_REPO_ROOT_FOR_PLAN_ENRICHMENT" -e "INPUT_DEEP_ANALYSIS" -e "INPUT_POLICY_METADATA_FILTER" -e "INPUT_POLICY_METADATA_FILTER_EXCEPTION" -e "INPUT_SKIP_PATH" -e "INPUT_SKIP_CVE_PACKAGE" -e "INPUT_SKIP_DOWNLOAD" -e "INPUT_PRISMA-API-URL" -e "API_KEY_VARIABLE" -e "GITHUB_PAT" -e "TFC_TOKEN" -e "TF_REGISTRY_TOKEN" -e "VCS_USERNAME" -e "VCS_BASE_URL" -e "VCS_TOKEN" -e "BITBUCKET_TOKEN" -e "BITBUCKET_USERNAME" -e "BITBUCKET_APP_PASSWORD" -e "PRISMA_API_URL" -e "CKV_VALIDATE_SECRETS" -e "HOME" -e "GITHUB_JOB" -e "GITHUB_REF" -e "GITHUB_SHA" -e "GITHUB_REPOSITORY" -e "GITHUB_REPOSITORY_OWNER" -e "GITHUB_REPOSITORY_OWNER_ID" -e "GITHUB_RUN_ID" -e "GITHUB_RUN_NUMBER" -e "GITHUB_RETENTION_DAYS" -e "GITHUB_RUN_ATTEMPT" -e "GITHUB_ACTOR_ID" -e "GITHUB_ACTOR" -e "GITHUB_WORKFLOW" -e "GITHUB_HEAD_REF" -e "GITHUB_BASE_REF" -e "GITHUB_EVENT_NAME" -e "GITHUB_SERVER_URL" -e "GITHUB_API_URL" -e "GITHUB_GRAPHQL_URL" -e "GITHUB_REF_NAME" -e "GITHUB_REF_PROTECTED" -e "GITHUB_REF_TYPE" -e "GITHUB_WORKFLOW_REF" -e "GITHUB_WORKFLOW_SHA" -e "GITHUB_REPOSITORY_ID" -e "GITHUB_TRIGGERING_ACTOR" -e "GITHUB_WORKSPACE" -e "GITHUB_ACTION" -e "GITHUB_EVENT_PATH" -e "GITHUB_ACTION_REPOSITORY" -e "GITHUB_ACTION_REF" -e "GITHUB_PATH" -e "GITHUB_ENV" -e "GITHUB_STEP_SUMMARY" -e "GITHUB_STATE" -e "GITHUB_OUTPUT" -e "RUNNER_OS" -e "RUNNER_ARCH" -e "RUNNER_NAME" -e "RUNNER_ENVIRONMENT" -e "RUNNER_TOOL_CACHE" -e "RUNNER_TEMP" -e "RUNNER_WORKSPACE" -e "ACTIONS_RUNTIME_URL" -e "ACTIONS_RUNTIME_TOKEN" -e "ACTIONS_CACHE_URL" -e "ACTIONS_RESULTS_URL" -e GITHUB_ACTIONS=true -e CI=true -v "/var/run/docker.sock":"/var/run/docker.sock" -v "/home/runner/work/_temp/_github_home":"/github/home" -v "/home/runner/work/_temp/_github_workflow":"/github/workflow" -v "/home/runner/work/_temp/_runner_file_commands":"/github/file_commands" -v "/home/runner/work/freightliner/freightliner":"/github/workspace" ghcr.io/bridgecrewio/checkov:3.2.457  "" "." "CKV_DOCKER_*,CKV_K8S_*,CKV_TF_*,CKV_GHA_*" "" "" "" "false" "" "" "" "terraform,kubernetes,dockerfile,github_actions" "" "" "" "sarif" "checkov-results.sarif" "" "" "WARNING" "" "" "" "" "" "" "" "" "" "" "" "" "" "" "--user 0"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4017556Z BC_FROM_BRANCH=master
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4018161Z BC_TO_BRANCH=
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4035282Z BC_PR_ID=master
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4036297Z BC_PR_URL=https://github.com/hemzaz/freightliner/pull/master
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4036924Z BC_COMMIT_HASH=24a2a293709dd2cfbac6fd7fdab281d5e725eb7f
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4040558Z BC_COMMIT_URL=https://github.com/hemzaz/freightliner/commit/24a2a293709dd2cfbac6fd7fdab281d5e725eb7f
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4041196Z BC_AUTHOR_NAME=hemzaz
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4041431Z BC_AUTHOR_URL=https://github.com/hemzaz
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4041684Z BC_RUN_ID=3
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4041964Z BC_RUN_URL=https://github.com/hemzaz/freightliner/actions/runs/16708356037
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4042372Z BC_REPOSITORY_URL=https://github.com/hemzaz/freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4042678Z running checkov on directory: .
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:01.4043454Z checkov -d .  --check CKV_DOCKER_* --check CKV_K8S_* --check CKV_TF_* --check CKV_GHA_*            --output sarif --output-file-path checkov-results.sarif      --framework terraform,kubernetes,dockerfile,github_actions         
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0628644Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0629336Z        _               _
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0629752Z    ___| |__   ___  ___| | _______   __
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0630150Z   / __| '_ \ / _ \/ __| |/ / _ \ \ / /
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0630526Z  | (__| | | |  __/ (__|   < (_) \ V /
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0630937Z   \___|_| |_|\___|\___|_|\_\___/ \_/
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0631240Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0631572Z By Prisma Cloud | version: 3.2.457 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0631979Z kubernetes scan results:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0632215Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0632442Z Passed checks: 96, Failed checks: 8, Skipped checks: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0632789Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0641995Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0642774Z 	PASSED for resource: ConfigMap.freightliner.freightliner-config
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0643452Z 	File: /deployments/kubernetes/configmap.yaml:1-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0644526Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0645635Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0646327Z 	PASSED for resource: ConfigMap.freightliner.freightliner-scripts
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0646981Z 	File: /deployments/kubernetes/configmap.yaml:121-216
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0648191Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0649649Z Check: CKV_K8S_27: "Do not expose the docker daemon socket to containers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0650357Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0650991Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0652059Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-26
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0653278Z Check: CKV_K8S_141: "Ensure that the --read-only-port argument is set to 0"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0654014Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0654631Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0655991Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-read-only-port-argument-is-set-to-0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0658316Z Check: CKV_K8S_31: "Ensure that the seccomp profile is set to docker/default or runtime/default"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0661550Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0662320Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0664151Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-29
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0665988Z Check: CKV_K8S_83: "Ensure that the admission control plugin NamespaceLifecycle is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0667518Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0670024Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0671733Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-namespacelifecycle-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0673643Z Check: CKV_K8S_112: "Ensure that the RotateKubeletServerCertificate argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0674603Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0675336Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0677105Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-rotatekubeletservercertificate-argument-is-set-to-true-for-controller-manager
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0679363Z Check: CKV_K8S_16: "Container should not be privileged"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0680283Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0681013Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0682234Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-15
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0683515Z Check: CKV_K8S_149: "Ensure that the --rotate-certificates argument is not set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0684492Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0685296Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0686963Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-rotate-certificates-argument-is-not-set-to-false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0722753Z Check: CKV_K8S_107: "Ensure that the --profiling argument is set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0723618Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0724254Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0725610Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-profiling-argument-is-set-to-false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0727146Z Check: CKV_K8S_74: "Ensure that the --authorization-mode argument is not set to AlwaysAllow"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0728265Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0729161Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0730722Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-authorization-mode-argument-is-not-set-to-alwaysallow-1
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0732371Z Check: CKV_K8S_26: "Do not specify hostPort unless absolutely necessary"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0733082Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0733686Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0734741Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-25
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0735974Z Check: CKV_K8S_28: "Minimize the admission of containers with the NET_RAW capability"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0736751Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0737369Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0738602Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-27
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0739780Z Check: CKV_K8S_71: "Ensure that the --kubelet-https argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0740501Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0741109Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0742664Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-kubelet-https-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0744204Z Check: CKV_K8S_96: "Ensure that the --service-account-lookup argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0744996Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0745611Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0747052Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-service-account-lookup-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0748830Z Check: CKV_K8S_79: "Ensure that the admission control plugin AlwaysAdmit is not set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0749607Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0750205Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0751666Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-alwaysadmit-is-not-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0753271Z Check: CKV_K8S_138: "Ensure that the --anonymous-auth argument is set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0753997Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0754611Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0755992Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-anonymous-auth-argument-is-set-to-false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0757557Z Check: CKV_K8S_80: "Ensure that the admission control plugin AlwaysPullImages is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0758473Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0759071Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0760528Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-alwayspullimages-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0762248Z Check: CKV_K8S_143: "Ensure that the --streaming-connection-idle-timeout argument is not set to 0"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0763087Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0763685Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0765219Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-streaming-connection-idle-timeout-argument-is-not-set-to-0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0767093Z Check: CKV_K8S_37: "Minimize the admission of containers with capabilities assigned"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0768108Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0768719Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0769786Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-34
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0771012Z Check: CKV_K8S_95: "Ensure that the --request-timeout argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0771786Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0772403Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0773873Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-request-timeout-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0775449Z Check: CKV_K8S_159: "Limit the use of git-sync to prevent code injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0776166Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0776784Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0777451Z Check: CKV_K8S_18: "Containers should not share the host IPC namespace"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0778310Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0779095Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0780272Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-17
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0781527Z Check: CKV_K8S_25: "Minimize the admission of containers with added capability"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0782274Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0782876Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0783931Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-24
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0785136Z Check: CKV_K8S_82: "Ensure that the admission control plugin ServiceAccount is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0785895Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0786506Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0788111Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-serviceaccount-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0789267Z Check: CKV_K8S_40: "Containers should run as a high UID to avoid host conflict"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0789709Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0790067Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0790681Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-37
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0791558Z Check: CKV_K8S_105: "Ensure that the API Server only makes use of Strong Cryptographic Ciphers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0792074Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0792423Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0793439Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-api-server-only-makes-use-of-strong-cryptographic-ciphers
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0795529Z Check: CKV_K8S_34: "Ensure that Tiller (Helm v2) is not deployed"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0796182Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0796791Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0797956Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0799156Z Check: CKV_K8S_84: "Ensure that the admission control plugin PodSecurityPolicy is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0800191Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0800869Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0802379Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-podsecuritypolicy-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0804206Z Check: CKV_K8S_100: "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0805115Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0805725Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0807385Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-tls-cert-file-and-tls-private-key-file-arguments-are-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0809282Z Check: CKV_K8S_23: "Minimize the admission of root containers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0809901Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0810490Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0811315Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-22
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0812567Z Check: CKV_K8S_85: "Ensure that the admission control plugin NodeRestriction is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0813598Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0814191Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0815143Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-noderestriction-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0816201Z Check: CKV_K8S_97: "Ensure that the --service-account-key-file argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0816918Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0817585Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0818916Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-service-account-key-file-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0820201Z Check: CKV_K8S_22: "Use read-only filesystem for containers where possible"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0820902Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0821515Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0822520Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-21
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0823666Z Check: CKV_K8S_90: "Ensure that the --profiling argument is set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0824376Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0824984Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0826391Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-profiling-argument-is-set-to-false-2
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0828127Z Check: CKV_K8S_118: "Ensure that the --auto-tls argument is not set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0828844Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0829447Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0830794Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-auto-tls-argument-is-not-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0832208Z Check: CKV_K8S_89: "Ensure that the --secure-port argument is not set to 0"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0832768Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0833364Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0834304Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-secure-port-argument-is-not-set-to-0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0835557Z Check: CKV_K8S_39: "Do not use the CAP_SYS_ADMIN linux capability"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0836028Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0836367Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0836956Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-36
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0837602Z Check: CKV_K8S_29: "Apply security context to your pods and containers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0838292Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0838639Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0839746Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-securitycontext-is-applied-to-pods-and-containers
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0840659Z Check: CKV_K8S_104: "Ensure that encryption providers are appropriately configured"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0841189Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0841539Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0842323Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-etcd-cafile-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0843353Z Check: CKV_K8S_115: "Ensure that the --bind-address argument is set to 127.0.0.1"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0843757Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0844093Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0844864Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-bind-address-argument-is-set-to-127001-1
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0845647Z Check: CKV_K8S_9: "Readiness Probe Should be Configured"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0845993Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0846319Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0846893Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-8
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0847522Z Check: CKV_K8S_91: "Ensure that the --audit-log-path argument is set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0848179Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0848520Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0849248Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-audit-log-path-argument-is-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0850053Z Check: CKV_K8S_77: "Ensure that the --authorization-mode argument includes RBAC"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0850466Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0850798Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0851571Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-authorization-mode-argument-includes-rbac
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0852417Z Check: CKV_K8S_75: "Ensure that the --authorization-mode argument includes Node"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0852825Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0853165Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0853932Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-authorization-mode-argument-includes-node
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0854757Z Check: CKV_K8S_14: "Image Tag should be fixed - not latest or blank"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0855141Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0855473Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0856199Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-13
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0856943Z Check: CKV_K8S_110: "Ensure that the --service-account-private-key-file argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0857435Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0857991Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0858898Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-service-account-private-key-file-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0860547Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0860943Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0861290Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0861876Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0862678Z Check: CKV_K8S_68: "Ensure that the --anonymous-auth argument is set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0863101Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0863449Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0864392Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-anonymous-auth-argument-is-set-to-false-1
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0865227Z Check: CKV_K8S_70: "Ensure that the --token-auth-file argument is not set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0865623Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0865956Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0866765Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-token-auth-file-parameter-is-not-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0867937Z Check: CKV_K8S_116: "Ensure that the --cert-file and --key-file arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0868411Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0868751Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0869594Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-cert-file-and-key-file-arguments-are-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0870618Z Check: CKV_K8S_148: "Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0871128Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0871458Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0872430Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-tls-cert-file-and-tls-private-key-file-arguments-are-set-as-appropriate-for-kubelet
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0873544Z Check: CKV_K8S_73: "Ensure that the --kubelet-certificate-authority argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0874016Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0874352Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0875212Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-kubelet-certificate-authority-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0876146Z Check: CKV_K8S_88: "Ensure that the --insecure-port argument is set to 0"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0876535Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0876866Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0877616Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-insecure-port-argument-is-set-to-0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0878979Z Check: CKV_K8S_72: "Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0879539Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0879885Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0880866Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-kubelet-client-certificate-and-kubelet-client-key-arguments-are-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0881968Z Check: CKV_K8S_108: "Ensure that the --use-service-account-credentials argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0882428Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0882763Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0883736Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-use-service-account-credentials-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0884836Z Check: CKV_K8S_93: "Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0885305Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0885645Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0886655Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-audit-log-maxbackup-argument-is-set-to-10-or-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0887591Z Check: CKV_K8S_113: "Ensure that the --bind-address argument is set to 127.0.0.1"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0888341Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0888776Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0889559Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-bind-address-argument-is-set-to-127001
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0890463Z Check: CKV_K8S_92: "Ensure that the --audit-log-maxage argument is set to 30 or as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0890916Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0891337Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0893244Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-audit-log-maxage-argument-is-set-to-30-or-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0894335Z Check: CKV_K8S_99: "Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0894948Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0895476Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0896479Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-etcd-certfile-and-etcd-keyfile-arguments-are-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0898362Z Check: CKV_K8S_146: "Ensure that the --hostname-override argument is not set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0898902Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0899540Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0900828Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-hostname-override-argument-is-not-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0902378Z Check: CKV_K8S_139: "Ensure that the --authorization-mode argument is not set to AlwaysAllow"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0902878Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0903315Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0904651Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-authorization-mode-argument-is-not-set-to-alwaysallow
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0906643Z Check: CKV_K8S_94: "Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0907526Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0908170Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0909059Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-audit-log-maxsize-argument-is-set-to-100-or-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0910048Z Check: CKV_K8S_145: "Ensure that the --make-iptables-util-chains argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0910495Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0910835Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0911654Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-make-iptables-util-chains-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0912651Z Check: CKV_K8S_147: "Ensure that the --event-qps argument is set to 0 or a level which ensures appropriate event capture"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0913331Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0913958Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0916026Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-event-qps-argument-is-set-to-0-or-a-level-which-ensures-appropriate-event-capture
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0918294Z Check: CKV_K8S_106: "Ensure that the --terminated-pod-gc-threshold argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0919197Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0919813Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0921388Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-terminated-pod-gc-threshold-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0923119Z Check: CKV_K8S_69: "Ensure that the --basic-auth-file argument is not set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0923848Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0924475Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0925896Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-basic-auth-file-argument-is-not-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0927474Z Check: CKV_K8S_86: "Ensure that the --insecure-bind-address argument is not set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0928469Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0929105Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0930546Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-insecure-bind-address-argument-is-not-set
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0932235Z Check: CKV_K8S_81: "Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0933033Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0933425Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0934607Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-admission-control-plugin-securitycontextdeny-is-set-if-podsecuritypolicy-is-not-used
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0935968Z Check: CKV_K8S_102: "Ensure that the --etcd-cafile argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0936422Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0936874Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0937972Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-etcd-cafile-argument-is-set-as-appropriate-1
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0939289Z Check: CKV_K8S_35: "Prefer using secrets as files over secrets as environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0939749Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0940187Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0940884Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-33
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0941737Z Check: CKV_K8S_119: "Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0942239Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0942687Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0943676Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-peer-cert-file-and-peer-key-file-arguments-are-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0944815Z Check: CKV_K8S_17: "Containers should not share the host process ID namespace"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0945265Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0945697Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0946588Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-16
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0947390Z Check: CKV_K8S_151: "Ensure that the Kubelet only makes use of Strong Cryptographic Ciphers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0948089Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0948429Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0949404Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-kubelet-only-makes-use-of-strong-cryptographic-ciphers
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0950455Z Check: CKV_K8S_30: "Apply security context to your containers"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0950909Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0951389Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0952104Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-28
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0952907Z Check: CKV_K8S_111: "Ensure that the --root-ca-file argument is set as appropriate"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0953349Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0953805Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0954698Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-root-ca-file-argument-is-set-as-appropriate
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0955938Z Check: CKV_K8S_117: "Ensure that the --client-cert-auth argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0956691Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0957278Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0958652Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-client-cert-auth-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0959690Z Check: CKV_K8S_144: "Ensure that the --protect-kernel-defaults argument is set to true"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0960250Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0960876Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0962177Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-protect-kernel-defaults-argument-is-set-to-true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0963239Z Check: CKV_K8S_20: "Containers should not run with allowPrivilegeEscalation"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0963922Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0964373Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0965149Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-19
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0966350Z Check: CKV_K8S_33: "Ensure the Kubernetes dashboard is not deployed"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0967105Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0967882Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0968761Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-31
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0969773Z Check: CKV_K8S_8: "Liveness Probe Should be Configured"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0970134Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0970592Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0971192Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-7
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0971947Z Check: CKV_K8S_19: "Containers should not share the host network namespace"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0972361Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0972699Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0973465Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-18
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0974226Z Check: CKV_K8S_114: "Ensure that the --profiling argument is set to false"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0974634Z 	PASSED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0974980Z 	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0975745Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/ensure-that-the-profiling-argument-is-set-to-false-1
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0976545Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0976929Z 	PASSED for resource: Secret.freightliner.freightliner-api-keys
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0977289Z 	File: /deployments/kubernetes/secrets.yaml:6-18
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0978158Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0978796Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0979232Z 	PASSED for resource: Secret.freightliner.freightliner-registry-credentials
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0979636Z 	File: /deployments/kubernetes/secrets.yaml:19-34
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0980216Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0980830Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0981236Z 	PASSED for resource: Secret.freightliner.freightliner-encryption-keys
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0981623Z 	File: /deployments/kubernetes/secrets.yaml:35-49
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0982205Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0982808Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0983174Z 	PASSED for resource: Secret.freightliner.freightliner-tls
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0983508Z 	File: /deployments/kubernetes/secrets.yaml:50-68
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0984082Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0984693Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0985196Z 	PASSED for resource: Secret.freightliner.freightliner-monitoring
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0985552Z 	File: /deployments/kubernetes/secrets.yaml:69-81
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0986120Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0987044Z Check: CKV_K8S_152: "Prevent NGINX Ingress annotation snippets which contain LUA code execution. See CVE-2021-25742"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0987595Z 	PASSED for resource: Ingress.freightliner.freightliner-ingress
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0988329Z 	File: /deployments/kubernetes/ingress.yaml:1-98
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0989411Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/prevent-nginx-ingress-annotation-snippets-which-contain-lua-code-execution
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0991151Z Check: CKV_K8S_154: "Prevent NGINX Ingress annotation snippets which contain alias statements See CVE-2021-25742"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0991851Z 	PASSED for resource: Ingress.freightliner.freightliner-ingress
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0992276Z 	File: /deployments/kubernetes/ingress.yaml:1-98
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0993358Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/prevent-nginx-ingress-annotation-snippets-which-contain-alias-statements
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0994370Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0994860Z 	PASSED for resource: Ingress.freightliner.freightliner-ingress
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0995213Z 	File: /deployments/kubernetes/ingress.yaml:1-98
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0996140Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0996896Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0997356Z 	PASSED for resource: Service.freightliner.freightliner-service
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0997916Z 	File: /deployments/kubernetes/service.yaml:1-41
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0998764Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0999508Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.0999927Z 	PASSED for resource: Service.freightliner.freightliner-internal
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1000396Z 	File: /deployments/kubernetes/service.yaml:42-69
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1001109Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1001791Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1002222Z 	PASSED for resource: Service.freightliner.freightliner-headless
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1002661Z 	File: /deployments/kubernetes/service.yaml:70-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1003440Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1004552Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1005251Z 	PASSED for resource: ConfigMap.freightliner.registry-security-config
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1005975Z 	File: /security/registry-security-config.yaml:4-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1007083Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1008231Z Check: CKV_K8S_21: "The default namespace should not be used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1008664Z 	PASSED for resource: ConfigMap.freightliner.freightliner-security-policy
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1009083Z 	File: /security/container-runtime-policy.yaml:3-455
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1009679Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1010253Z Check: CKV_K8S_11: "CPU limits should be set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1012178Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1040944Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1047251Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-10
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1048234Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1048486Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1048949Z Check: CKV_K8S_10: "CPU requests should be set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1049311Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1050028Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1051111Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-9
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1051595Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1051825Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1052261Z Check: CKV_K8S_13: "Memory limits should be set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1052613Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1053183Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1054160Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-12
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1054641Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1054882Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1055323Z Check: CKV_K8S_12: "Memory requests should be set"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1055677Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1056404Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1057406Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-11
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1058208Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1058445Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1058880Z Check: CKV_K8S_43: "Image should use digest"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1059210Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1059791Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1060820Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-39
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1061298Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1061524Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1062060Z Check: CKV_K8S_38: "Ensure that Service Account Tokens are only mounted where necessary"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1062521Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1063162Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1064199Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-35
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1064681Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1064907Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1065357Z Check: CKV_K8S_15: "Image Pull Policy should be Always"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1065711Z 	FAILED for resource: Deployment.freightliner.freightliner
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1066255Z ##[error]	File: /deployments/kubernetes/deployment.yaml:1-279
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1067229Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/bc-k8s-14
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1067976Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1068365Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1068992Z Check: CKV_K8S_153: "Prevent All NGINX Ingress annotation snippets. See CVE-2021-25742"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1069466Z 	FAILED for resource: Ingress.freightliner.freightliner-ingress
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1070168Z ##[error]	File: /deployments/kubernetes/ingress.yaml:1-98
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1071358Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/kubernetes-policies/kubernetes-policy-index/prevent-all-nginx-ingress-annotation-snippets
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1072160Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1072390Z 		Code lines for this resource are too many. Please use IDE of your choice to review the file.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1072701Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1072788Z dockerfile scan results:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1072918Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1073046Z Passed checks: 8, Failed checks: 0, Skipped checks: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1073261Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1073461Z Check: CKV_DOCKER_11: "Ensure From Alias are unique for multistage builds."
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1073820Z 	PASSED for resource: /Dockerfile.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1074054Z 	File: /Dockerfile:1-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1074737Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-docker-from-alias-is-unique-for-multistage-builds
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1075546Z Check: CKV_DOCKER_7: "Ensure the base image uses a non latest version tag"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1075897Z 	PASSED for resource: /Dockerfile.FROM
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1076148Z 	File: /Dockerfile:43-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1076969Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-the-base-image-uses-a-non-latest-version-tag
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1077913Z Check: CKV_DOCKER_8: "Ensure the last USER is not root"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1078219Z 	PASSED for resource: /Dockerfile.USER
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1078465Z 	File: /Dockerfile:51-51
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1079179Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-the-last-user-is-not-root
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1079882Z Check: CKV_DOCKER_10: "Ensure that WORKDIR values are absolute paths"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1080220Z 	PASSED for resource: /Dockerfile.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1080446Z 	File: /Dockerfile:1-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1081076Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-docker-workdir-values-are-absolute-paths
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1081907Z Check: CKV_DOCKER_5: "Ensure update instructions are not use alone in the Dockerfile"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1082288Z 	PASSED for resource: /Dockerfile.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1082517Z 	File: /Dockerfile:1-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1083212Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-update-instructions-are-not-used-alone-in-the-dockerfile
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1083993Z Check: CKV_DOCKER_9: "Ensure that APT isn't used"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1084272Z 	PASSED for resource: /Dockerfile.
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1084491Z 	File: /Dockerfile:1-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1085046Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-docker-apt-is-not-used
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1085745Z Check: CKV_DOCKER_3: "Ensure that a user for the container has been created"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1086101Z 	PASSED for resource: /Dockerfile.USER
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1086345Z 	File: /Dockerfile:51-51
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1087006Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-that-a-user-for-the-container-has-been-created
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1088091Z Check: CKV_DOCKER_2: "Ensure that HEALTHCHECK instructions have been added to container images"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1088532Z 	PASSED for resource: /Dockerfile.HEALTHCHECK
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1088806Z 	File: /Dockerfile:57-58
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1089559Z 	Guide: https://docs.prismacloud.io/en/enterprise-edition/policy-reference/docker-policies/docker-policy-index/ensure-that-healthcheck-instructions-have-been-added-to-container-images
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090269Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090356Z github_actions scan results:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090506Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090638Z Passed checks: 1209, Failed checks: 4, Skipped checks: 0
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090855Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1090996Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1091313Z 	PASSED for resource: jobs(security-posture)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1091765Z 	File: /.github/workflows/security-monitoring.yml:45-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1092106Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1092436Z 	PASSED for resource: jobs(vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1092769Z 	File: /.github/workflows/security-monitoring.yml:195-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1093107Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1093431Z 	PASSED for resource: jobs(container-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1093749Z 	File: /.github/workflows/security-monitoring.yml:295-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1094080Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1094379Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1094689Z 	File: /.github/workflows/security-monitoring.yml:377-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1095135Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1095535Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1095806Z 	File: /.github/workflows/security-monitoring.yml:44-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1096244Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1096668Z 	PASSED for resource: jobs(security-posture)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1096975Z 	File: /.github/workflows/security-monitoring.yml:45-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1097532Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1098257Z 	PASSED for resource: jobs(vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1098606Z 	File: /.github/workflows/security-monitoring.yml:195-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1099065Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1099498Z 	PASSED for resource: jobs(container-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1099817Z 	File: /.github/workflows/security-monitoring.yml:295-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1100260Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1100717Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1101028Z 	File: /.github/workflows/security-monitoring.yml:377-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1101376Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1101690Z 	PASSED for resource: jobs(security-posture)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1102005Z 	File: /.github/workflows/security-monitoring.yml:45-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1102342Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1102682Z 	PASSED for resource: jobs(vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1103007Z 	File: /.github/workflows/security-monitoring.yml:195-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1103348Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1103676Z 	PASSED for resource: jobs(container-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1103996Z 	File: /.github/workflows/security-monitoring.yml:295-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1104345Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1104659Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1104966Z 	File: /.github/workflows/security-monitoring.yml:377-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1105356Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1105719Z 	PASSED for resource: jobs(security-posture)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1106028Z 	File: /.github/workflows/security-monitoring.yml:45-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1106411Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1106798Z 	PASSED for resource: jobs(vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1107128Z 	File: /.github/workflows/security-monitoring.yml:195-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1107515Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1108001Z 	PASSED for resource: jobs(container-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1108332Z 	File: /.github/workflows/security-monitoring.yml:295-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1108871Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1109242Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1109939Z 	File: /.github/workflows/security-monitoring.yml:377-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1110405Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1110821Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1111089Z 	File: /.github/workflows/security-monitoring.yml:44-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1111429Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1111795Z 	PASSED for resource: jobs(security-posture).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1112169Z 	File: /.github/workflows/security-monitoring.yml:55-61
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1112491Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1112856Z 	PASSED for resource: jobs(security-posture).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1113232Z 	File: /.github/workflows/security-monitoring.yml:60-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1113553Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1113969Z 	PASSED for resource: jobs(security-posture).steps[3](Security posture assessment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1114414Z 	File: /.github/workflows/security-monitoring.yml:66-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1114759Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1115278Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1115684Z 	File: /.github/workflows/security-monitoring.yml:202-208
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1116024Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1116411Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1116822Z 	File: /.github/workflows/security-monitoring.yml:207-213
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1117160Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1117535Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1118043Z 	File: /.github/workflows/security-monitoring.yml:212-218
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1118379Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1118854Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[4](Install vulnerability scanning tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1119346Z 	File: /.github/workflows/security-monitoring.yml:217-224
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1119686Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1120159Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[5](Run comprehensive vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1120647Z 	File: /.github/workflows/security-monitoring.yml:223-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1120985Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1121436Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[6](Upload vulnerability reports)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1121893Z 	File: /.github/workflows/security-monitoring.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1122240Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1122616Z 	PASSED for resource: jobs(container-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1123008Z 	File: /.github/workflows/security-monitoring.yml:302-308
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1123349Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1123729Z 	PASSED for resource: jobs(container-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1124127Z 	File: /.github/workflows/security-monitoring.yml:307-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1124461Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1124869Z 	PASSED for resource: jobs(container-monitoring).steps[3](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1125281Z 	File: /.github/workflows/security-monitoring.yml:312-316
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1125623Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1126056Z 	PASSED for resource: jobs(container-monitoring).steps[4](Build container for monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1126623Z 	File: /.github/workflows/security-monitoring.yml:315-329
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1126963Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1127374Z 	PASSED for resource: jobs(container-monitoring).steps[5](Run Trivy container scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1128028Z 	File: /.github/workflows/security-monitoring.yml:328-339
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1128390Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1128818Z 	PASSED for resource: jobs(container-monitoring).steps[6](Process container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1129279Z 	File: /.github/workflows/security-monitoring.yml:338-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1129613Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1130050Z 	PASSED for resource: jobs(container-monitoring).steps[7](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1130498Z 	File: /.github/workflows/security-monitoring.yml:367-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1130829Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1131208Z 	PASSED for resource: jobs(security-alerting).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1131589Z 	File: /.github/workflows/security-monitoring.yml:384-390
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1131944Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1132357Z 	PASSED for resource: jobs(security-alerting).steps[2](Evaluate alert conditions)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1132904Z 	File: /.github/workflows/security-monitoring.yml:389-437
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1133242Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1133629Z 	PASSED for resource: jobs(security-alerting).steps[3](Send security alerts)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1134030Z 	File: /.github/workflows/security-monitoring.yml:436-528
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1134365Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1134751Z 	PASSED for resource: jobs(security-alerting).steps[4](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1135162Z 	File: /.github/workflows/security-monitoring.yml:527-569
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1135499Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1135913Z 	PASSED for resource: jobs(security-alerting).steps[5](Generate monitoring summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1136343Z 	File: /.github/workflows/security-monitoring.yml:568-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1136790Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1137284Z 	PASSED for resource: jobs(security-posture).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1137829Z 	File: /.github/workflows/security-monitoring.yml:55-61
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1138348Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1138839Z 	PASSED for resource: jobs(security-posture).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1139213Z 	File: /.github/workflows/security-monitoring.yml:60-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1139655Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1140186Z 	PASSED for resource: jobs(security-posture).steps[3](Security posture assessment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1140615Z 	File: /.github/workflows/security-monitoring.yml:66-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1141059Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1141563Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1141975Z 	File: /.github/workflows/security-monitoring.yml:202-208
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1142417Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1142925Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1143330Z 	File: /.github/workflows/security-monitoring.yml:207-213
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1143766Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1144405Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1144801Z 	File: /.github/workflows/security-monitoring.yml:212-218
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1145243Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1145833Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[4](Install vulnerability scanning tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1146324Z 	File: /.github/workflows/security-monitoring.yml:217-224
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1146770Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1147783Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[5](Run comprehensive vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1148727Z 	File: /.github/workflows/security-monitoring.yml:223-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1149531Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1150341Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[6](Upload vulnerability reports)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1151166Z 	File: /.github/workflows/security-monitoring.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1152110Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1152909Z 	PASSED for resource: jobs(container-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1153606Z 	File: /.github/workflows/security-monitoring.yml:302-308
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1154615Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1155695Z 	PASSED for resource: jobs(container-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1156370Z 	File: /.github/workflows/security-monitoring.yml:307-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1156860Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1157432Z 	PASSED for resource: jobs(container-monitoring).steps[3](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1158383Z 	File: /.github/workflows/security-monitoring.yml:312-316
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1158984Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1159543Z 	PASSED for resource: jobs(container-monitoring).steps[4](Build container for monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1159994Z 	File: /.github/workflows/security-monitoring.yml:315-329
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1160460Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1160990Z 	PASSED for resource: jobs(container-monitoring).steps[5](Run Trivy container scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1161413Z 	File: /.github/workflows/security-monitoring.yml:328-339
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1161868Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1162413Z 	PASSED for resource: jobs(container-monitoring).steps[6](Process container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1162862Z 	File: /.github/workflows/security-monitoring.yml:338-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1163321Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1163866Z 	PASSED for resource: jobs(container-monitoring).steps[7](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1164310Z 	File: /.github/workflows/security-monitoring.yml:367-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1164759Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1165247Z 	PASSED for resource: jobs(security-alerting).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1165637Z 	File: /.github/workflows/security-monitoring.yml:384-390
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1166078Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1166603Z 	PASSED for resource: jobs(security-alerting).steps[2](Evaluate alert conditions)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1167023Z 	File: /.github/workflows/security-monitoring.yml:389-437
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1167826Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1168371Z 	PASSED for resource: jobs(security-alerting).steps[3](Send security alerts)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1168770Z 	File: /.github/workflows/security-monitoring.yml:436-528
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1169221Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1169736Z 	PASSED for resource: jobs(security-alerting).steps[4](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1170135Z 	File: /.github/workflows/security-monitoring.yml:527-569
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1170584Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1171106Z 	PASSED for resource: jobs(security-alerting).steps[5](Generate monitoring summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1171535Z 	File: /.github/workflows/security-monitoring.yml:568-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1171900Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1172310Z 	PASSED for resource: jobs(security-posture).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1172700Z 	File: /.github/workflows/security-monitoring.yml:55-61
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1173048Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1173437Z 	PASSED for resource: jobs(security-posture).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1173816Z 	File: /.github/workflows/security-monitoring.yml:60-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1174295Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1174727Z 	PASSED for resource: jobs(security-posture).steps[3](Security posture assessment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1175148Z 	File: /.github/workflows/security-monitoring.yml:66-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1175499Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1175910Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1176316Z 	File: /.github/workflows/security-monitoring.yml:202-208
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1176679Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1177081Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1177491Z 	File: /.github/workflows/security-monitoring.yml:207-213
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1178092Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1178493Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1178900Z 	File: /.github/workflows/security-monitoring.yml:212-218
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1179248Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1179746Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[4](Install vulnerability scanning tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1180248Z 	File: /.github/workflows/security-monitoring.yml:217-224
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1180598Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1181082Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[5](Run comprehensive vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1181571Z 	File: /.github/workflows/security-monitoring.yml:223-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1181921Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1182381Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[6](Upload vulnerability reports)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1182835Z 	File: /.github/workflows/security-monitoring.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1183185Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1183571Z 	PASSED for resource: jobs(container-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1183961Z 	File: /.github/workflows/security-monitoring.yml:302-308
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1184304Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1184692Z 	PASSED for resource: jobs(container-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1185085Z 	File: /.github/workflows/security-monitoring.yml:307-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1185560Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1185973Z 	PASSED for resource: jobs(container-monitoring).steps[3](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1186386Z 	File: /.github/workflows/security-monitoring.yml:312-316
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1186730Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1187174Z 	PASSED for resource: jobs(container-monitoring).steps[4](Build container for monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1187770Z 	File: /.github/workflows/security-monitoring.yml:315-329
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1188183Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1188612Z 	PASSED for resource: jobs(container-monitoring).steps[5](Run Trivy container scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1189035Z 	File: /.github/workflows/security-monitoring.yml:328-339
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1189387Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1189829Z 	PASSED for resource: jobs(container-monitoring).steps[6](Process container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1190280Z 	File: /.github/workflows/security-monitoring.yml:338-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1190624Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1191063Z 	PASSED for resource: jobs(container-monitoring).steps[7](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1191507Z 	File: /.github/workflows/security-monitoring.yml:367-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1192022Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1192409Z 	PASSED for resource: jobs(security-alerting).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1192795Z 	File: /.github/workflows/security-monitoring.yml:384-390
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1193137Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1193558Z 	PASSED for resource: jobs(security-alerting).steps[2](Evaluate alert conditions)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1193976Z 	File: /.github/workflows/security-monitoring.yml:389-437
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1194326Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1194736Z 	PASSED for resource: jobs(security-alerting).steps[3](Send security alerts)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1195131Z 	File: /.github/workflows/security-monitoring.yml:436-528
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1195476Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1195878Z 	PASSED for resource: jobs(security-alerting).steps[4](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1196295Z 	File: /.github/workflows/security-monitoring.yml:527-569
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1196646Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1197066Z 	PASSED for resource: jobs(security-alerting).steps[5](Generate monitoring summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1197500Z 	File: /.github/workflows/security-monitoring.yml:568-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1198160Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1198769Z 	PASSED for resource: jobs(security-posture).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1199546Z 	File: /.github/workflows/security-monitoring.yml:55-61
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1200362Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1201212Z 	PASSED for resource: jobs(security-posture).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1201888Z 	File: /.github/workflows/security-monitoring.yml:60-67
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1202591Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1203432Z 	PASSED for resource: jobs(security-posture).steps[3](Security posture assessment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1204177Z 	File: /.github/workflows/security-monitoring.yml:66-195
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1204874Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1205654Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1206355Z 	File: /.github/workflows/security-monitoring.yml:202-208
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1207032Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1208208Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1208936Z 	File: /.github/workflows/security-monitoring.yml:207-213
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1209630Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1210394Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1211081Z 	File: /.github/workflows/security-monitoring.yml:212-218
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1211748Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1212679Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[4](Install vulnerability scanning tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1213492Z 	File: /.github/workflows/security-monitoring.yml:217-224
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1214131Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1215053Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[5](Run comprehensive vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1215922Z 	File: /.github/workflows/security-monitoring.yml:223-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1216741Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1217810Z 	PASSED for resource: jobs(vulnerability-monitoring).steps[6](Upload vulnerability reports)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1218862Z 	File: /.github/workflows/security-monitoring.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1219581Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1220354Z 	PASSED for resource: jobs(container-monitoring).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1221043Z 	File: /.github/workflows/security-monitoring.yml:302-308
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1221719Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1222442Z 	PASSED for resource: jobs(container-monitoring).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1223062Z 	File: /.github/workflows/security-monitoring.yml:307-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1223696Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1224480Z 	PASSED for resource: jobs(container-monitoring).steps[3](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1225194Z 	File: /.github/workflows/security-monitoring.yml:312-316
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1225885Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1226754Z 	PASSED for resource: jobs(container-monitoring).steps[4](Build container for monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1227524Z 	File: /.github/workflows/security-monitoring.yml:315-329
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1228502Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1229305Z 	PASSED for resource: jobs(container-monitoring).steps[5](Run Trivy container scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1230016Z 	File: /.github/workflows/security-monitoring.yml:328-339
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1230678Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1231553Z 	PASSED for resource: jobs(container-monitoring).steps[6](Process container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1232359Z 	File: /.github/workflows/security-monitoring.yml:338-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1233075Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1233994Z 	PASSED for resource: jobs(container-monitoring).steps[7](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1234659Z 	File: /.github/workflows/security-monitoring.yml:367-377
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1235242Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1236014Z 	PASSED for resource: jobs(security-alerting).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1236797Z 	File: /.github/workflows/security-monitoring.yml:384-390
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1237610Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1238625Z 	PASSED for resource: jobs(security-alerting).steps[2](Evaluate alert conditions)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1239619Z 	File: /.github/workflows/security-monitoring.yml:389-437
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1240346Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1241143Z 	PASSED for resource: jobs(security-alerting).steps[3](Send security alerts)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1241850Z 	File: /.github/workflows/security-monitoring.yml:436-528
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1242539Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1243313Z 	PASSED for resource: jobs(security-alerting).steps[4](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1244000Z 	File: /.github/workflows/security-monitoring.yml:527-569
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1244675Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1245521Z 	PASSED for resource: jobs(security-alerting).steps[5](Generate monitoring summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1246278Z 	File: /.github/workflows/security-monitoring.yml:568-601
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1246897Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1247856Z 	PASSED for resource: jobs(comprehensive-testing)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1248506Z 	File: /.github/workflows/scheduled-comprehensive.yml:37-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1249352Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1250314Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1250806Z 	File: /.github/workflows/scheduled-comprehensive.yml:36-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1251655Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1252437Z 	PASSED for resource: jobs(comprehensive-testing)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1253022Z 	File: /.github/workflows/scheduled-comprehensive.yml:37-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1253688Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1254383Z 	PASSED for resource: jobs(comprehensive-testing)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1255061Z 	File: /.github/workflows/scheduled-comprehensive.yml:37-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1255857Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1256536Z 	PASSED for resource: jobs(comprehensive-testing)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1257140Z 	File: /.github/workflows/scheduled-comprehensive.yml:37-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1258153Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1258933Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1259507Z 	File: /.github/workflows/scheduled-comprehensive.yml:36-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1260144Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1260828Z 	PASSED for resource: jobs(comprehensive-testing).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1261540Z 	File: /.github/workflows/scheduled-comprehensive.yml:53-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1262152Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1262886Z 	PASSED for resource: jobs(comprehensive-testing).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1263691Z 	File: /.github/workflows/scheduled-comprehensive.yml:56-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1264332Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1265110Z 	PASSED for resource: jobs(comprehensive-testing).steps[3](Run comprehensive test suite)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1265934Z 	File: /.github/workflows/scheduled-comprehensive.yml:62-75
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1266566Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1267348Z 	PASSED for resource: jobs(comprehensive-testing).steps[4](Run external dependency tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1268368Z 	File: /.github/workflows/scheduled-comprehensive.yml:74-96
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1268986Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1269698Z 	PASSED for resource: jobs(comprehensive-testing).steps[5](Flaky test detection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1284750Z 	File: /.github/workflows/scheduled-comprehensive.yml:95-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1285571Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1286058Z 	PASSED for resource: jobs(comprehensive-testing).steps[6](Upload comprehensive test results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1286553Z 	File: /.github/workflows/scheduled-comprehensive.yml:126-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1286912Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1287340Z 	PASSED for resource: jobs(comprehensive-testing).steps[7](Generate summary report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1287997Z 	File: /.github/workflows/scheduled-comprehensive.yml:136-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1288465Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1288963Z 	PASSED for resource: jobs(comprehensive-testing).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1289382Z 	File: /.github/workflows/scheduled-comprehensive.yml:53-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1290213Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1291167Z 	PASSED for resource: jobs(comprehensive-testing).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1291903Z 	File: /.github/workflows/scheduled-comprehensive.yml:56-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1292725Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1293448Z 	PASSED for resource: jobs(comprehensive-testing).steps[3](Run comprehensive test suite)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1294072Z 	File: /.github/workflows/scheduled-comprehensive.yml:62-75
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1294522Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1295069Z 	PASSED for resource: jobs(comprehensive-testing).steps[4](Run external dependency tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1295516Z 	File: /.github/workflows/scheduled-comprehensive.yml:74-96
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1295964Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1296479Z 	PASSED for resource: jobs(comprehensive-testing).steps[5](Flaky test detection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1296901Z 	File: /.github/workflows/scheduled-comprehensive.yml:95-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1297359Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1298107Z 	PASSED for resource: jobs(comprehensive-testing).steps[6](Upload comprehensive test results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1298595Z 	File: /.github/workflows/scheduled-comprehensive.yml:126-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1299054Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1299572Z 	PASSED for resource: jobs(comprehensive-testing).steps[7](Generate summary report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1300009Z 	File: /.github/workflows/scheduled-comprehensive.yml:136-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1300363Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1301032Z 	PASSED for resource: jobs(comprehensive-testing).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1301458Z 	File: /.github/workflows/scheduled-comprehensive.yml:53-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1301809Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1302220Z 	PASSED for resource: jobs(comprehensive-testing).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1302631Z 	File: /.github/workflows/scheduled-comprehensive.yml:56-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1302998Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1303435Z 	PASSED for resource: jobs(comprehensive-testing).steps[3](Run comprehensive test suite)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1303872Z 	File: /.github/workflows/scheduled-comprehensive.yml:62-75
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1304218Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1304648Z 	PASSED for resource: jobs(comprehensive-testing).steps[4](Run external dependency tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1305093Z 	File: /.github/workflows/scheduled-comprehensive.yml:74-96
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1305593Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1306000Z 	PASSED for resource: jobs(comprehensive-testing).steps[5](Flaky test detection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1306418Z 	File: /.github/workflows/scheduled-comprehensive.yml:95-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1306764Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1307255Z 	PASSED for resource: jobs(comprehensive-testing).steps[6](Upload comprehensive test results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1307913Z 	File: /.github/workflows/scheduled-comprehensive.yml:126-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1308283Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1308731Z 	PASSED for resource: jobs(comprehensive-testing).steps[7](Generate summary report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1309169Z 	File: /.github/workflows/scheduled-comprehensive.yml:136-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1309580Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1310025Z 	PASSED for resource: jobs(comprehensive-testing).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1310426Z 	File: /.github/workflows/scheduled-comprehensive.yml:53-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1310826Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1311285Z 	PASSED for resource: jobs(comprehensive-testing).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1311698Z 	File: /.github/workflows/scheduled-comprehensive.yml:56-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1312225Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1312705Z 	PASSED for resource: jobs(comprehensive-testing).steps[3](Run comprehensive test suite)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1313155Z 	File: /.github/workflows/scheduled-comprehensive.yml:62-75
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1313549Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1314030Z 	PASSED for resource: jobs(comprehensive-testing).steps[4](Run external dependency tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1314473Z 	File: /.github/workflows/scheduled-comprehensive.yml:74-96
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1314864Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1315327Z 	PASSED for resource: jobs(comprehensive-testing).steps[5](Flaky test detection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1315747Z 	File: /.github/workflows/scheduled-comprehensive.yml:95-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1316136Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1316754Z 	PASSED for resource: jobs(comprehensive-testing).steps[6](Upload comprehensive test results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1317782Z 	File: /.github/workflows/scheduled-comprehensive.yml:126-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1318601Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1319488Z 	PASSED for resource: jobs(comprehensive-testing).steps[7](Generate summary report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1320304Z 	File: /.github/workflows/scheduled-comprehensive.yml:136-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1321727Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1322998Z 	PASSED for resource: on(Security Gates & Policy Enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1323559Z 	File: /.github/workflows/security-gates.yml:4-12
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1324163Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1324508Z 	PASSED for resource: jobs(policy-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1324812Z 	File: /.github/workflows/security-gates.yml:28-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1325316Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1325869Z 	PASSED for resource: jobs(pre-commit-security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1326435Z 	File: /.github/workflows/security-gates.yml:154-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1327015Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1327839Z 	PASSED for resource: jobs(branch-protection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1328637Z 	File: /.github/workflows/security-gates.yml:214-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1329248Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1329831Z 	PASSED for resource: jobs(security-gate-enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1330399Z 	File: /.github/workflows/security-gates.yml:333-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1331204Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1331803Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1332065Z 	File: /.github/workflows/security-gates.yml:27-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1332610Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1333035Z 	PASSED for resource: jobs(policy-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1333348Z 	File: /.github/workflows/security-gates.yml:28-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1333771Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1334201Z 	PASSED for resource: jobs(pre-commit-security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1334511Z 	File: /.github/workflows/security-gates.yml:154-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1334932Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1335349Z 	PASSED for resource: jobs(branch-protection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1335633Z 	File: /.github/workflows/security-gates.yml:214-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1336221Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1336663Z 	PASSED for resource: jobs(security-gate-enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1336969Z 	File: /.github/workflows/security-gates.yml:333-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1337300Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1337613Z 	PASSED for resource: jobs(policy-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1338287Z 	File: /.github/workflows/security-gates.yml:28-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1338642Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1338968Z 	PASSED for resource: jobs(pre-commit-security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1339264Z 	File: /.github/workflows/security-gates.yml:154-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1339624Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1340207Z 	PASSED for resource: jobs(branch-protection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1340714Z 	File: /.github/workflows/security-gates.yml:214-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1341317Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1341898Z 	PASSED for resource: jobs(security-gate-enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1342440Z 	File: /.github/workflows/security-gates.yml:333-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1343082Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1343720Z 	PASSED for resource: jobs(policy-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1344215Z 	File: /.github/workflows/security-gates.yml:28-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1344853Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1345509Z 	PASSED for resource: jobs(pre-commit-security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1346020Z 	File: /.github/workflows/security-gates.yml:154-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1346655Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1347284Z 	PASSED for resource: jobs(branch-protection)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1347964Z 	File: /.github/workflows/security-gates.yml:214-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1348620Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1349287Z 	PASSED for resource: jobs(security-gate-enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1349829Z 	File: /.github/workflows/security-gates.yml:333-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1350586Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1351316Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1351733Z 	File: /.github/workflows/security-gates.yml:27-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1352286Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1353135Z 	PASSED for resource: jobs(policy-validation).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1353753Z 	File: /.github/workflows/security-gates.yml:36-42
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1354301Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1354934Z 	PASSED for resource: jobs(policy-validation).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1355573Z 	File: /.github/workflows/security-gates.yml:41-48
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1356132Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1356883Z 	PASSED for resource: jobs(policy-validation).steps[3](Validate security policy compliance)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1357798Z 	File: /.github/workflows/security-gates.yml:47-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1358349Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1358995Z 	PASSED for resource: jobs(pre-commit-security).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1359630Z 	File: /.github/workflows/security-gates.yml:160-166
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1360170Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1360812Z 	PASSED for resource: jobs(pre-commit-security).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1361442Z 	File: /.github/workflows/security-gates.yml:165-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1361985Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1362608Z 	PASSED for resource: jobs(pre-commit-security).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1363404Z 	File: /.github/workflows/security-gates.yml:171-177
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1363949Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1364664Z 	PASSED for resource: jobs(pre-commit-security).steps[4](Run pre-commit security hooks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1365400Z 	File: /.github/workflows/security-gates.yml:176-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1365980Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1366626Z 	PASSED for resource: jobs(branch-protection).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1367302Z 	File: /.github/workflows/security-gates.yml:220-226
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1367981Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1368407Z 	PASSED for resource: jobs(branch-protection).steps[2](Check branch protection rules)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1368821Z 	File: /.github/workflows/security-gates.yml:225-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1369159Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1374076Z 	PASSED for resource: jobs(security-gate-enforcement).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1374917Z 	File: /.github/workflows/security-gates.yml:340-346
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1375547Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1376324Z 	PASSED for resource: jobs(security-gate-enforcement).steps[2](Evaluate security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1377097Z 	File: /.github/workflows/security-gates.yml:345-387
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1377863Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1378623Z 	PASSED for resource: jobs(security-gate-enforcement).steps[3](Update commit status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1379557Z 	File: /.github/workflows/security-gates.yml:386-405
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1380185Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1381023Z 	PASSED for resource: jobs(security-gate-enforcement).steps[4](Add PR comment with security status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1381880Z 	File: /.github/workflows/security-gates.yml:404-448
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1382473Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1383292Z 	PASSED for resource: jobs(security-gate-enforcement).steps[5](Generate security gates summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1384112Z 	File: /.github/workflows/security-gates.yml:447-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1384884Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1385754Z 	PASSED for resource: jobs(policy-validation).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1386395Z 	File: /.github/workflows/security-gates.yml:36-42
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1387411Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1388531Z 	PASSED for resource: jobs(policy-validation).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1389167Z 	File: /.github/workflows/security-gates.yml:41-48
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1389915Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1390892Z 	PASSED for resource: jobs(policy-validation).steps[3](Validate security policy compliance)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1391658Z 	File: /.github/workflows/security-gates.yml:47-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1392423Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1393289Z 	PASSED for resource: jobs(pre-commit-security).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1393947Z 	File: /.github/workflows/security-gates.yml:160-166
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1394683Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1395564Z 	PASSED for resource: jobs(pre-commit-security).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1396210Z 	File: /.github/workflows/security-gates.yml:165-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1396938Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1397927Z 	PASSED for resource: jobs(pre-commit-security).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1398802Z 	File: /.github/workflows/security-gates.yml:171-177
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1399572Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1400554Z 	PASSED for resource: jobs(pre-commit-security).steps[4](Run pre-commit security hooks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1401394Z 	File: /.github/workflows/security-gates.yml:176-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1402174Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1403045Z 	PASSED for resource: jobs(branch-protection).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1403686Z 	File: /.github/workflows/security-gates.yml:220-226
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1404457Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1405399Z 	PASSED for resource: jobs(branch-protection).steps[2](Check branch protection rules)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1406148Z 	File: /.github/workflows/security-gates.yml:225-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1406927Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1410903Z 	PASSED for resource: jobs(security-gate-enforcement).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1411688Z 	File: /.github/workflows/security-gates.yml:340-346
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1412513Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1413496Z 	PASSED for resource: jobs(security-gate-enforcement).steps[2](Evaluate security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1414265Z 	File: /.github/workflows/security-gates.yml:345-387
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1415043Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1415988Z 	PASSED for resource: jobs(security-gate-enforcement).steps[3](Update commit status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1416717Z 	File: /.github/workflows/security-gates.yml:386-405
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1417483Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1418758Z 	PASSED for resource: jobs(security-gate-enforcement).steps[4](Add PR comment with security status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1419516Z 	File: /.github/workflows/security-gates.yml:404-448
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1420270Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1421299Z 	PASSED for resource: jobs(security-gate-enforcement).steps[5](Generate security gates summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1422127Z 	File: /.github/workflows/security-gates.yml:447-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1423074Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1423771Z 	PASSED for resource: jobs(policy-validation).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1424437Z 	File: /.github/workflows/security-gates.yml:36-42
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1425029Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1425727Z 	PASSED for resource: jobs(policy-validation).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1426379Z 	File: /.github/workflows/security-gates.yml:41-48
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1426976Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1427951Z 	PASSED for resource: jobs(policy-validation).steps[3](Validate security policy compliance)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1428736Z 	File: /.github/workflows/security-gates.yml:47-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1429001Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1429315Z 	PASSED for resource: jobs(pre-commit-security).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1429528Z 	File: /.github/workflows/security-gates.yml:160-166
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1429776Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1430091Z 	PASSED for resource: jobs(pre-commit-security).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1430292Z 	File: /.github/workflows/security-gates.yml:165-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1430746Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1431040Z 	PASSED for resource: jobs(pre-commit-security).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1431236Z 	File: /.github/workflows/security-gates.yml:171-177
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1431482Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1431881Z 	PASSED for resource: jobs(pre-commit-security).steps[4](Run pre-commit security hooks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1432077Z 	File: /.github/workflows/security-gates.yml:176-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1432320Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1432628Z 	PASSED for resource: jobs(branch-protection).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1432828Z 	File: /.github/workflows/security-gates.yml:220-226
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1433077Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1433467Z 	PASSED for resource: jobs(branch-protection).steps[2](Check branch protection rules)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1433669Z 	File: /.github/workflows/security-gates.yml:225-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1433912Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1434253Z 	PASSED for resource: jobs(security-gate-enforcement).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1434446Z 	File: /.github/workflows/security-gates.yml:340-346
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1434689Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1435091Z 	PASSED for resource: jobs(security-gate-enforcement).steps[2](Evaluate security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1435285Z 	File: /.github/workflows/security-gates.yml:345-387
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1435535Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1435919Z 	PASSED for resource: jobs(security-gate-enforcement).steps[3](Update commit status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1436118Z 	File: /.github/workflows/security-gates.yml:386-405
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1436368Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1436860Z 	PASSED for resource: jobs(security-gate-enforcement).steps[4](Add PR comment with security status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1437051Z 	File: /.github/workflows/security-gates.yml:404-448
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1437294Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1437921Z 	PASSED for resource: jobs(security-gate-enforcement).steps[5](Generate security gates summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1438136Z 	File: /.github/workflows/security-gates.yml:447-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1438482Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1439035Z 	PASSED for resource: jobs(policy-validation).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1439240Z 	File: /.github/workflows/security-gates.yml:36-42
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1439588Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1439895Z 	PASSED for resource: jobs(policy-validation).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1440092Z 	File: /.github/workflows/security-gates.yml:41-48
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1440440Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1440859Z 	PASSED for resource: jobs(policy-validation).steps[3](Validate security policy compliance)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1441631Z 	File: /.github/workflows/security-gates.yml:47-154
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1442030Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1442344Z 	PASSED for resource: jobs(pre-commit-security).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1442909Z 	File: /.github/workflows/security-gates.yml:160-166
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1443254Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1443550Z 	PASSED for resource: jobs(pre-commit-security).steps[2](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1443736Z 	File: /.github/workflows/security-gates.yml:165-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1444140Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1444740Z 	PASSED for resource: jobs(pre-commit-security).steps[3](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1444952Z 	File: /.github/workflows/security-gates.yml:171-177
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1445305Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1445566Z 	PASSED for resource: jobs(pre-commit-security).steps[4](Run pre-commit security hooks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1445687Z 	File: /.github/workflows/security-gates.yml:176-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1445881Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446064Z 	PASSED for resource: jobs(branch-protection).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446185Z 	File: /.github/workflows/security-gates.yml:220-226
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446379Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446610Z 	PASSED for resource: jobs(branch-protection).steps[2](Check branch protection rules)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446734Z 	File: /.github/workflows/security-gates.yml:225-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1446920Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1447116Z 	PASSED for resource: jobs(security-gate-enforcement).steps[1](Harden Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1447225Z 	File: /.github/workflows/security-gates.yml:340-346
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1447407Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1447851Z 	PASSED for resource: jobs(security-gate-enforcement).steps[2](Evaluate security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1448061Z 	File: /.github/workflows/security-gates.yml:345-387
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1448268Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1448649Z 	PASSED for resource: jobs(security-gate-enforcement).steps[3](Update commit status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1448859Z 	File: /.github/workflows/security-gates.yml:386-405
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1449197Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1449701Z 	PASSED for resource: jobs(security-gate-enforcement).steps[4](Add PR comment with security status)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1449910Z 	File: /.github/workflows/security-gates.yml:404-448
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1450255Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1450718Z 	PASSED for resource: jobs(security-gate-enforcement).steps[5](Generate security gates summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1450913Z 	File: /.github/workflows/security-gates.yml:447-491
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1452148Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1452368Z 	PASSED for resource: on(Security Gates Enhanced)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1452611Z 	File: /.github/workflows/security-gates-enhanced.yml:8-27
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1452862Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1453081Z 	PASSED for resource: jobs(security-preflight)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1453337Z 	File: /.github/workflows/security-gates-enhanced.yml:43-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1453847Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454037Z 	PASSED for resource: jobs(secret-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454197Z 	File: /.github/workflows/security-gates-enhanced.yml:89-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454365Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454472Z 	PASSED for resource: jobs(sast-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454634Z 	File: /.github/workflows/security-gates-enhanced.yml:147-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454769Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1454966Z 	PASSED for resource: jobs(dependency-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1455231Z 	File: /.github/workflows/security-gates-enhanced.yml:205-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1455469Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1455920Z 	PASSED for resource: jobs(container-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456120Z 	File: /.github/workflows/security-gates-enhanced.yml:281-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456273Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456375Z 	PASSED for resource: jobs(iac-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456522Z 	File: /.github/workflows/security-gates-enhanced.yml:375-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456653Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456763Z 	PASSED for resource: jobs(compliance-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1456911Z 	File: /.github/workflows/security-gates-enhanced.yml:449-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1457171Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1457263Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1457415Z 	File: /.github/workflows/security-gates-enhanced.yml:42-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1457871Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1458081Z 	PASSED for resource: jobs(security-preflight)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1458318Z 	File: /.github/workflows/security-gates-enhanced.yml:43-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1458607Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1458713Z 	PASSED for resource: jobs(secret-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1458854Z 	File: /.github/workflows/security-gates-enhanced.yml:89-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459088Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459190Z 	PASSED for resource: jobs(sast-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459331Z 	File: /.github/workflows/security-gates-enhanced.yml:147-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459561Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459674Z 	PASSED for resource: jobs(dependency-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1459811Z 	File: /.github/workflows/security-gates-enhanced.yml:205-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460051Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460156Z 	PASSED for resource: jobs(container-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460295Z 	File: /.github/workflows/security-gates-enhanced.yml:281-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460533Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460627Z 	PASSED for resource: jobs(iac-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1460764Z 	File: /.github/workflows/security-gates-enhanced.yml:375-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461148Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461253Z 	PASSED for resource: jobs(compliance-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461390Z 	File: /.github/workflows/security-gates-enhanced.yml:449-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461529Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461645Z 	PASSED for resource: jobs(security-preflight)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461780Z 	File: /.github/workflows/security-gates-enhanced.yml:43-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1461919Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462018Z 	PASSED for resource: jobs(secret-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462152Z 	File: /.github/workflows/security-gates-enhanced.yml:89-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462287Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462382Z 	PASSED for resource: jobs(sast-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462519Z 	File: /.github/workflows/security-gates-enhanced.yml:147-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462664Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462773Z 	PASSED for resource: jobs(dependency-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1462910Z 	File: /.github/workflows/security-gates-enhanced.yml:205-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463049Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463264Z 	PASSED for resource: jobs(container-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463403Z 	File: /.github/workflows/security-gates-enhanced.yml:281-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463543Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463636Z 	PASSED for resource: jobs(iac-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463773Z 	File: /.github/workflows/security-gates-enhanced.yml:375-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1463907Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464014Z 	PASSED for resource: jobs(compliance-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464150Z 	File: /.github/workflows/security-gates-enhanced.yml:449-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464345Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464457Z 	PASSED for resource: jobs(security-preflight)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464591Z 	File: /.github/workflows/security-gates-enhanced.yml:43-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464769Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1464881Z 	PASSED for resource: jobs(secret-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465015Z 	File: /.github/workflows/security-gates-enhanced.yml:89-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465189Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465291Z 	PASSED for resource: jobs(sast-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465427Z 	File: /.github/workflows/security-gates-enhanced.yml:147-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465602Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465717Z 	PASSED for resource: jobs(dependency-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1465860Z 	File: /.github/workflows/security-gates-enhanced.yml:205-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466034Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466145Z 	PASSED for resource: jobs(container-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466281Z 	File: /.github/workflows/security-gates-enhanced.yml:281-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466453Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466554Z 	PASSED for resource: jobs(iac-scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466689Z 	File: /.github/workflows/security-gates-enhanced.yml:375-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466862Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1466970Z 	PASSED for resource: jobs(compliance-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1467106Z 	File: /.github/workflows/security-gates-enhanced.yml:449-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1467355Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1467530Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1467841Z 	File: /.github/workflows/security-gates-enhanced.yml:42-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1467983Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468170Z 	PASSED for resource: jobs(security-preflight).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468304Z 	File: /.github/workflows/security-gates-enhanced.yml:51-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468432Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468638Z 	PASSED for resource: jobs(security-preflight).steps[2](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468769Z 	File: /.github/workflows/security-gates-enhanced.yml:56-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1468890Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469055Z 	PASSED for resource: jobs(secret-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469191Z 	File: /.github/workflows/security-gates-enhanced.yml:95-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469311Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469519Z 	PASSED for resource: jobs(secret-scanning).steps[2](Run TruffleHog secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469657Z 	File: /.github/workflows/security-gates-enhanced.yml:101-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469777Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1469971Z 	PASSED for resource: jobs(secret-scanning).steps[3](Run GitLeaks secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1470234Z 	File: /.github/workflows/security-gates-enhanced.yml:110-117
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1470360Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1470587Z 	PASSED for resource: jobs(secret-scanning).steps[4](Custom secret pattern validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1470725Z 	File: /.github/workflows/security-gates-enhanced.yml:116-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1470845Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471007Z 	PASSED for resource: jobs(sast-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471150Z 	File: /.github/workflows/security-gates-enhanced.yml:153-157
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471272Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471417Z 	PASSED for resource: jobs(sast-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471556Z 	File: /.github/workflows/security-gates-enhanced.yml:156-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471675Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471844Z 	PASSED for resource: jobs(sast-scanning).steps[3](Run Gosec SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1471980Z 	File: /.github/workflows/security-gates-enhanced.yml:163-183
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472097Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472268Z 	PASSED for resource: jobs(sast-scanning).steps[4](Run Semgrep SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472403Z 	File: /.github/workflows/security-gates-enhanced.yml:182-197
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472521Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472697Z 	PASSED for resource: jobs(sast-scanning).steps[5](Upload SAST results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472836Z 	File: /.github/workflows/security-gates-enhanced.yml:196-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1472954Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473128Z 	PASSED for resource: jobs(dependency-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473264Z 	File: /.github/workflows/security-gates-enhanced.yml:211-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473384Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473547Z 	PASSED for resource: jobs(dependency-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473683Z 	File: /.github/workflows/security-gates-enhanced.yml:214-222
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473801Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1473983Z 	PASSED for resource: jobs(dependency-scanning).steps[3](Run govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1474120Z 	File: /.github/workflows/security-gates-enhanced.yml:221-239
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1474239Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1474571Z 	PASSED for resource: jobs(dependency-scanning).steps[4](Run Nancy dependency scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1474709Z 	File: /.github/workflows/security-gates-enhanced.yml:238-252
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1474829Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475038Z 	PASSED for resource: jobs(dependency-scanning).steps[5](License compliance check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475177Z 	File: /.github/workflows/security-gates-enhanced.yml:251-274
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475295Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475499Z 	PASSED for resource: jobs(dependency-scanning).steps[6](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475635Z 	File: /.github/workflows/security-gates-enhanced.yml:273-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475754Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1475921Z 	PASSED for resource: jobs(container-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476062Z 	File: /.github/workflows/security-gates-enhanced.yml:288-293
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476185Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476412Z 	PASSED for resource: jobs(container-scanning).steps[2](Build container for security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476553Z 	File: /.github/workflows/security-gates-enhanced.yml:292-315
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476670Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1476952Z 	PASSED for resource: jobs(container-scanning).steps[3](Run Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1477093Z 	File: /.github/workflows/security-gates-enhanced.yml:314-327
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1477211Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1477400Z 	PASSED for resource: jobs(container-scanning).steps[4](Analyze Trivy results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1477540Z 	File: /.github/workflows/security-gates-enhanced.yml:326-358
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1477831Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478042Z 	PASSED for resource: jobs(container-scanning).steps[5](Run Grype security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478186Z 	File: /.github/workflows/security-gates-enhanced.yml:357-367
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478311Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478527Z 	PASSED for resource: jobs(container-scanning).steps[6](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478673Z 	File: /.github/workflows/security-gates-enhanced.yml:366-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478791Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1478938Z 	PASSED for resource: jobs(iac-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479080Z 	File: /.github/workflows/security-gates-enhanced.yml:381-386
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479197Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479362Z 	PASSED for resource: jobs(iac-scanning).steps[2](Run Checkov IaC scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479501Z 	File: /.github/workflows/security-gates-enhanced.yml:385-397
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479623Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479803Z 	PASSED for resource: jobs(iac-scanning).steps[3](Run TFSec Terraform scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1479942Z 	File: /.github/workflows/security-gates-enhanced.yml:396-404
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480060Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480257Z 	PASSED for resource: jobs(iac-scanning).steps[4](Docker security best practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480401Z 	File: /.github/workflows/security-gates-enhanced.yml:403-441
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480515Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480690Z 	PASSED for resource: jobs(iac-scanning).steps[5](Upload IaC scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480829Z 	File: /.github/workflows/security-gates-enhanced.yml:440-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1480946Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481114Z 	PASSED for resource: jobs(compliance-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481382Z 	File: /.github/workflows/security-gates-enhanced.yml:456-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481501Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481709Z 	PASSED for resource: jobs(compliance-check).steps[2](Security compliance analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481849Z 	File: /.github/workflows/security-gates-enhanced.yml:460-531
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1481971Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482170Z 	PASSED for resource: jobs(compliance-check).steps[3](Generate compliance report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482307Z 	File: /.github/workflows/security-gates-enhanced.yml:530-572
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482423Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482584Z 	PASSED for resource: jobs(compliance-check).steps[4](Comment on PR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482725Z 	File: /.github/workflows/security-gates-enhanced.yml:571-608
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1482843Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1483034Z 	PASSED for resource: jobs(compliance-check).steps[5](Enforce security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1483175Z 	File: /.github/workflows/security-gates-enhanced.yml:607-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1483411Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1483581Z 	PASSED for resource: jobs(security-preflight).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1483837Z 	File: /.github/workflows/security-gates-enhanced.yml:51-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484069Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484268Z 	PASSED for resource: jobs(security-preflight).steps[2](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484407Z 	File: /.github/workflows/security-gates-enhanced.yml:56-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484636Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484795Z 	PASSED for resource: jobs(secret-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1484941Z 	File: /.github/workflows/security-gates-enhanced.yml:95-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1485169Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1485367Z 	PASSED for resource: jobs(secret-scanning).steps[2](Run TruffleHog secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1485512Z 	File: /.github/workflows/security-gates-enhanced.yml:101-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1485742Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1485929Z 	PASSED for resource: jobs(secret-scanning).steps[3](Run GitLeaks secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1486072Z 	File: /.github/workflows/security-gates-enhanced.yml:110-117
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1486300Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1486537Z 	PASSED for resource: jobs(secret-scanning).steps[4](Custom secret pattern validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1486694Z 	File: /.github/workflows/security-gates-enhanced.yml:116-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1486929Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1487095Z 	PASSED for resource: jobs(sast-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1487237Z 	File: /.github/workflows/security-gates-enhanced.yml:153-157
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1487472Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1487618Z 	PASSED for resource: jobs(sast-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1488060Z 	File: /.github/workflows/security-gates-enhanced.yml:156-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1488301Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1488467Z 	PASSED for resource: jobs(sast-scanning).steps[3](Run Gosec SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1488608Z 	File: /.github/workflows/security-gates-enhanced.yml:163-183
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1488974Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1489146Z 	PASSED for resource: jobs(sast-scanning).steps[4](Run Semgrep SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1489283Z 	File: /.github/workflows/security-gates-enhanced.yml:182-197
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1489509Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1489694Z 	PASSED for resource: jobs(sast-scanning).steps[5](Upload SAST results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1489831Z 	File: /.github/workflows/security-gates-enhanced.yml:196-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490059Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490247Z 	PASSED for resource: jobs(dependency-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490392Z 	File: /.github/workflows/security-gates-enhanced.yml:211-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490623Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490794Z 	PASSED for resource: jobs(dependency-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1490935Z 	File: /.github/workflows/security-gates-enhanced.yml:214-222
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1491165Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1491353Z 	PASSED for resource: jobs(dependency-scanning).steps[3](Run govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1491603Z 	File: /.github/workflows/security-gates-enhanced.yml:221-239
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1491833Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1492052Z 	PASSED for resource: jobs(dependency-scanning).steps[4](Run Nancy dependency scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1492191Z 	File: /.github/workflows/security-gates-enhanced.yml:238-252
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1492419Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1492642Z 	PASSED for resource: jobs(dependency-scanning).steps[5](License compliance check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1492794Z 	File: /.github/workflows/security-gates-enhanced.yml:251-274
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493022Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493220Z 	PASSED for resource: jobs(dependency-scanning).steps[6](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493370Z 	File: /.github/workflows/security-gates-enhanced.yml:273-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493599Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493769Z 	PASSED for resource: jobs(container-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1493912Z 	File: /.github/workflows/security-gates-enhanced.yml:288-293
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1494141Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1494375Z 	PASSED for resource: jobs(container-scanning).steps[2](Build container for security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1494524Z 	File: /.github/workflows/security-gates-enhanced.yml:292-315
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1494751Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1494948Z 	PASSED for resource: jobs(container-scanning).steps[3](Run Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1495094Z 	File: /.github/workflows/security-gates-enhanced.yml:314-327
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1495327Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1495525Z 	PASSED for resource: jobs(container-scanning).steps[4](Analyze Trivy results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1495662Z 	File: /.github/workflows/security-gates-enhanced.yml:326-358
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1495890Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1496092Z 	PASSED for resource: jobs(container-scanning).steps[5](Run Grype security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1496314Z 	File: /.github/workflows/security-gates-enhanced.yml:357-367
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1496543Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1496770Z 	PASSED for resource: jobs(container-scanning).steps[6](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1496909Z 	File: /.github/workflows/security-gates-enhanced.yml:366-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1497139Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1497296Z 	PASSED for resource: jobs(iac-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1497435Z 	File: /.github/workflows/security-gates-enhanced.yml:381-386
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1497841Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498033Z 	PASSED for resource: jobs(iac-scanning).steps[2](Run Checkov IaC scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498172Z 	File: /.github/workflows/security-gates-enhanced.yml:385-397
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498414Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498601Z 	PASSED for resource: jobs(iac-scanning).steps[3](Run TFSec Terraform scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498738Z 	File: /.github/workflows/security-gates-enhanced.yml:396-404
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1498969Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1499389Z 	PASSED for resource: jobs(iac-scanning).steps[4](Docker security best practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1499529Z 	File: /.github/workflows/security-gates-enhanced.yml:403-441
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1499757Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1499940Z 	PASSED for resource: jobs(iac-scanning).steps[5](Upload IaC scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1500079Z 	File: /.github/workflows/security-gates-enhanced.yml:440-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1500307Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1500483Z 	PASSED for resource: jobs(compliance-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1500619Z 	File: /.github/workflows/security-gates-enhanced.yml:456-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1500847Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1501095Z 	PASSED for resource: jobs(compliance-check).steps[2](Security compliance analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1501238Z 	File: /.github/workflows/security-gates-enhanced.yml:460-531
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1501468Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1501675Z 	PASSED for resource: jobs(compliance-check).steps[3](Generate compliance report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1501811Z 	File: /.github/workflows/security-gates-enhanced.yml:530-572
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502038Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502216Z 	PASSED for resource: jobs(compliance-check).steps[4](Comment on PR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502352Z 	File: /.github/workflows/security-gates-enhanced.yml:571-608
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502579Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502777Z 	PASSED for resource: jobs(compliance-check).steps[5](Enforce security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1502916Z 	File: /.github/workflows/security-gates-enhanced.yml:607-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503053Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503235Z 	PASSED for resource: jobs(security-preflight).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503374Z 	File: /.github/workflows/security-gates-enhanced.yml:51-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503507Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503710Z 	PASSED for resource: jobs(security-preflight).steps[2](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1503997Z 	File: /.github/workflows/security-gates-enhanced.yml:56-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504130Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504295Z 	PASSED for resource: jobs(secret-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504431Z 	File: /.github/workflows/security-gates-enhanced.yml:95-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504562Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504771Z 	PASSED for resource: jobs(secret-scanning).steps[2](Run TruffleHog secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1504910Z 	File: /.github/workflows/security-gates-enhanced.yml:101-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505042Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505237Z 	PASSED for resource: jobs(secret-scanning).steps[3](Run GitLeaks secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505375Z 	File: /.github/workflows/security-gates-enhanced.yml:110-117
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505507Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505735Z 	PASSED for resource: jobs(secret-scanning).steps[4](Custom secret pattern validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1505872Z 	File: /.github/workflows/security-gates-enhanced.yml:116-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506004Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506165Z 	PASSED for resource: jobs(sast-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506302Z 	File: /.github/workflows/security-gates-enhanced.yml:153-157
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506514Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506655Z 	PASSED for resource: jobs(sast-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506792Z 	File: /.github/workflows/security-gates-enhanced.yml:156-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1506921Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507088Z 	PASSED for resource: jobs(sast-scanning).steps[3](Run Gosec SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507228Z 	File: /.github/workflows/security-gates-enhanced.yml:163-183
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507367Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507534Z 	PASSED for resource: jobs(sast-scanning).steps[4](Run Semgrep SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507839Z 	File: /.github/workflows/security-gates-enhanced.yml:182-197
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1507982Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508159Z 	PASSED for resource: jobs(sast-scanning).steps[5](Upload SAST results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508302Z 	File: /.github/workflows/security-gates-enhanced.yml:196-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508434Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508609Z 	PASSED for resource: jobs(dependency-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508744Z 	File: /.github/workflows/security-gates-enhanced.yml:211-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1508874Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509039Z 	PASSED for resource: jobs(dependency-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509182Z 	File: /.github/workflows/security-gates-enhanced.yml:214-222
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509311Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509495Z 	PASSED for resource: jobs(dependency-scanning).steps[3](Run govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509631Z 	File: /.github/workflows/security-gates-enhanced.yml:221-239
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509762Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1509979Z 	PASSED for resource: jobs(dependency-scanning).steps[4](Run Nancy dependency scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1510116Z 	File: /.github/workflows/security-gates-enhanced.yml:238-252
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1510249Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1510460Z 	PASSED for resource: jobs(dependency-scanning).steps[5](License compliance check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1510598Z 	File: /.github/workflows/security-gates-enhanced.yml:251-274
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1510732Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511060Z 	PASSED for resource: jobs(dependency-scanning).steps[6](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511198Z 	File: /.github/workflows/security-gates-enhanced.yml:273-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511329Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511508Z 	PASSED for resource: jobs(container-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511652Z 	File: /.github/workflows/security-gates-enhanced.yml:288-293
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1511786Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512028Z 	PASSED for resource: jobs(container-scanning).steps[2](Build container for security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512165Z 	File: /.github/workflows/security-gates-enhanced.yml:292-315
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512296Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512507Z 	PASSED for resource: jobs(container-scanning).steps[3](Run Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512648Z 	File: /.github/workflows/security-gates-enhanced.yml:314-327
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512777Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1512971Z 	PASSED for resource: jobs(container-scanning).steps[4](Analyze Trivy results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1513108Z 	File: /.github/workflows/security-gates-enhanced.yml:326-358
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1513237Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1513547Z 	PASSED for resource: jobs(container-scanning).steps[5](Run Grype security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1513684Z 	File: /.github/workflows/security-gates-enhanced.yml:357-367
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1513816Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514037Z 	PASSED for resource: jobs(container-scanning).steps[6](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514172Z 	File: /.github/workflows/security-gates-enhanced.yml:366-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514302Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514463Z 	PASSED for resource: jobs(iac-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514602Z 	File: /.github/workflows/security-gates-enhanced.yml:381-386
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514730Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1514906Z 	PASSED for resource: jobs(iac-scanning).steps[2](Run Checkov IaC scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515045Z 	File: /.github/workflows/security-gates-enhanced.yml:385-397
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515174Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515357Z 	PASSED for resource: jobs(iac-scanning).steps[3](Run TFSec Terraform scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515491Z 	File: /.github/workflows/security-gates-enhanced.yml:396-404
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515621Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515826Z 	PASSED for resource: jobs(iac-scanning).steps[4](Docker security best practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1515963Z 	File: /.github/workflows/security-gates-enhanced.yml:403-441
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516097Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516278Z 	PASSED for resource: jobs(iac-scanning).steps[5](Upload IaC scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516414Z 	File: /.github/workflows/security-gates-enhanced.yml:440-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516543Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516716Z 	PASSED for resource: jobs(compliance-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516852Z 	File: /.github/workflows/security-gates-enhanced.yml:456-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1516981Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1517194Z 	PASSED for resource: jobs(compliance-check).steps[2](Security compliance analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1517329Z 	File: /.github/workflows/security-gates-enhanced.yml:460-531
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1517461Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1517828Z 	PASSED for resource: jobs(compliance-check).steps[3](Generate compliance report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518098Z 	File: /.github/workflows/security-gates-enhanced.yml:530-572
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518234Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518404Z 	PASSED for resource: jobs(compliance-check).steps[4](Comment on PR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518544Z 	File: /.github/workflows/security-gates-enhanced.yml:571-608
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518681Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1518876Z 	PASSED for resource: jobs(compliance-check).steps[5](Enforce security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519012Z 	File: /.github/workflows/security-gates-enhanced.yml:607-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519188Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519362Z 	PASSED for resource: jobs(security-preflight).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519497Z 	File: /.github/workflows/security-gates-enhanced.yml:51-57
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519677Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1519885Z 	PASSED for resource: jobs(security-preflight).steps[2](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1520018Z 	File: /.github/workflows/security-gates-enhanced.yml:56-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1520194Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1520493Z 	PASSED for resource: jobs(secret-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1520630Z 	File: /.github/workflows/security-gates-enhanced.yml:95-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1520805Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521008Z 	PASSED for resource: jobs(secret-scanning).steps[2](Run TruffleHog secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521146Z 	File: /.github/workflows/security-gates-enhanced.yml:101-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521320Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521516Z 	PASSED for resource: jobs(secret-scanning).steps[3](Run GitLeaks secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521653Z 	File: /.github/workflows/security-gates-enhanced.yml:110-117
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1521825Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522049Z 	PASSED for resource: jobs(secret-scanning).steps[4](Custom secret pattern validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522190Z 	File: /.github/workflows/security-gates-enhanced.yml:116-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522365Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522526Z 	PASSED for resource: jobs(sast-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522662Z 	File: /.github/workflows/security-gates-enhanced.yml:153-157
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522835Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1522983Z 	PASSED for resource: jobs(sast-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523124Z 	File: /.github/workflows/security-gates-enhanced.yml:156-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523298Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523459Z 	PASSED for resource: jobs(sast-scanning).steps[3](Run Gosec SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523595Z 	File: /.github/workflows/security-gates-enhanced.yml:163-183
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523769Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1523940Z 	PASSED for resource: jobs(sast-scanning).steps[4](Run Semgrep SAST)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1524079Z 	File: /.github/workflows/security-gates-enhanced.yml:182-197
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1524252Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1524427Z 	PASSED for resource: jobs(sast-scanning).steps[5](Upload SAST results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1524562Z 	File: /.github/workflows/security-gates-enhanced.yml:196-205
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1524736Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525005Z 	PASSED for resource: jobs(dependency-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525143Z 	File: /.github/workflows/security-gates-enhanced.yml:211-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525318Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525483Z 	PASSED for resource: jobs(dependency-scanning).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525627Z 	File: /.github/workflows/security-gates-enhanced.yml:214-222
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525799Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1525986Z 	PASSED for resource: jobs(dependency-scanning).steps[3](Run govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1526122Z 	File: /.github/workflows/security-gates-enhanced.yml:221-239
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1526293Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1526508Z 	PASSED for resource: jobs(dependency-scanning).steps[4](Run Nancy dependency scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1526648Z 	File: /.github/workflows/security-gates-enhanced.yml:238-252
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1526821Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1527033Z 	PASSED for resource: jobs(dependency-scanning).steps[5](License compliance check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1527167Z 	File: /.github/workflows/security-gates-enhanced.yml:251-274
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1527424Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1527774Z 	PASSED for resource: jobs(dependency-scanning).steps[6](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1527972Z 	File: /.github/workflows/security-gates-enhanced.yml:273-281
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1528149Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1528325Z 	PASSED for resource: jobs(container-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1528467Z 	File: /.github/workflows/security-gates-enhanced.yml:288-293
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1528643Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1528877Z 	PASSED for resource: jobs(container-scanning).steps[2](Build container for security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529013Z 	File: /.github/workflows/security-gates-enhanced.yml:292-315
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529188Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529400Z 	PASSED for resource: jobs(container-scanning).steps[3](Run Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529538Z 	File: /.github/workflows/security-gates-enhanced.yml:314-327
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529712Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1529906Z 	PASSED for resource: jobs(container-scanning).steps[4](Analyze Trivy results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530045Z 	File: /.github/workflows/security-gates-enhanced.yml:326-358
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530218Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530421Z 	PASSED for resource: jobs(container-scanning).steps[5](Run Grype security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530556Z 	File: /.github/workflows/security-gates-enhanced.yml:357-367
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530728Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1530951Z 	PASSED for resource: jobs(container-scanning).steps[6](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531092Z 	File: /.github/workflows/security-gates-enhanced.yml:366-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531262Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531415Z 	PASSED for resource: jobs(iac-scanning).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531552Z 	File: /.github/workflows/security-gates-enhanced.yml:381-386
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531723Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1531897Z 	PASSED for resource: jobs(iac-scanning).steps[2](Run Checkov IaC scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1532163Z 	File: /.github/workflows/security-gates-enhanced.yml:385-397
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1532337Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1532524Z 	PASSED for resource: jobs(iac-scanning).steps[3](Run TFSec Terraform scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1532662Z 	File: /.github/workflows/security-gates-enhanced.yml:396-404
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1532843Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533043Z 	PASSED for resource: jobs(iac-scanning).steps[4](Docker security best practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533177Z 	File: /.github/workflows/security-gates-enhanced.yml:403-441
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533350Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533530Z 	PASSED for resource: jobs(iac-scanning).steps[5](Upload IaC scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533667Z 	File: /.github/workflows/security-gates-enhanced.yml:440-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1533844Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534014Z 	PASSED for resource: jobs(compliance-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534149Z 	File: /.github/workflows/security-gates-enhanced.yml:456-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534321Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534642Z 	PASSED for resource: jobs(compliance-check).steps[2](Security compliance analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534779Z 	File: /.github/workflows/security-gates-enhanced.yml:460-531
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1534950Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535154Z 	PASSED for resource: jobs(compliance-check).steps[3](Generate compliance report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535291Z 	File: /.github/workflows/security-gates-enhanced.yml:530-572
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535469Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535637Z 	PASSED for resource: jobs(compliance-check).steps[4](Comment on PR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535771Z 	File: /.github/workflows/security-gates-enhanced.yml:571-608
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1535949Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1536134Z 	PASSED for resource: jobs(compliance-check).steps[5](Enforce security gates)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1536274Z 	File: /.github/workflows/security-gates-enhanced.yml:607-636
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1536827Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1536920Z 	PASSED for resource: on(CI)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537015Z 	File: /.github/workflows/ci.yml:4-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537151Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537242Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537331Z 	File: /.github/workflows/ci.yml:29-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537467Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537555Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537810Z 	File: /.github/workflows/ci.yml:64-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1537971Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538075Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538174Z 	File: /.github/workflows/ci.yml:93-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538304Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538399Z 	PASSED for resource: jobs(docker)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538492Z 	File: /.github/workflows/ci.yml:120-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538742Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538833Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1538922Z 	File: /.github/workflows/ci.yml:28-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539158Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539380Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539468Z 	File: /.github/workflows/ci.yml:29-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539700Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539789Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1539874Z 	File: /.github/workflows/ci.yml:64-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540105Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540197Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540282Z 	File: /.github/workflows/ci.yml:93-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540508Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540598Z 	PASSED for resource: jobs(docker)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540685Z 	File: /.github/workflows/ci.yml:120-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540826Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1540920Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541004Z 	File: /.github/workflows/ci.yml:29-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541140Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541231Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541317Z 	File: /.github/workflows/ci.yml:64-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541446Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541652Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541737Z 	File: /.github/workflows/ci.yml:93-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541869Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1541960Z 	PASSED for resource: jobs(docker)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542046Z 	File: /.github/workflows/ci.yml:120-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542225Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542316Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542403Z 	File: /.github/workflows/ci.yml:29-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542581Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542670Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542754Z 	File: /.github/workflows/ci.yml:64-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1542928Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543019Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543109Z 	File: /.github/workflows/ci.yml:93-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543285Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543368Z 	PASSED for resource: jobs(docker)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543458Z 	File: /.github/workflows/ci.yml:120-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543709Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543792Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1543884Z 	File: /.github/workflows/ci.yml:28-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544009Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544132Z 	PASSED for resource: jobs(test).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544223Z 	File: /.github/workflows/ci.yml:34-38
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544342Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544457Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544548Z 	File: /.github/workflows/ci.yml:37-44
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544672Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544820Z 	PASSED for resource: jobs(test).steps[3](Verify dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1544909Z 	File: /.github/workflows/ci.yml:43-49
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1545028Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1545139Z 	PASSED for resource: jobs(test).steps[4](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1545298Z 	File: /.github/workflows/ci.yml:48-52
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1545504Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1545691Z 	PASSED for resource: jobs(test).steps[5](Test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1546287Z 	File: /.github/workflows/ci.yml:51-55
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1546539Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1546747Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1546846Z 	File: /.github/workflows/ci.yml:54-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1546975Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547103Z 	PASSED for resource: jobs(lint).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547194Z 	File: /.github/workflows/ci.yml:69-73
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547314Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547427Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547515Z 	File: /.github/workflows/ci.yml:72-79
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547802Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1547963Z 	PASSED for resource: jobs(lint).steps[3](Format check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548050Z 	File: /.github/workflows/ci.yml:78-87
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548190Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548352Z 	PASSED for resource: jobs(lint).steps[4](Lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548490Z 	File: /.github/workflows/ci.yml:86-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548679Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1548865Z 	PASSED for resource: jobs(security).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1549207Z 	File: /.github/workflows/ci.yml:98-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1549423Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1549627Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1549777Z 	File: /.github/workflows/ci.yml:101-108
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1549986Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550227Z 	PASSED for resource: jobs(security).steps[3](Security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550374Z 	File: /.github/workflows/ci.yml:107-113
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550508Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550687Z 	PASSED for resource: jobs(security).steps[4](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550774Z 	File: /.github/workflows/ci.yml:112-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1550900Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551023Z 	PASSED for resource: jobs(docker).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551109Z 	File: /.github/workflows/ci.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551239Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551468Z 	PASSED for resource: jobs(docker).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551619Z 	File: /.github/workflows/ci.yml:133-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551799Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1551918Z 	PASSED for resource: jobs(docker).steps[3](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552005Z 	File: /.github/workflows/ci.yml:136-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552138Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552262Z 	PASSED for resource: jobs(docker).steps[4](Test image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552349Z 	File: /.github/workflows/ci.yml:146-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552625Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552814Z 	PASSED for resource: jobs(test).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1552906Z 	File: /.github/workflows/ci.yml:34-38
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553157Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553268Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553352Z 	File: /.github/workflows/ci.yml:37-44
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553584Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553732Z 	PASSED for resource: jobs(test).steps[3](Verify dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1553817Z 	File: /.github/workflows/ci.yml:43-49
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554053Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554311Z 	PASSED for resource: jobs(test).steps[4](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554395Z 	File: /.github/workflows/ci.yml:48-52
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554631Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554738Z 	PASSED for resource: jobs(test).steps[5](Test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1554827Z 	File: /.github/workflows/ci.yml:51-55
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555065Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555199Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555282Z 	File: /.github/workflows/ci.yml:54-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555554Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555745Z 	PASSED for resource: jobs(lint).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1555889Z 	File: /.github/workflows/ci.yml:69-73
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1556301Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1556481Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1556569Z 	File: /.github/workflows/ci.yml:72-79
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1556838Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1557184Z 	PASSED for resource: jobs(lint).steps[3](Format check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1557273Z 	File: /.github/workflows/ci.yml:78-87
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1557510Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1557617Z 	PASSED for resource: jobs(lint).steps[4](Lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1557978Z 	File: /.github/workflows/ci.yml:86-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558258Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558380Z 	PASSED for resource: jobs(security).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558477Z 	File: /.github/workflows/ci.yml:98-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558707Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558824Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1558910Z 	File: /.github/workflows/ci.yml:101-108
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559142Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559290Z 	PASSED for resource: jobs(security).steps[3](Security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559378Z 	File: /.github/workflows/ci.yml:107-113
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559611Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559782Z 	PASSED for resource: jobs(security).steps[4](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1559869Z 	File: /.github/workflows/ci.yml:112-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560107Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560229Z 	PASSED for resource: jobs(docker).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560314Z 	File: /.github/workflows/ci.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560542Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560696Z 	PASSED for resource: jobs(docker).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1560784Z 	File: /.github/workflows/ci.yml:133-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561017Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561129Z 	PASSED for resource: jobs(docker).steps[3](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561215Z 	File: /.github/workflows/ci.yml:136-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561447Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561568Z 	PASSED for resource: jobs(docker).steps[4](Test image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561653Z 	File: /.github/workflows/ci.yml:146-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1561946Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562054Z 	PASSED for resource: jobs(test).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562138Z 	File: /.github/workflows/ci.yml:34-38
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562272Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562377Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562465Z 	File: /.github/workflows/ci.yml:37-44
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562592Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562745Z 	PASSED for resource: jobs(test).steps[3](Verify dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562827Z 	File: /.github/workflows/ci.yml:43-49
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1562963Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563078Z 	PASSED for resource: jobs(test).steps[4](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563163Z 	File: /.github/workflows/ci.yml:48-52
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563290Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563408Z 	PASSED for resource: jobs(test).steps[5](Test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563492Z 	File: /.github/workflows/ci.yml:51-55
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563621Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563759Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1563844Z 	File: /.github/workflows/ci.yml:54-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564085Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564198Z 	PASSED for resource: jobs(lint).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564280Z 	File: /.github/workflows/ci.yml:69-73
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564408Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564524Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564607Z 	File: /.github/workflows/ci.yml:72-79
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564734Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564863Z 	PASSED for resource: jobs(lint).steps[3](Format check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1564946Z 	File: /.github/workflows/ci.yml:78-87
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565074Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565183Z 	PASSED for resource: jobs(lint).steps[4](Lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565266Z 	File: /.github/workflows/ci.yml:86-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565393Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565520Z 	PASSED for resource: jobs(security).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565605Z 	File: /.github/workflows/ci.yml:98-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565731Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565854Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1565941Z 	File: /.github/workflows/ci.yml:101-108
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566068Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566208Z 	PASSED for resource: jobs(security).steps[3](Security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566297Z 	File: /.github/workflows/ci.yml:107-113
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566424Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566593Z 	PASSED for resource: jobs(security).steps[4](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566677Z 	File: /.github/workflows/ci.yml:112-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566804Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1566925Z 	PASSED for resource: jobs(docker).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567016Z 	File: /.github/workflows/ci.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567144Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567297Z 	PASSED for resource: jobs(docker).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567388Z 	File: /.github/workflows/ci.yml:133-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567515Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1567788Z 	PASSED for resource: jobs(docker).steps[3](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568021Z 	File: /.github/workflows/ci.yml:136-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568159Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568280Z 	PASSED for resource: jobs(docker).steps[4](Test image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568372Z 	File: /.github/workflows/ci.yml:146-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568557Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568671Z 	PASSED for resource: jobs(test).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568760Z 	File: /.github/workflows/ci.yml:34-38
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1568935Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569043Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569136Z 	File: /.github/workflows/ci.yml:37-44
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569309Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569452Z 	PASSED for resource: jobs(test).steps[3](Verify dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569547Z 	File: /.github/workflows/ci.yml:43-49
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569719Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569830Z 	PASSED for resource: jobs(test).steps[4](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1569933Z 	File: /.github/workflows/ci.yml:48-52
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570109Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570333Z 	PASSED for resource: jobs(test).steps[5](Test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570426Z 	File: /.github/workflows/ci.yml:51-55
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570607Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570740Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1570831Z 	File: /.github/workflows/ci.yml:54-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571002Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571111Z 	PASSED for resource: jobs(lint).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571206Z 	File: /.github/workflows/ci.yml:69-73
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571377Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571487Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571578Z 	File: /.github/workflows/ci.yml:72-79
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571750Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571872Z 	PASSED for resource: jobs(lint).steps[3](Format check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1571961Z 	File: /.github/workflows/ci.yml:78-87
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572133Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572236Z 	PASSED for resource: jobs(lint).steps[4](Lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572324Z 	File: /.github/workflows/ci.yml:86-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572494Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572611Z 	PASSED for resource: jobs(security).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572703Z 	File: /.github/workflows/ci.yml:98-102
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572874Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1572990Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573082Z 	File: /.github/workflows/ci.yml:101-108
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573252Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573401Z 	PASSED for resource: jobs(security).steps[3](Security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573491Z 	File: /.github/workflows/ci.yml:107-113
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573663Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573828Z 	PASSED for resource: jobs(security).steps[4](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1573916Z 	File: /.github/workflows/ci.yml:112-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574089Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574205Z 	PASSED for resource: jobs(docker).steps[1](Checkout)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574379Z 	File: /.github/workflows/ci.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574551Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574698Z 	PASSED for resource: jobs(docker).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574791Z 	File: /.github/workflows/ci.yml:133-137
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1574966Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1575077Z 	PASSED for resource: jobs(docker).steps[3](Build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1575161Z 	File: /.github/workflows/ci.yml:136-147
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1575343Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1575461Z 	PASSED for resource: jobs(docker).steps[4](Test image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1575548Z 	File: /.github/workflows/ci.yml:146-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576099Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576194Z 	PASSED for resource: on(Security Scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576295Z 	File: /.github/workflows/security.yml:5-13
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576417Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576517Z 	PASSED for resource: jobs(dependency-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576699Z 	File: /.github/workflows/security.yml:19-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576819Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1576909Z 	PASSED for resource: jobs(sast-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577001Z 	File: /.github/workflows/security.yml:43-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577124Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577224Z 	PASSED for resource: jobs(container-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577323Z 	File: /.github/workflows/security.yml:82-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577447Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577543Z 	PASSED for resource: jobs(license-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577826Z 	File: /.github/workflows/security.yml:118-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1577963Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578055Z 	PASSED for resource: jobs(secrets-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578153Z 	File: /.github/workflows/security.yml:146-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578277Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578370Z 	PASSED for resource: jobs(iac-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578467Z 	File: /.github/workflows/security.yml:169-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578590Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578692Z 	PASSED for resource: jobs(security-summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1578787Z 	File: /.github/workflows/security.yml:199-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579032Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579117Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579222Z 	File: /.github/workflows/security.yml:18-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579464Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579634Z 	PASSED for resource: jobs(dependency-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1579791Z 	File: /.github/workflows/security.yml:19-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580209Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580332Z 	PASSED for resource: jobs(sast-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580427Z 	File: /.github/workflows/security.yml:43-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580663Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580761Z 	PASSED for resource: jobs(container-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1580857Z 	File: /.github/workflows/security.yml:82-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1581144Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1581530Z 	PASSED for resource: jobs(license-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1581717Z 	File: /.github/workflows/security.yml:118-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1581991Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582091Z 	PASSED for resource: jobs(secrets-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582193Z 	File: /.github/workflows/security.yml:146-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582435Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582524Z 	PASSED for resource: jobs(iac-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582621Z 	File: /.github/workflows/security.yml:169-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582851Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1582954Z 	PASSED for resource: jobs(security-summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583051Z 	File: /.github/workflows/security.yml:199-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583199Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583305Z 	PASSED for resource: jobs(dependency-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583401Z 	File: /.github/workflows/security.yml:19-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583549Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583640Z 	PASSED for resource: jobs(sast-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583734Z 	File: /.github/workflows/security.yml:43-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1583869Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584099Z 	PASSED for resource: jobs(container-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584198Z 	File: /.github/workflows/security.yml:82-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584336Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584430Z 	PASSED for resource: jobs(license-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584529Z 	File: /.github/workflows/security.yml:118-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584656Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584752Z 	PASSED for resource: jobs(secrets-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584854Z 	File: /.github/workflows/security.yml:146-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1584980Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585076Z 	PASSED for resource: jobs(iac-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585173Z 	File: /.github/workflows/security.yml:169-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585301Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585413Z 	PASSED for resource: jobs(security-summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585509Z 	File: /.github/workflows/security.yml:199-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585687Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585793Z 	PASSED for resource: jobs(dependency-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1585886Z 	File: /.github/workflows/security.yml:19-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586062Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586159Z 	PASSED for resource: jobs(sast-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586251Z 	File: /.github/workflows/security.yml:43-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586432Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586536Z 	PASSED for resource: jobs(container-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586632Z 	File: /.github/workflows/security.yml:82-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586804Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1586902Z 	PASSED for resource: jobs(license-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587001Z 	File: /.github/workflows/security.yml:118-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587170Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587266Z 	PASSED for resource: jobs(secrets-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587360Z 	File: /.github/workflows/security.yml:146-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587534Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587791Z 	PASSED for resource: jobs(iac-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1587915Z 	File: /.github/workflows/security.yml:169-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588224Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588331Z 	PASSED for resource: jobs(security-summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588427Z 	File: /.github/workflows/security.yml:199-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588687Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588778Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1588886Z 	File: /.github/workflows/security.yml:18-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589012Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589193Z 	PASSED for resource: jobs(dependency-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589292Z 	File: /.github/workflows/security.yml:22-26
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589417Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589578Z 	PASSED for resource: jobs(dependency-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589677Z 	File: /.github/workflows/security.yml:25-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1589804Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590035Z 	PASSED for resource: jobs(dependency-scan).steps[3](Run Nancy (Go dependency scanner))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590128Z 	File: /.github/workflows/security.yml:31-37
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590250Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590426Z 	PASSED for resource: jobs(dependency-scan).steps[4](Run Govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590629Z 	File: /.github/workflows/security.yml:36-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590751Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1590902Z 	PASSED for resource: jobs(sast-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591067Z 	File: /.github/workflows/security.yml:46-50
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591281Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591467Z 	PASSED for resource: jobs(sast-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591569Z 	File: /.github/workflows/security.yml:49-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591698Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1591860Z 	PASSED for resource: jobs(sast-scan).steps[3](Install and run Gosec)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1592001Z 	File: /.github/workflows/security.yml:55-65
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1592212Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1592450Z 	PASSED for resource: jobs(sast-scan).steps[4](Upload Gosec results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1592622Z 	File: /.github/workflows/security.yml:64-70
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1592836Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593078Z 	PASSED for resource: jobs(sast-scan).steps[5](Run Semgrep)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593258Z 	File: /.github/workflows/security.yml:69-76
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593410Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593585Z 	PASSED for resource: jobs(sast-scan).steps[6](Upload Semgrep results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593686Z 	File: /.github/workflows/security.yml:75-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1593815Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1594136Z 	PASSED for resource: jobs(container-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1594303Z 	File: /.github/workflows/security.yml:85-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1594511Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1594787Z 	PASSED for resource: jobs(container-scan).steps[2](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1594946Z 	File: /.github/workflows/security.yml:88-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1595140Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1595402Z 	PASSED for resource: jobs(container-scan).steps[3](Run Trivy scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1595560Z 	File: /.github/workflows/security.yml:92-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1595772Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1596104Z 	PASSED for resource: jobs(container-scan).steps[4](Upload Trivy scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1596276Z 	File: /.github/workflows/security.yml:99-105
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1596604Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1596800Z 	PASSED for resource: jobs(container-scan).steps[5](Run Anchore Grype scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1596910Z 	File: /.github/workflows/security.yml:104-112
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597032Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597230Z 	PASSED for resource: jobs(container-scan).steps[6](Upload Anchore scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597344Z 	File: /.github/workflows/security.yml:111-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597464Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597618Z 	PASSED for resource: jobs(license-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1597914Z 	File: /.github/workflows/security.yml:121-125
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598041Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598185Z 	PASSED for resource: jobs(license-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598290Z 	File: /.github/workflows/security.yml:124-131
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598417Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598585Z 	PASSED for resource: jobs(license-scan).steps[3](Install go-licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598688Z 	File: /.github/workflows/security.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1598809Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1602369Z 	PASSED for resource: jobs(license-scan).steps[4](Check licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1602558Z 	File: /.github/workflows/security.yml:133-139
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1602695Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1602876Z 	PASSED for resource: jobs(license-scan).steps[5](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1602975Z 	File: /.github/workflows/security.yml:138-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603105Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603254Z 	PASSED for resource: jobs(secrets-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603357Z 	File: /.github/workflows/security.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603482Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603634Z 	PASSED for resource: jobs(secrets-scan).steps[2](Run TruffleHog)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1603790Z 	File: /.github/workflows/security.yml:154-163
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604037Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604299Z 	PASSED for resource: jobs(secrets-scan).steps[3](Run GitLeaks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604450Z 	File: /.github/workflows/security.yml:162-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604586Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604728Z 	PASSED for resource: jobs(iac-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604828Z 	File: /.github/workflows/security.yml:172-176
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1604955Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605090Z 	PASSED for resource: jobs(iac-scan).steps[2](Run Checkov)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605186Z 	File: /.github/workflows/security.yml:175-184
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605317Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605480Z 	PASSED for resource: jobs(iac-scan).steps[3](Upload Checkov results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605577Z 	File: /.github/workflows/security.yml:183-189
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605702Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605838Z 	PASSED for resource: jobs(iac-scan).steps[4](Run Terrascan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1605941Z 	File: /.github/workflows/security.yml:188-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1606066Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1606253Z 	PASSED for resource: jobs(security-summary).steps[1](Security scan summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1606349Z 	File: /.github/workflows/security.yml:204-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1606604Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1606776Z 	PASSED for resource: jobs(dependency-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1607082Z 	File: /.github/workflows/security.yml:22-26
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1607328Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1607478Z 	PASSED for resource: jobs(dependency-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1607573Z 	File: /.github/workflows/security.yml:25-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1608187Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1608489Z 	PASSED for resource: jobs(dependency-scan).steps[3](Run Nancy (Go dependency scanner))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1608591Z 	File: /.github/workflows/security.yml:31-37
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1608841Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609013Z 	PASSED for resource: jobs(dependency-scan).steps[4](Run Govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609107Z 	File: /.github/workflows/security.yml:36-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609410Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609631Z 	PASSED for resource: jobs(sast-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609728Z 	File: /.github/workflows/security.yml:46-50
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1609965Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1610100Z 	PASSED for resource: jobs(sast-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1610348Z 	File: /.github/workflows/security.yml:49-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1610588Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1610755Z 	PASSED for resource: jobs(sast-scan).steps[3](Install and run Gosec)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1610847Z 	File: /.github/workflows/security.yml:55-65
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611079Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611243Z 	PASSED for resource: jobs(sast-scan).steps[4](Upload Gosec results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611343Z 	File: /.github/workflows/security.yml:64-70
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611576Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611713Z 	PASSED for resource: jobs(sast-scan).steps[5](Run Semgrep)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1611805Z 	File: /.github/workflows/security.yml:69-76
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612042Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612212Z 	PASSED for resource: jobs(sast-scan).steps[6](Upload Semgrep results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612303Z 	File: /.github/workflows/security.yml:75-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612537Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612695Z 	PASSED for resource: jobs(container-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1612788Z 	File: /.github/workflows/security.yml:85-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613021Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613199Z 	PASSED for resource: jobs(container-scan).steps[2](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613292Z 	File: /.github/workflows/security.yml:88-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613526Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613696Z 	PASSED for resource: jobs(container-scan).steps[3](Run Trivy scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1613796Z 	File: /.github/workflows/security.yml:92-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614030Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614222Z 	PASSED for resource: jobs(container-scan).steps[4](Upload Trivy scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614319Z 	File: /.github/workflows/security.yml:99-105
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614553Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614857Z 	PASSED for resource: jobs(container-scan).steps[5](Run Anchore Grype scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1614959Z 	File: /.github/workflows/security.yml:104-112
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615196Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615392Z 	PASSED for resource: jobs(container-scan).steps[6](Upload Anchore scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615496Z 	File: /.github/workflows/security.yml:111-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615729Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615880Z 	PASSED for resource: jobs(license-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1615978Z 	File: /.github/workflows/security.yml:121-125
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616211Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616352Z 	PASSED for resource: jobs(license-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616448Z 	File: /.github/workflows/security.yml:124-131
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616688Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616856Z 	PASSED for resource: jobs(license-scan).steps[3](Install go-licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1616954Z 	File: /.github/workflows/security.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1617188Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1617424Z 	PASSED for resource: jobs(license-scan).steps[4](Check licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1617520Z 	File: /.github/workflows/security.yml:133-139
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618023Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618220Z 	PASSED for resource: jobs(license-scan).steps[5](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618338Z 	File: /.github/workflows/security.yml:138-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618575Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618736Z 	PASSED for resource: jobs(secrets-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1618845Z 	File: /.github/workflows/security.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619074Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619226Z 	PASSED for resource: jobs(secrets-scan).steps[2](Run TruffleHog)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619333Z 	File: /.github/workflows/security.yml:154-163
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619563Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619712Z 	PASSED for resource: jobs(secrets-scan).steps[3](Run GitLeaks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1619815Z 	File: /.github/workflows/security.yml:162-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620042Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620180Z 	PASSED for resource: jobs(iac-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620291Z 	File: /.github/workflows/security.yml:172-176
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620519Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620653Z 	PASSED for resource: jobs(iac-scan).steps[2](Run Checkov)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620755Z 	File: /.github/workflows/security.yml:175-184
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1620988Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1621161Z 	PASSED for resource: jobs(iac-scan).steps[3](Upload Checkov results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1621266Z 	File: /.github/workflows/security.yml:183-189
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1621494Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1621633Z 	PASSED for resource: jobs(iac-scan).steps[4](Run Terrascan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1621735Z 	File: /.github/workflows/security.yml:188-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1687016Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1687928Z 	PASSED for resource: jobs(security-summary).steps[1](Security scan summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1688602Z 	File: /.github/workflows/security.yml:204-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1689064Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1689540Z 	PASSED for resource: jobs(dependency-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1689932Z 	File: /.github/workflows/security.yml:22-26
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1690263Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1690649Z 	PASSED for resource: jobs(dependency-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1690979Z 	File: /.github/workflows/security.yml:25-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1691302Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1691739Z 	PASSED for resource: jobs(dependency-scan).steps[3](Run Nancy (Go dependency scanner))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1692425Z 	File: /.github/workflows/security.yml:31-37
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1692763Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1693385Z 	PASSED for resource: jobs(dependency-scan).steps[4](Run Govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1693807Z 	File: /.github/workflows/security.yml:36-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1694346Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1694715Z 	PASSED for resource: jobs(sast-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1695219Z 	File: /.github/workflows/security.yml:46-50
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1695526Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1695872Z 	PASSED for resource: jobs(sast-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1696357Z 	File: /.github/workflows/security.yml:49-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1696658Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1697039Z 	PASSED for resource: jobs(sast-scan).steps[3](Install and run Gosec)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1697373Z 	File: /.github/workflows/security.yml:55-65
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1697876Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1698324Z 	PASSED for resource: jobs(sast-scan).steps[4](Upload Gosec results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1698660Z 	File: /.github/workflows/security.yml:64-70
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1698967Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1699310Z 	PASSED for resource: jobs(sast-scan).steps[5](Run Semgrep)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1699629Z 	File: /.github/workflows/security.yml:69-76
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1699935Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1700321Z 	PASSED for resource: jobs(sast-scan).steps[6](Upload Semgrep results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1700677Z 	File: /.github/workflows/security.yml:75-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1700981Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1701354Z 	PASSED for resource: jobs(container-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1701684Z 	File: /.github/workflows/security.yml:85-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1701986Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1702372Z 	PASSED for resource: jobs(container-scan).steps[2](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1702735Z 	File: /.github/workflows/security.yml:88-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1703053Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1703438Z 	PASSED for resource: jobs(container-scan).steps[3](Run Trivy scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1703796Z 	File: /.github/workflows/security.yml:92-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1704150Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1704558Z 	PASSED for resource: jobs(container-scan).steps[4](Upload Trivy scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1704927Z 	File: /.github/workflows/security.yml:99-105
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1705234Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1705626Z 	PASSED for resource: jobs(container-scan).steps[5](Run Anchore Grype scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1706163Z 	File: /.github/workflows/security.yml:104-112
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1706471Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1706869Z 	PASSED for resource: jobs(container-scan).steps[6](Upload Anchore scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1707247Z 	File: /.github/workflows/security.yml:111-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1707546Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1708088Z 	PASSED for resource: jobs(license-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1708420Z 	File: /.github/workflows/security.yml:121-125
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1708722Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1709063Z 	PASSED for resource: jobs(license-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1709371Z 	File: /.github/workflows/security.yml:124-131
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1709671Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1710042Z 	PASSED for resource: jobs(license-scan).steps[3](Install go-licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1710393Z 	File: /.github/workflows/security.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1710696Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1711051Z 	PASSED for resource: jobs(license-scan).steps[4](Check licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1711386Z 	File: /.github/workflows/security.yml:133-139
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1711701Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1712227Z 	PASSED for resource: jobs(license-scan).steps[5](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1712588Z 	File: /.github/workflows/security.yml:138-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1712899Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1713262Z 	PASSED for resource: jobs(secrets-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1713588Z 	File: /.github/workflows/security.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1713899Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1714266Z 	PASSED for resource: jobs(secrets-scan).steps[2](Run TruffleHog)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1714593Z 	File: /.github/workflows/security.yml:154-163
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1714902Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1715254Z 	PASSED for resource: jobs(secrets-scan).steps[3](Run GitLeaks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1715582Z 	File: /.github/workflows/security.yml:162-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1715884Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1716227Z 	PASSED for resource: jobs(iac-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1716540Z 	File: /.github/workflows/security.yml:172-176
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1716841Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1717180Z 	PASSED for resource: jobs(iac-scan).steps[2](Run Checkov)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1717484Z 	File: /.github/workflows/security.yml:175-184
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1717955Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1718335Z 	PASSED for resource: jobs(iac-scan).steps[3](Upload Checkov results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1718680Z 	File: /.github/workflows/security.yml:183-189
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1718986Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1719322Z 	PASSED for resource: jobs(iac-scan).steps[4](Run Terrascan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1719631Z 	File: /.github/workflows/security.yml:188-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1719937Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1720329Z 	PASSED for resource: jobs(security-summary).steps[1](Security scan summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1720695Z 	File: /.github/workflows/security.yml:204-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1721053Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1721471Z 	PASSED for resource: jobs(dependency-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1721800Z 	File: /.github/workflows/security.yml:22-26
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1722262Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1722807Z 	PASSED for resource: jobs(dependency-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1723123Z 	File: /.github/workflows/security.yml:25-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1723466Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1723934Z 	PASSED for resource: jobs(dependency-scan).steps[3](Run Nancy (Go dependency scanner))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1724327Z 	File: /.github/workflows/security.yml:31-37
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1724660Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1725079Z 	PASSED for resource: jobs(dependency-scan).steps[4](Run Govulncheck)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1725413Z 	File: /.github/workflows/security.yml:36-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1725754Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1726151Z 	PASSED for resource: jobs(sast-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1726465Z 	File: /.github/workflows/security.yml:46-50
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1726811Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1727195Z 	PASSED for resource: jobs(sast-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1727499Z 	File: /.github/workflows/security.yml:49-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1728054Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1728472Z 	PASSED for resource: jobs(sast-scan).steps[3](Install and run Gosec)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1729030Z 	File: /.github/workflows/security.yml:55-65
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1729379Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1729797Z 	PASSED for resource: jobs(sast-scan).steps[4](Upload Gosec results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1730133Z 	File: /.github/workflows/security.yml:64-70
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1730472Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1730862Z 	PASSED for resource: jobs(sast-scan).steps[5](Run Semgrep)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1731171Z 	File: /.github/workflows/security.yml:69-76
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1731509Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1731936Z 	PASSED for resource: jobs(sast-scan).steps[6](Upload Semgrep results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1732271Z 	File: /.github/workflows/security.yml:75-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1732614Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1733029Z 	PASSED for resource: jobs(container-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1733357Z 	File: /.github/workflows/security.yml:85-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1733697Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1734119Z 	PASSED for resource: jobs(container-scan).steps[2](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1734468Z 	File: /.github/workflows/security.yml:88-93
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1734807Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1735230Z 	PASSED for resource: jobs(container-scan).steps[3](Run Trivy scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1735640Z 	File: /.github/workflows/security.yml:92-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1736046Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1736592Z 	PASSED for resource: jobs(container-scan).steps[4](Upload Trivy scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1736967Z 	File: /.github/workflows/security.yml:99-105
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1737322Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1738059Z 	PASSED for resource: jobs(container-scan).steps[5](Run Anchore Grype scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1738615Z 	File: /.github/workflows/security.yml:104-112
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1738983Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1739618Z 	PASSED for resource: jobs(container-scan).steps[6](Upload Anchore scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1740286Z 	File: /.github/workflows/security.yml:111-118
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1741133Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1741684Z 	PASSED for resource: jobs(license-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1742230Z 	File: /.github/workflows/security.yml:121-125
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1742883Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1743605Z 	PASSED for resource: jobs(license-scan).steps[2](Set up Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1744199Z 	File: /.github/workflows/security.yml:124-131
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1744809Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1745433Z 	PASSED for resource: jobs(license-scan).steps[3](Install go-licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1745819Z 	File: /.github/workflows/security.yml:130-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1746454Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1747140Z 	PASSED for resource: jobs(license-scan).steps[4](Check licenses)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1748044Z 	File: /.github/workflows/security.yml:133-139
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1748663Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1749234Z 	PASSED for resource: jobs(license-scan).steps[5](Upload license report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1749822Z 	File: /.github/workflows/security.yml:138-146
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1750534Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1751079Z 	PASSED for resource: jobs(secrets-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1751564Z 	File: /.github/workflows/security.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1752180Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1752906Z 	PASSED for resource: jobs(secrets-scan).steps[2](Run TruffleHog)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1753482Z 	File: /.github/workflows/security.yml:154-163
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1754056Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1754715Z 	PASSED for resource: jobs(secrets-scan).steps[3](Run GitLeaks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1755233Z 	File: /.github/workflows/security.yml:162-169
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1755792Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1756429Z 	PASSED for resource: jobs(iac-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1756943Z 	File: /.github/workflows/security.yml:172-176
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1757517Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1758325Z 	PASSED for resource: jobs(iac-scan).steps[2](Run Checkov)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1758838Z 	File: /.github/workflows/security.yml:175-184
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1759408Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1760095Z 	PASSED for resource: jobs(iac-scan).steps[3](Upload Checkov results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1760651Z 	File: /.github/workflows/security.yml:183-189
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1761219Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1761863Z 	PASSED for resource: jobs(iac-scan).steps[4](Run Terrascan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1762366Z 	File: /.github/workflows/security.yml:188-199
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1762933Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1763652Z 	PASSED for resource: jobs(security-summary).steps[1](Security scan summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1764278Z 	File: /.github/workflows/security.yml:204-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1765469Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1766656Z 	PASSED for resource: on(Security-Hardened CI)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1767106Z 	File: /.github/workflows/ci-secure.yml:8-25
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1767611Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1768444Z 	PASSED for resource: jobs(security-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1768891Z 	File: /.github/workflows/ci-secure.yml:55-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1769402Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1769939Z 	PASSED for resource: jobs(secure-quick-checks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1770439Z 	File: /.github/workflows/ci-secure.yml:143-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1770968Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1771471Z 	PASSED for resource: jobs(secure-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1771923Z 	File: /.github/workflows/ci-secure.yml:214-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1772438Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1772939Z 	PASSED for resource: jobs(secure-docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1773394Z 	File: /.github/workflows/ci-secure.yml:300-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1773883Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1774359Z 	PASSED for resource: jobs(security-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1774808Z 	File: /.github/workflows/ci-secure.yml:389-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1775309Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1775805Z 	PASSED for resource: jobs(secure-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1776280Z 	File: /.github/workflows/ci-secure.yml:451-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1776815Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1777537Z 	PASSED for resource: jobs(security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1778100Z 	File: /.github/workflows/ci-secure.yml:475-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1778543Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1778965Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1779206Z 	File: /.github/workflows/ci-secure.yml:54-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1779636Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1780053Z 	PASSED for resource: jobs(security-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1780335Z 	File: /.github/workflows/ci-secure.yml:55-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1780746Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1781175Z 	PASSED for resource: jobs(secure-quick-checks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1781462Z 	File: /.github/workflows/ci-secure.yml:143-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1781869Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1782279Z 	PASSED for resource: jobs(secure-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1782543Z 	File: /.github/workflows/ci-secure.yml:214-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1782941Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1783360Z 	PASSED for resource: jobs(secure-docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1783647Z 	File: /.github/workflows/ci-secure.yml:300-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1784060Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1784479Z 	PASSED for resource: jobs(security-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1784743Z 	File: /.github/workflows/ci-secure.yml:389-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1785148Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1785544Z 	PASSED for resource: jobs(secure-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1785803Z 	File: /.github/workflows/ci-secure.yml:451-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1786213Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1786625Z 	PASSED for resource: jobs(security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1786912Z 	File: /.github/workflows/ci-secure.yml:475-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1787229Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1787541Z 	PASSED for resource: jobs(security-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1787972Z 	File: /.github/workflows/ci-secure.yml:55-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1788284Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1788756Z 	PASSED for resource: jobs(secure-quick-checks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1789033Z 	File: /.github/workflows/ci-secure.yml:143-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1789344Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1789644Z 	PASSED for resource: jobs(secure-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1789899Z 	File: /.github/workflows/ci-secure.yml:214-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1790212Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1790536Z 	PASSED for resource: jobs(secure-docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1790818Z 	File: /.github/workflows/ci-secure.yml:300-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1791119Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1791425Z 	PASSED for resource: jobs(security-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1791686Z 	File: /.github/workflows/ci-secure.yml:389-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1791987Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1792289Z 	PASSED for resource: jobs(secure-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1792550Z 	File: /.github/workflows/ci-secure.yml:451-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1792855Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1793169Z 	PASSED for resource: jobs(security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1793440Z 	File: /.github/workflows/ci-secure.yml:475-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1793793Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1794265Z 	PASSED for resource: jobs(security-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1794529Z 	File: /.github/workflows/ci-secure.yml:55-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1794879Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1795236Z 	PASSED for resource: jobs(secure-quick-checks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1795511Z 	File: /.github/workflows/ci-secure.yml:143-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1795858Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1796205Z 	PASSED for resource: jobs(secure-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1796457Z 	File: /.github/workflows/ci-secure.yml:214-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1796812Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1797173Z 	PASSED for resource: jobs(secure-docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1797446Z 	File: /.github/workflows/ci-secure.yml:300-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1798114Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1798507Z 	PASSED for resource: jobs(security-scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1798786Z 	File: /.github/workflows/ci-secure.yml:389-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1799150Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1799500Z 	PASSED for resource: jobs(secure-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1799768Z 	File: /.github/workflows/ci-secure.yml:451-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1800122Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1800487Z 	PASSED for resource: jobs(security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1800775Z 	File: /.github/workflows/ci-secure.yml:475-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1801197Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1801611Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1801840Z 	File: /.github/workflows/ci-secure.yml:54-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1802149Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1802562Z 	PASSED for resource: jobs(security-init).steps[1](Harden GitHub Actions Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1802944Z 	File: /.github/workflows/ci-secure.yml:63-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1803242Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1803665Z 	PASSED for resource: jobs(security-init).steps[2](Checkout code with security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1804106Z 	File: /.github/workflows/ci-secure.yml:81-88
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1804404Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1804809Z 	PASSED for resource: jobs(security-init).steps[3](Repository security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1805343Z 	File: /.github/workflows/ci-secure.yml:87-104
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1805639Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1806021Z 	PASSED for resource: jobs(security-init).steps[4](Detect changes securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1806390Z 	File: /.github/workflows/ci-secure.yml:103-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1806692Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1807086Z 	PASSED for resource: jobs(security-init).steps[5](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1807454Z 	File: /.github/workflows/ci-secure.yml:126-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1807925Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1808307Z 	PASSED for resource: jobs(secure-quick-checks).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1808665Z 	File: /.github/workflows/ci-secure.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1808965Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1809375Z 	PASSED for resource: jobs(secure-quick-checks).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1809797Z 	File: /.github/workflows/ci-secure.yml:154-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1810096Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1810505Z 	PASSED for resource: jobs(secure-quick-checks).steps[3](Secure code formatting check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1810906Z 	File: /.github/workflows/ci-secure.yml:163-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1811327Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1811733Z 	PASSED for resource: jobs(secure-quick-checks).steps[4](Secure Go mod verification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1812123Z 	File: /.github/workflows/ci-secure.yml:179-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1812422Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1812807Z 	PASSED for resource: jobs(secure-quick-checks).steps[5](Secure build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1813171Z 	File: /.github/workflows/ci-secure.yml:199-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1813469Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1813815Z 	PASSED for resource: jobs(secure-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1814146Z 	File: /.github/workflows/ci-secure.yml:237-243
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1814444Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1814826Z 	PASSED for resource: jobs(secure-test).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1815212Z 	File: /.github/workflows/ci-secure.yml:242-251
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1815506Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1815907Z 	PASSED for resource: jobs(secure-test).steps[3](Validate registry service securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1816298Z 	File: /.github/workflows/ci-secure.yml:250-275
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1816592Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1816943Z 	PASSED for resource: jobs(secure-test).steps[4](Run secure tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1817276Z 	File: /.github/workflows/ci-secure.yml:274-288
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1817578Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1818131Z 	PASSED for resource: jobs(secure-test).steps[5](Upload coverage securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1818490Z 	File: /.github/workflows/ci-secure.yml:287-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1818798Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1819167Z 	PASSED for resource: jobs(secure-docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1819525Z 	File: /.github/workflows/ci-secure.yml:307-314
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1819828Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1820253Z 	PASSED for resource: jobs(secure-docker-build).steps[2](Setup secure Docker environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1820665Z 	File: /.github/workflows/ci-secure.yml:313-330
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1820958Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1821354Z 	PASSED for resource: jobs(secure-docker-build).steps[3](Build secure container)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1821866Z 	File: /.github/workflows/ci-secure.yml:329-355
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1822160Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1822556Z 	PASSED for resource: jobs(secure-docker-build).steps[4](Container security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1822938Z 	File: /.github/workflows/ci-secure.yml:354-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1823234Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1823651Z 	PASSED for resource: jobs(secure-docker-build).steps[5](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1824056Z 	File: /.github/workflows/ci-secure.yml:365-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1824361Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1824767Z 	PASSED for resource: jobs(secure-docker-build).steps[6](Secure container smoke test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1825166Z 	File: /.github/workflows/ci-secure.yml:374-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1825459Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1825807Z 	PASSED for resource: jobs(security-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1826147Z 	File: /.github/workflows/ci-secure.yml:396-402
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1826442Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1826810Z 	PASSED for resource: jobs(security-scan).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1827159Z 	File: /.github/workflows/ci-secure.yml:401-409
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1827457Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1828173Z 	PASSED for resource: jobs(security-scan).steps[3](Run Gosec security scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1828539Z 	File: /.github/workflows/ci-secure.yml:408-421
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1828840Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1829240Z 	PASSED for resource: jobs(security-scan).steps[4](Run dependency vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1829628Z 	File: /.github/workflows/ci-secure.yml:420-433
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1829935Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1830311Z 	PASSED for resource: jobs(security-scan).steps[5](Run secret scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1830671Z 	File: /.github/workflows/ci-secure.yml:432-442
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1830979Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1831382Z 	PASSED for resource: jobs(security-scan).steps[6](Upload security scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1831764Z 	File: /.github/workflows/ci-secure.yml:441-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1832065Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1832426Z 	PASSED for resource: jobs(secure-lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1832749Z 	File: /.github/workflows/ci-secure.yml:457-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1833052Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1833424Z 	PASSED for resource: jobs(secure-lint).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1833771Z 	File: /.github/workflows/ci-secure.yml:460-468
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1834070Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1834431Z 	PASSED for resource: jobs(secure-lint).steps[3](Run secure linting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1834769Z 	File: /.github/workflows/ci-secure.yml:467-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1835063Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1835436Z 	PASSED for resource: jobs(security-validation).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1835789Z 	File: /.github/workflows/ci-secure.yml:482-487
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1836086Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1836486Z 	PASSED for resource: jobs(security-validation).steps[2](Security status analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1836869Z 	File: /.github/workflows/ci-secure.yml:486-533
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1837163Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1837564Z 	PASSED for resource: jobs(security-validation).steps[3](Generate security report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1838130Z 	File: /.github/workflows/ci-secure.yml:532-575
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1838593Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1839001Z 	PASSED for resource: jobs(security-validation).steps[4](Security gate enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1839392Z 	File: /.github/workflows/ci-secure.yml:574-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1839799Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1840310Z 	PASSED for resource: jobs(security-init).steps[1](Harden GitHub Actions Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1840694Z 	File: /.github/workflows/ci-secure.yml:63-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1841102Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1841642Z 	PASSED for resource: jobs(security-init).steps[2](Checkout code with security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1842045Z 	File: /.github/workflows/ci-secure.yml:81-88
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1842451Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1842967Z 	PASSED for resource: jobs(security-init).steps[3](Repository security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1843348Z 	File: /.github/workflows/ci-secure.yml:87-104
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1843753Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1844244Z 	PASSED for resource: jobs(security-init).steps[4](Detect changes securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1844725Z 	File: /.github/workflows/ci-secure.yml:103-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1845132Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1845619Z 	PASSED for resource: jobs(security-init).steps[5](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1845983Z 	File: /.github/workflows/ci-secure.yml:126-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1846386Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1847218Z 	PASSED for resource: jobs(secure-quick-checks).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1848147Z 	File: /.github/workflows/ci-secure.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1848751Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1849398Z 	PASSED for resource: jobs(secure-quick-checks).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1849811Z 	File: /.github/workflows/ci-secure.yml:154-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1850234Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1851354Z 	PASSED for resource: jobs(secure-quick-checks).steps[3](Secure code formatting check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1851758Z 	File: /.github/workflows/ci-secure.yml:163-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1852172Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1852693Z 	PASSED for resource: jobs(secure-quick-checks).steps[4](Secure Go mod verification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1853127Z 	File: /.github/workflows/ci-secure.yml:179-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1853549Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1854043Z 	PASSED for resource: jobs(secure-quick-checks).steps[5](Secure build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1854416Z 	File: /.github/workflows/ci-secure.yml:199-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1854817Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1855282Z 	PASSED for resource: jobs(secure-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1855608Z 	File: /.github/workflows/ci-secure.yml:237-243
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1856013Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1856509Z 	PASSED for resource: jobs(secure-test).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1856882Z 	File: /.github/workflows/ci-secure.yml:242-251
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1857278Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1858183Z 	PASSED for resource: jobs(secure-test).steps[3](Validate registry service securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1858573Z 	File: /.github/workflows/ci-secure.yml:250-275
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1858984Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1859457Z 	PASSED for resource: jobs(secure-test).steps[4](Run secure tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1859801Z 	File: /.github/workflows/ci-secure.yml:274-288
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1860205Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1860691Z 	PASSED for resource: jobs(secure-test).steps[5](Upload coverage securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1861055Z 	File: /.github/workflows/ci-secure.yml:287-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1861461Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1861943Z 	PASSED for resource: jobs(secure-docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1862303Z 	File: /.github/workflows/ci-secure.yml:307-314
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1862710Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1863255Z 	PASSED for resource: jobs(secure-docker-build).steps[2](Setup secure Docker environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1863669Z 	File: /.github/workflows/ci-secure.yml:313-330
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1864071Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1864716Z 	PASSED for resource: jobs(secure-docker-build).steps[3](Build secure container)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1865099Z 	File: /.github/workflows/ci-secure.yml:329-355
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1865506Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1866019Z 	PASSED for resource: jobs(secure-docker-build).steps[4](Container security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1866399Z 	File: /.github/workflows/ci-secure.yml:354-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1866809Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1867340Z 	PASSED for resource: jobs(secure-docker-build).steps[5](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1867948Z 	File: /.github/workflows/ci-secure.yml:365-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1868360Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1868886Z 	PASSED for resource: jobs(secure-docker-build).steps[6](Secure container smoke test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1869277Z 	File: /.github/workflows/ci-secure.yml:374-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1869683Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1870144Z 	PASSED for resource: jobs(security-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1870476Z 	File: /.github/workflows/ci-secure.yml:396-402
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1870875Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1871365Z 	PASSED for resource: jobs(security-scan).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1871719Z 	File: /.github/workflows/ci-secure.yml:401-409
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1872114Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1872606Z 	PASSED for resource: jobs(security-scan).steps[3](Run Gosec security scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1872968Z 	File: /.github/workflows/ci-secure.yml:408-421
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1873370Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1873894Z 	PASSED for resource: jobs(security-scan).steps[4](Run dependency vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1874288Z 	File: /.github/workflows/ci-secure.yml:420-433
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1874690Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1875162Z 	PASSED for resource: jobs(security-scan).steps[5](Run secret scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1875642Z 	File: /.github/workflows/ci-secure.yml:432-442
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1876051Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1876555Z 	PASSED for resource: jobs(security-scan).steps[6](Upload security scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1876926Z 	File: /.github/workflows/ci-secure.yml:441-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1877334Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1877987Z 	PASSED for resource: jobs(secure-lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1878315Z 	File: /.github/workflows/ci-secure.yml:457-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1878718Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1879196Z 	PASSED for resource: jobs(secure-lint).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1879542Z 	File: /.github/workflows/ci-secure.yml:460-468
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1879953Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1880424Z 	PASSED for resource: jobs(secure-lint).steps[3](Run secure linting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1880759Z 	File: /.github/workflows/ci-secure.yml:467-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1881162Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1881771Z 	PASSED for resource: jobs(security-validation).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1882114Z 	File: /.github/workflows/ci-secure.yml:482-487
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1882522Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1883030Z 	PASSED for resource: jobs(security-validation).steps[2](Security status analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1883412Z 	File: /.github/workflows/ci-secure.yml:486-533
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1883815Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1884329Z 	PASSED for resource: jobs(security-validation).steps[3](Generate security report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1884718Z 	File: /.github/workflows/ci-secure.yml:532-575
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1885114Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1885637Z 	PASSED for resource: jobs(security-validation).steps[4](Security gate enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1886030Z 	File: /.github/workflows/ci-secure.yml:574-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1886337Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1886749Z 	PASSED for resource: jobs(security-init).steps[1](Harden GitHub Actions Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1887123Z 	File: /.github/workflows/ci-secure.yml:63-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1887438Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1888029Z 	PASSED for resource: jobs(security-init).steps[2](Checkout code with security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1888431Z 	File: /.github/workflows/ci-secure.yml:81-88
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1888747Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1889155Z 	PASSED for resource: jobs(security-init).steps[3](Repository security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1889544Z 	File: /.github/workflows/ci-secure.yml:87-104
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1890010Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1890689Z 	PASSED for resource: jobs(security-init).steps[4](Detect changes securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1891125Z 	File: /.github/workflows/ci-secure.yml:103-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1891435Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1891824Z 	PASSED for resource: jobs(security-init).steps[5](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1892189Z 	File: /.github/workflows/ci-secure.yml:126-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1892491Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1892874Z 	PASSED for resource: jobs(secure-quick-checks).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1893367Z 	File: /.github/workflows/ci-secure.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1893676Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1894109Z 	PASSED for resource: jobs(secure-quick-checks).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1894512Z 	File: /.github/workflows/ci-secure.yml:154-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1894816Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1895242Z 	PASSED for resource: jobs(secure-quick-checks).steps[3](Secure code formatting check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1895639Z 	File: /.github/workflows/ci-secure.yml:163-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1895945Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1896359Z 	PASSED for resource: jobs(secure-quick-checks).steps[4](Secure Go mod verification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1896749Z 	File: /.github/workflows/ci-secure.yml:179-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1897047Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1897450Z 	PASSED for resource: jobs(secure-quick-checks).steps[5](Secure build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1897550Z 	File: /.github/workflows/ci-secure.yml:199-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1897862Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898019Z 	PASSED for resource: jobs(secure-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898118Z 	File: /.github/workflows/ci-secure.yml:237-243
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898385Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898576Z 	PASSED for resource: jobs(secure-test).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898674Z 	File: /.github/workflows/ci-secure.yml:242-251
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1898809Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899013Z 	PASSED for resource: jobs(secure-test).steps[3](Validate registry service securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899112Z 	File: /.github/workflows/ci-secure.yml:250-275
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899257Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899415Z 	PASSED for resource: jobs(secure-test).steps[4](Run secure tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899513Z 	File: /.github/workflows/ci-secure.yml:274-288
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899646Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899825Z 	PASSED for resource: jobs(secure-test).steps[5](Upload coverage securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1899928Z 	File: /.github/workflows/ci-secure.yml:287-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900054Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900229Z 	PASSED for resource: jobs(secure-docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900323Z 	File: /.github/workflows/ci-secure.yml:307-314
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900452Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900680Z 	PASSED for resource: jobs(secure-docker-build).steps[2](Setup secure Docker environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900780Z 	File: /.github/workflows/ci-secure.yml:313-330
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1900909Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901109Z 	PASSED for resource: jobs(secure-docker-build).steps[3](Build secure container)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901206Z 	File: /.github/workflows/ci-secure.yml:329-355
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901333Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901539Z 	PASSED for resource: jobs(secure-docker-build).steps[4](Container security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901635Z 	File: /.github/workflows/ci-secure.yml:354-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1901762Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902079Z 	PASSED for resource: jobs(secure-docker-build).steps[5](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902181Z 	File: /.github/workflows/ci-secure.yml:365-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902312Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902522Z 	PASSED for resource: jobs(secure-docker-build).steps[6](Secure container smoke test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902743Z 	File: /.github/workflows/ci-secure.yml:374-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1902873Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903030Z 	PASSED for resource: jobs(security-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903127Z 	File: /.github/workflows/ci-secure.yml:396-402
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903262Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903443Z 	PASSED for resource: jobs(security-scan).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903539Z 	File: /.github/workflows/ci-secure.yml:401-409
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903665Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903853Z 	PASSED for resource: jobs(security-scan).steps[3](Run Gosec security scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1903949Z 	File: /.github/workflows/ci-secure.yml:408-421
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904109Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904330Z 	PASSED for resource: jobs(security-scan).steps[4](Run dependency vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904434Z 	File: /.github/workflows/ci-secure.yml:420-433
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904561Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904729Z 	PASSED for resource: jobs(security-scan).steps[5](Run secret scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1904926Z 	File: /.github/workflows/ci-secure.yml:432-442
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905057Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905254Z 	PASSED for resource: jobs(security-scan).steps[6](Upload security scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905348Z 	File: /.github/workflows/ci-secure.yml:441-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905476Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905627Z 	PASSED for resource: jobs(secure-lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905722Z 	File: /.github/workflows/ci-secure.yml:457-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1905858Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906029Z 	PASSED for resource: jobs(secure-lint).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906123Z 	File: /.github/workflows/ci-secure.yml:460-468
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906251Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906415Z 	PASSED for resource: jobs(secure-lint).steps[3](Run secure linting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906516Z 	File: /.github/workflows/ci-secure.yml:467-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906643Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906815Z 	PASSED for resource: jobs(security-validation).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1906910Z 	File: /.github/workflows/ci-secure.yml:482-487
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907038Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907245Z 	PASSED for resource: jobs(security-validation).steps[2](Security status analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907344Z 	File: /.github/workflows/ci-secure.yml:486-533
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907476Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907828Z 	PASSED for resource: jobs(security-validation).steps[3](Generate security report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1907931Z 	File: /.github/workflows/ci-secure.yml:532-575
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908064Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908277Z 	PASSED for resource: jobs(security-validation).steps[4](Security gate enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908373Z 	File: /.github/workflows/ci-secure.yml:574-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908548Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908743Z 	PASSED for resource: jobs(security-init).steps[1](Harden GitHub Actions Runner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1908841Z 	File: /.github/workflows/ci-secure.yml:63-82
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909016Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909373Z 	PASSED for resource: jobs(security-init).steps[2](Checkout code with security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909469Z 	File: /.github/workflows/ci-secure.yml:81-88
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909644Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909848Z 	PASSED for resource: jobs(security-init).steps[3](Repository security validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1909955Z 	File: /.github/workflows/ci-secure.yml:87-104
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910130Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910318Z 	PASSED for resource: jobs(security-init).steps[4](Detect changes securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910418Z 	File: /.github/workflows/ci-secure.yml:103-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910590Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910777Z 	PASSED for resource: jobs(security-init).steps[5](Security configuration)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1910882Z 	File: /.github/workflows/ci-secure.yml:126-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911061Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911239Z 	PASSED for resource: jobs(secure-quick-checks).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911336Z 	File: /.github/workflows/ci-secure.yml:149-155
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911507Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911840Z 	PASSED for resource: jobs(secure-quick-checks).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1911937Z 	File: /.github/workflows/ci-secure.yml:154-164
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912112Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912334Z 	PASSED for resource: jobs(secure-quick-checks).steps[3](Secure code formatting check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912429Z 	File: /.github/workflows/ci-secure.yml:163-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912598Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912812Z 	PASSED for resource: jobs(secure-quick-checks).steps[4](Secure Go mod verification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1912909Z 	File: /.github/workflows/ci-secure.yml:179-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913081Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913271Z 	PASSED for resource: jobs(secure-quick-checks).steps[5](Secure build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913374Z 	File: /.github/workflows/ci-secure.yml:199-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913544Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913695Z 	PASSED for resource: jobs(secure-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913790Z 	File: /.github/workflows/ci-secure.yml:237-243
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1913962Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914157Z 	PASSED for resource: jobs(secure-test).steps[2](Setup Go environment securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914261Z 	File: /.github/workflows/ci-secure.yml:242-251
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914438Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914645Z 	PASSED for resource: jobs(secure-test).steps[3](Validate registry service securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914743Z 	File: /.github/workflows/ci-secure.yml:250-275
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1914914Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915079Z 	PASSED for resource: jobs(secure-test).steps[4](Run secure tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915175Z 	File: /.github/workflows/ci-secure.yml:274-288
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915346Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915527Z 	PASSED for resource: jobs(secure-test).steps[5](Upload coverage securely)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915622Z 	File: /.github/workflows/ci-secure.yml:287-300
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915794Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1915968Z 	PASSED for resource: jobs(secure-docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1916154Z 	File: /.github/workflows/ci-secure.yml:307-314
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1916327Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1916558Z 	PASSED for resource: jobs(secure-docker-build).steps[2](Setup secure Docker environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1916656Z 	File: /.github/workflows/ci-secure.yml:313-330
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1916842Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1917048Z 	PASSED for resource: jobs(secure-docker-build).steps[3](Build secure container)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1917146Z 	File: /.github/workflows/ci-secure.yml:329-355
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1917319Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1917526Z 	PASSED for resource: jobs(secure-docker-build).steps[4](Container security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1917800Z 	File: /.github/workflows/ci-secure.yml:354-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1918119Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1918505Z 	PASSED for resource: jobs(secure-docker-build).steps[5](Upload container scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1918627Z 	File: /.github/workflows/ci-secure.yml:365-375
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1918809Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919175Z 	PASSED for resource: jobs(secure-docker-build).steps[6](Secure container smoke test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919272Z 	File: /.github/workflows/ci-secure.yml:374-389
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919448Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919611Z 	PASSED for resource: jobs(security-scan).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919709Z 	File: /.github/workflows/ci-secure.yml:396-402
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1919879Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920063Z 	PASSED for resource: jobs(security-scan).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920160Z 	File: /.github/workflows/ci-secure.yml:401-409
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920334Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920526Z 	PASSED for resource: jobs(security-scan).steps[3](Run Gosec security scanner)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920626Z 	File: /.github/workflows/ci-secure.yml:408-421
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1920803Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921017Z 	PASSED for resource: jobs(security-scan).steps[4](Run dependency vulnerability scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921112Z 	File: /.github/workflows/ci-secure.yml:420-433
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921285Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921460Z 	PASSED for resource: jobs(security-scan).steps[5](Run secret scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921556Z 	File: /.github/workflows/ci-secure.yml:432-442
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921730Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1921929Z 	PASSED for resource: jobs(security-scan).steps[6](Upload security scan results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922025Z 	File: /.github/workflows/ci-secure.yml:441-451
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922197Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922354Z 	PASSED for resource: jobs(secure-lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922449Z 	File: /.github/workflows/ci-secure.yml:457-461
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922621Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922792Z 	PASSED for resource: jobs(secure-lint).steps[2](Setup Go environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1922889Z 	File: /.github/workflows/ci-secure.yml:460-468
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923062Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923226Z 	PASSED for resource: jobs(secure-lint).steps[3](Run secure linting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923475Z 	File: /.github/workflows/ci-secure.yml:467-475
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923648Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923823Z 	PASSED for resource: jobs(security-validation).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1923920Z 	File: /.github/workflows/ci-secure.yml:482-487
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924102Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924310Z 	PASSED for resource: jobs(security-validation).steps[2](Security status analysis)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924406Z 	File: /.github/workflows/ci-secure.yml:486-533
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924578Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924785Z 	PASSED for resource: jobs(security-validation).steps[3](Generate security report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1924881Z 	File: /.github/workflows/ci-secure.yml:532-575
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1925056Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1925268Z 	PASSED for resource: jobs(security-validation).steps[4](Security gate enforcement)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1925364Z 	File: /.github/workflows/ci-secure.yml:574-588
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1925908Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926087Z 	PASSED for resource: on(Test Matrix)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926190Z 	File: /.github/workflows/test-matrix.yml:4-20
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926318Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926417Z 	PASSED for resource: jobs(test-matrix)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926523Z 	File: /.github/workflows/test-matrix.yml:26-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926646Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926741Z 	PASSED for resource: jobs(benchmark)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926852Z 	File: /.github/workflows/test-matrix.yml:152-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1926973Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1927066Z 	PASSED for resource: jobs(load-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1927169Z 	File: /.github/workflows/test-matrix.yml:180-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1927408Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1927516Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1927774Z 	File: /.github/workflows/test-matrix.yml:25-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928017Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928118Z 	PASSED for resource: jobs(test-matrix)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928219Z 	File: /.github/workflows/test-matrix.yml:26-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928447Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928542Z 	PASSED for resource: jobs(benchmark)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928649Z 	File: /.github/workflows/test-matrix.yml:152-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928877Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1928970Z 	PASSED for resource: jobs(load-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929069Z 	File: /.github/workflows/test-matrix.yml:180-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929204Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929305Z 	PASSED for resource: jobs(test-matrix)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929407Z 	File: /.github/workflows/test-matrix.yml:26-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929540Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929638Z 	PASSED for resource: jobs(benchmark)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929740Z 	File: /.github/workflows/test-matrix.yml:152-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929870Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1929961Z 	PASSED for resource: jobs(load-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930062Z 	File: /.github/workflows/test-matrix.yml:180-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930372Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930466Z 	PASSED for resource: jobs(test-matrix)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930569Z 	File: /.github/workflows/test-matrix.yml:26-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930744Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930840Z 	PASSED for resource: jobs(benchmark)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1930941Z 	File: /.github/workflows/test-matrix.yml:152-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931113Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931205Z 	PASSED for resource: jobs(load-test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931309Z 	File: /.github/workflows/test-matrix.yml:180-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931555Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931642Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931742Z 	File: /.github/workflows/test-matrix.yml:25-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1931868Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932017Z 	PASSED for resource: jobs(test-matrix).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932127Z 	File: /.github/workflows/test-matrix.yml:59-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932247Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932382Z 	PASSED for resource: jobs(test-matrix).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932602Z 	File: /.github/workflows/test-matrix.yml:62-69
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932727Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1932904Z 	PASSED for resource: jobs(test-matrix).steps[3](Setup test environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933009Z 	File: /.github/workflows/test-matrix.yml:68-91
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933129Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933340Z 	PASSED for resource: jobs(test-matrix).steps[4](Wait for registry (Linux/macOS only))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933448Z 	File: /.github/workflows/test-matrix.yml:90-99
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933567Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933702Z 	PASSED for resource: jobs(test-matrix).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933810Z 	File: /.github/workflows/test-matrix.yml:98-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1933929Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934092Z 	PASSED for resource: jobs(test-matrix).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934199Z 	File: /.github/workflows/test-matrix.yml:126-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934320Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934465Z 	PASSED for resource: jobs(test-matrix).steps[7](Test summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934577Z 	File: /.github/workflows/test-matrix.yml:135-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934694Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934835Z 	PASSED for resource: jobs(benchmark).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1934943Z 	File: /.github/workflows/test-matrix.yml:157-161
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935061Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935188Z 	PASSED for resource: jobs(benchmark).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935292Z 	File: /.github/workflows/test-matrix.yml:160-167
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935408Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935556Z 	PASSED for resource: jobs(benchmark).steps[3](Run benchmarks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935660Z 	File: /.github/workflows/test-matrix.yml:166-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935780Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1935952Z 	PASSED for resource: jobs(benchmark).steps[4](Upload benchmark results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936057Z 	File: /.github/workflows/test-matrix.yml:171-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936173Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936312Z 	PASSED for resource: jobs(load-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936506Z 	File: /.github/workflows/test-matrix.yml:192-196
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936626Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936749Z 	PASSED for resource: jobs(load-test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936853Z 	File: /.github/workflows/test-matrix.yml:195-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1936973Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937129Z 	PASSED for resource: jobs(load-test).steps[3](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937235Z 	File: /.github/workflows/test-matrix.yml:201-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937357Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937495Z 	PASSED for resource: jobs(load-test).steps[4](Run load tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937594Z 	File: /.github/workflows/test-matrix.yml:205-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1937942Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938090Z 	PASSED for resource: jobs(test-matrix).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938195Z 	File: /.github/workflows/test-matrix.yml:59-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938428Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938561Z 	PASSED for resource: jobs(test-matrix).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938665Z 	File: /.github/workflows/test-matrix.yml:62-69
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1938889Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1939178Z 	PASSED for resource: jobs(test-matrix).steps[3](Setup test environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1939283Z 	File: /.github/workflows/test-matrix.yml:68-91
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1939514Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1939725Z 	PASSED for resource: jobs(test-matrix).steps[4](Wait for registry (Linux/macOS only))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1939826Z 	File: /.github/workflows/test-matrix.yml:90-99
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940059Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940194Z 	PASSED for resource: jobs(test-matrix).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940301Z 	File: /.github/workflows/test-matrix.yml:98-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940526Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940679Z 	PASSED for resource: jobs(test-matrix).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1940785Z 	File: /.github/workflows/test-matrix.yml:126-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941010Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941152Z 	PASSED for resource: jobs(test-matrix).steps[7](Test summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941258Z 	File: /.github/workflows/test-matrix.yml:135-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941482Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941625Z 	PASSED for resource: jobs(benchmark).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941734Z 	File: /.github/workflows/test-matrix.yml:157-161
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1941963Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942088Z 	PASSED for resource: jobs(benchmark).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942193Z 	File: /.github/workflows/test-matrix.yml:160-167
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942421Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942562Z 	PASSED for resource: jobs(benchmark).steps[3](Run benchmarks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942669Z 	File: /.github/workflows/test-matrix.yml:166-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1942893Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943062Z 	PASSED for resource: jobs(benchmark).steps[4](Upload benchmark results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943170Z 	File: /.github/workflows/test-matrix.yml:171-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943516Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943659Z 	PASSED for resource: jobs(load-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943764Z 	File: /.github/workflows/test-matrix.yml:192-196
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1943988Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944118Z 	PASSED for resource: jobs(load-test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944225Z 	File: /.github/workflows/test-matrix.yml:195-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944451Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944599Z 	PASSED for resource: jobs(load-test).steps[3](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944706Z 	File: /.github/workflows/test-matrix.yml:201-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1944929Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945074Z 	PASSED for resource: jobs(load-test).steps[4](Run load tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945180Z 	File: /.github/workflows/test-matrix.yml:205-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945316Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945459Z 	PASSED for resource: jobs(test-matrix).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945562Z 	File: /.github/workflows/test-matrix.yml:59-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945776Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1945907Z 	PASSED for resource: jobs(test-matrix).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946010Z 	File: /.github/workflows/test-matrix.yml:62-69
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946138Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946309Z 	PASSED for resource: jobs(test-matrix).steps[3](Setup test environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946413Z 	File: /.github/workflows/test-matrix.yml:68-91
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946544Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946757Z 	PASSED for resource: jobs(test-matrix).steps[4](Wait for registry (Linux/macOS only))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946865Z 	File: /.github/workflows/test-matrix.yml:90-99
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1946995Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947129Z 	PASSED for resource: jobs(test-matrix).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947237Z 	File: /.github/workflows/test-matrix.yml:98-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947370Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947519Z 	PASSED for resource: jobs(test-matrix).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947730Z 	File: /.github/workflows/test-matrix.yml:126-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1947866Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948007Z 	PASSED for resource: jobs(test-matrix).steps[7](Test summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948116Z 	File: /.github/workflows/test-matrix.yml:135-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948244Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948388Z 	PASSED for resource: jobs(benchmark).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948488Z 	File: /.github/workflows/test-matrix.yml:157-161
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948620Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948771Z 	PASSED for resource: jobs(benchmark).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1948882Z 	File: /.github/workflows/test-matrix.yml:160-167
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949020Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949162Z 	PASSED for resource: jobs(benchmark).steps[3](Run benchmarks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949264Z 	File: /.github/workflows/test-matrix.yml:166-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949401Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949572Z 	PASSED for resource: jobs(benchmark).steps[4](Upload benchmark results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949673Z 	File: /.github/workflows/test-matrix.yml:171-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1949938Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950078Z 	PASSED for resource: jobs(load-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950181Z 	File: /.github/workflows/test-matrix.yml:192-196
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950315Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950442Z 	PASSED for resource: jobs(load-test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950547Z 	File: /.github/workflows/test-matrix.yml:195-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950682Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950829Z 	PASSED for resource: jobs(load-test).steps[3](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1950928Z 	File: /.github/workflows/test-matrix.yml:201-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951063Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951199Z 	PASSED for resource: jobs(load-test).steps[4](Run load tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951299Z 	File: /.github/workflows/test-matrix.yml:205-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951485Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951630Z 	PASSED for resource: jobs(test-matrix).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951729Z 	File: /.github/workflows/test-matrix.yml:59-63
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1951911Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952044Z 	PASSED for resource: jobs(test-matrix).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952255Z 	File: /.github/workflows/test-matrix.yml:62-69
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952435Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952604Z 	PASSED for resource: jobs(test-matrix).steps[3](Setup test environment)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952703Z 	File: /.github/workflows/test-matrix.yml:68-91
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1952881Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953091Z 	PASSED for resource: jobs(test-matrix).steps[4](Wait for registry (Linux/macOS only))
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953196Z 	File: /.github/workflows/test-matrix.yml:90-99
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953375Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953508Z 	PASSED for resource: jobs(test-matrix).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953612Z 	File: /.github/workflows/test-matrix.yml:98-127
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953790Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1953943Z 	PASSED for resource: jobs(test-matrix).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954045Z 	File: /.github/workflows/test-matrix.yml:126-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954222Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954361Z 	PASSED for resource: jobs(test-matrix).steps[7](Test summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954463Z 	File: /.github/workflows/test-matrix.yml:135-152
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954641Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954783Z 	PASSED for resource: jobs(benchmark).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1954885Z 	File: /.github/workflows/test-matrix.yml:157-161
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955063Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955186Z 	PASSED for resource: jobs(benchmark).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955286Z 	File: /.github/workflows/test-matrix.yml:160-167
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955467Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955606Z 	PASSED for resource: jobs(benchmark).steps[3](Run benchmarks)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955708Z 	File: /.github/workflows/test-matrix.yml:166-172
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1955887Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956055Z 	PASSED for resource: jobs(benchmark).steps[4](Upload benchmark results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956155Z 	File: /.github/workflows/test-matrix.yml:171-180
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956423Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956560Z 	PASSED for resource: jobs(load-test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956658Z 	File: /.github/workflows/test-matrix.yml:192-196
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956835Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1956963Z 	PASSED for resource: jobs(load-test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957064Z 	File: /.github/workflows/test-matrix.yml:195-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957241Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957389Z 	PASSED for resource: jobs(load-test).steps[3](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957491Z 	File: /.github/workflows/test-matrix.yml:201-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957772Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1957912Z 	PASSED for resource: jobs(load-test).steps[4](Run load tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958017Z 	File: /.github/workflows/test-matrix.yml:205-215
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958149Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958253Z 	PASSED for resource: jobs(monitoring-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958414Z 	File: /.github/workflows/security-monitoring-enhanced.yml:52-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958542Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958798Z 	PASSED for resource: jobs(continuous-secret-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1958962Z 	File: /.github/workflows/security-monitoring-enhanced.yml:120-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959090Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959247Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959407Z 	File: /.github/workflows/security-monitoring-enhanced.yml:191-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959530Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959671Z 	PASSED for resource: jobs(container-security-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959831Z 	File: /.github/workflows/security-monitoring-enhanced.yml:286-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1959962Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960093Z 	PASSED for resource: jobs(security-baseline-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960253Z 	File: /.github/workflows/security-monitoring-enhanced.yml:415-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960385Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960492Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960647Z 	File: /.github/workflows/security-monitoring-enhanced.yml:514-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960891Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1960973Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961124Z 	File: /.github/workflows/security-monitoring-enhanced.yml:51-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961359Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961463Z 	PASSED for resource: jobs(monitoring-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961620Z 	File: /.github/workflows/security-monitoring-enhanced.yml:52-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961856Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1961989Z 	PASSED for resource: jobs(continuous-secret-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1962148Z 	File: /.github/workflows/security-monitoring-enhanced.yml:120-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1962377Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1962534Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1962689Z 	File: /.github/workflows/security-monitoring-enhanced.yml:191-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1962919Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1963054Z 	PASSED for resource: jobs(container-security-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1963326Z 	File: /.github/workflows/security-monitoring-enhanced.yml:286-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1963563Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1963696Z 	PASSED for resource: jobs(security-baseline-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1963850Z 	File: /.github/workflows/security-monitoring-enhanced.yml:415-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964083Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964190Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964346Z 	File: /.github/workflows/security-monitoring-enhanced.yml:514-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964483Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964582Z 	PASSED for resource: jobs(monitoring-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964736Z 	File: /.github/workflows/security-monitoring-enhanced.yml:52-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1964876Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965014Z 	PASSED for resource: jobs(continuous-secret-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965168Z 	File: /.github/workflows/security-monitoring-enhanced.yml:120-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965302Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965458Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965610Z 	File: /.github/workflows/security-monitoring-enhanced.yml:191-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965828Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1965966Z 	PASSED for resource: jobs(container-security-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966120Z 	File: /.github/workflows/security-monitoring-enhanced.yml:286-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966253Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966384Z 	PASSED for resource: jobs(security-baseline-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966537Z 	File: /.github/workflows/security-monitoring-enhanced.yml:415-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966675Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966778Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1966934Z 	File: /.github/workflows/security-monitoring-enhanced.yml:514-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967116Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967214Z 	PASSED for resource: jobs(monitoring-init)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967369Z 	File: /.github/workflows/security-monitoring-enhanced.yml:52-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967547Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967783Z 	PASSED for resource: jobs(continuous-secret-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1967939Z 	File: /.github/workflows/security-monitoring-enhanced.yml:120-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968121Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968277Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968437Z 	File: /.github/workflows/security-monitoring-enhanced.yml:191-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968620Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968759Z 	PASSED for resource: jobs(container-security-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1968913Z 	File: /.github/workflows/security-monitoring-enhanced.yml:286-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969094Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969226Z 	PASSED for resource: jobs(security-baseline-monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969381Z 	File: /.github/workflows/security-monitoring-enhanced.yml:415-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969563Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969665Z 	PASSED for resource: jobs(security-alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1969824Z 	File: /.github/workflows/security-monitoring-enhanced.yml:514-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970076Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970281Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970437Z 	File: /.github/workflows/security-monitoring-enhanced.yml:51-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970567Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970782Z 	PASSED for resource: jobs(monitoring-init).steps[1](Initialize security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1970941Z 	File: /.github/workflows/security-monitoring-enhanced.yml:61-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1971065Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1971332Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[1](Checkout full repository history)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1971488Z 	File: /.github/workflows/security-monitoring-enhanced.yml:127-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1971612Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1971886Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[2](TruffleHog historical secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1972046Z 	File: /.github/workflows/security-monitoring-enhanced.yml:133-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1972172Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1972422Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[3](GitLeaks comprehensive scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1972575Z 	File: /.github/workflows/security-monitoring-enhanced.yml:142-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1972811Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973048Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[4](Analyze secret findings)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973204Z 	File: /.github/workflows/security-monitoring-enhanced.yml:157-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973328Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973563Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973719Z 	File: /.github/workflows/security-monitoring-enhanced.yml:198-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1973848Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974062Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974218Z 	File: /.github/workflows/security-monitoring-enhanced.yml:201-209
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974340Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974637Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[3](Enhanced vulnerability scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974798Z 	File: /.github/workflows/security-monitoring-enhanced.yml:208-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1974920Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1975210Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[4](License compliance monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1975365Z 	File: /.github/workflows/security-monitoring-enhanced.yml:247-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1975489Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1975693Z 	PASSED for resource: jobs(container-security-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1975860Z 	File: /.github/workflows/security-monitoring-enhanced.yml:293-298
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1982370Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1982765Z 	PASSED for resource: jobs(container-security-monitoring).steps[2](Build containers for security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1982961Z 	File: /.github/workflows/security-monitoring-enhanced.yml:297-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1983124Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1983422Z 	PASSED for resource: jobs(container-security-monitoring).steps[3](Comprehensive Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1983600Z 	File: /.github/workflows/security-monitoring-enhanced.yml:320-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1983733Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984038Z 	PASSED for resource: jobs(container-security-monitoring).steps[4](Docker security best practices audit)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984375Z 	File: /.github/workflows/security-monitoring-enhanced.yml:367-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984501Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984711Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984871Z 	File: /.github/workflows/security-monitoring-enhanced.yml:422-427
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1984990Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1985251Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[2](Generate security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1985406Z 	File: /.github/workflows/security-monitoring-enhanced.yml:426-483
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1985523Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1985758Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[3](Store security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1985911Z 	File: /.github/workflows/security-monitoring-enhanced.yml:482-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986035Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986248Z 	PASSED for resource: jobs(security-alerting).steps[1](Determine alert severity)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986409Z 	File: /.github/workflows/security-monitoring-enhanced.yml:521-562
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986528Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986721Z 	PASSED for resource: jobs(security-alerting).steps[2](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1986993Z 	File: /.github/workflows/security-monitoring-enhanced.yml:561-611
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1987114Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1987313Z 	PASSED for resource: jobs(security-alerting).steps[3](Send Slack notification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1987466Z 	File: /.github/workflows/security-monitoring-enhanced.yml:610-652
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1987584Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1987932Z 	PASSED for resource: jobs(security-alerting).steps[4](Generate monitoring report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1988098Z 	File: /.github/workflows/security-monitoring-enhanced.yml:651-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1988339Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1988556Z 	PASSED for resource: jobs(monitoring-init).steps[1](Initialize security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1988713Z 	File: /.github/workflows/security-monitoring-enhanced.yml:61-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1988956Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1989228Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[1](Checkout full repository history)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1989382Z 	File: /.github/workflows/security-monitoring-enhanced.yml:127-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1989613Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1989884Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[2](TruffleHog historical secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1990045Z 	File: /.github/workflows/security-monitoring-enhanced.yml:133-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1990276Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1990526Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[3](GitLeaks comprehensive scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1990680Z 	File: /.github/workflows/security-monitoring-enhanced.yml:142-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1990917Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1991151Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[4](Analyze secret findings)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1991304Z 	File: /.github/workflows/security-monitoring-enhanced.yml:157-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1991537Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1991768Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1992045Z 	File: /.github/workflows/security-monitoring-enhanced.yml:198-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1992278Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1992492Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1992646Z 	File: /.github/workflows/security-monitoring-enhanced.yml:201-209
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1992882Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1993182Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[3](Enhanced vulnerability scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1993338Z 	File: /.github/workflows/security-monitoring-enhanced.yml:208-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1993569Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1993858Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[4](License compliance monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1994019Z 	File: /.github/workflows/security-monitoring-enhanced.yml:247-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1994251Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1994463Z 	PASSED for resource: jobs(container-security-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1994625Z 	File: /.github/workflows/security-monitoring-enhanced.yml:293-298
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1994976Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1995275Z 	PASSED for resource: jobs(container-security-monitoring).steps[2](Build containers for security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1995433Z 	File: /.github/workflows/security-monitoring-enhanced.yml:297-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1995664Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1995940Z 	PASSED for resource: jobs(container-security-monitoring).steps[3](Comprehensive Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1996104Z 	File: /.github/workflows/security-monitoring-enhanced.yml:320-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1996332Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1996614Z 	PASSED for resource: jobs(container-security-monitoring).steps[4](Docker security best practices audit)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1996775Z 	File: /.github/workflows/security-monitoring-enhanced.yml:367-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1997000Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1997199Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1997356Z 	File: /.github/workflows/security-monitoring-enhanced.yml:422-427
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1997584Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1997947Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[2](Generate security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1998112Z 	File: /.github/workflows/security-monitoring-enhanced.yml:426-483
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1998339Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1998574Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[3](Store security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1998737Z 	File: /.github/workflows/security-monitoring-enhanced.yml:482-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1998971Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1999166Z 	PASSED for resource: jobs(security-alerting).steps[1](Determine alert severity)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1999325Z 	File: /.github/workflows/security-monitoring-enhanced.yml:521-562
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1999554Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.1999741Z 	PASSED for resource: jobs(security-alerting).steps[2](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2000058Z 	File: /.github/workflows/security-monitoring-enhanced.yml:561-611
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2000316Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2000509Z 	PASSED for resource: jobs(security-alerting).steps[3](Send Slack notification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2000668Z 	File: /.github/workflows/security-monitoring-enhanced.yml:610-652
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2000899Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001099Z 	PASSED for resource: jobs(security-alerting).steps[4](Generate monitoring report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001256Z 	File: /.github/workflows/security-monitoring-enhanced.yml:651-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001393Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001603Z 	PASSED for resource: jobs(monitoring-init).steps[1](Initialize security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001762Z 	File: /.github/workflows/security-monitoring-enhanced.yml:61-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2001902Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2002167Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[1](Checkout full repository history)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2002699Z 	File: /.github/workflows/security-monitoring-enhanced.yml:127-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2002845Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2003415Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[2](TruffleHog historical secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2003577Z 	File: /.github/workflows/security-monitoring-enhanced.yml:133-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2003713Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2004164Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[3](GitLeaks comprehensive scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2004328Z 	File: /.github/workflows/security-monitoring-enhanced.yml:142-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2004462Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2004704Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[4](Analyze secret findings)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2004944Z 	File: /.github/workflows/security-monitoring-enhanced.yml:157-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2005190Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2005526Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2005696Z 	File: /.github/workflows/security-monitoring-enhanced.yml:198-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2005827Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2006197Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2006358Z 	File: /.github/workflows/security-monitoring-enhanced.yml:201-209
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2006489Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2006784Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[3](Enhanced vulnerability scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2007061Z 	File: /.github/workflows/security-monitoring-enhanced.yml:208-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2007273Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2007564Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[4](License compliance monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2007841Z 	File: /.github/workflows/security-monitoring-enhanced.yml:247-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2007976Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2008181Z 	PASSED for resource: jobs(container-security-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2008341Z 	File: /.github/workflows/security-monitoring-enhanced.yml:293-298
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2008471Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2008767Z 	PASSED for resource: jobs(container-security-monitoring).steps[2](Build containers for security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2009064Z 	File: /.github/workflows/security-monitoring-enhanced.yml:297-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2009197Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2009469Z 	PASSED for resource: jobs(container-security-monitoring).steps[3](Comprehensive Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2009626Z 	File: /.github/workflows/security-monitoring-enhanced.yml:320-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2009760Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010040Z 	PASSED for resource: jobs(container-security-monitoring).steps[4](Docker security best practices audit)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010199Z 	File: /.github/workflows/security-monitoring-enhanced.yml:367-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010327Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010525Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010683Z 	File: /.github/workflows/security-monitoring-enhanced.yml:422-427
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2010817Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011062Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[2](Generate security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011221Z 	File: /.github/workflows/security-monitoring-enhanced.yml:426-483
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011350Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011692Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[3](Store security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011850Z 	File: /.github/workflows/security-monitoring-enhanced.yml:482-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2011979Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012175Z 	PASSED for resource: jobs(security-alerting).steps[1](Determine alert severity)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012336Z 	File: /.github/workflows/security-monitoring-enhanced.yml:521-562
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012465Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012656Z 	PASSED for resource: jobs(security-alerting).steps[2](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012817Z 	File: /.github/workflows/security-monitoring-enhanced.yml:561-611
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2012947Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013140Z 	PASSED for resource: jobs(security-alerting).steps[3](Send Slack notification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013304Z 	File: /.github/workflows/security-monitoring-enhanced.yml:610-652
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013443Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013648Z 	PASSED for resource: jobs(security-alerting).steps[4](Generate monitoring report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013809Z 	File: /.github/workflows/security-monitoring-enhanced.yml:651-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2013990Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2014209Z 	PASSED for resource: jobs(monitoring-init).steps[1](Initialize security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2014368Z 	File: /.github/workflows/security-monitoring-enhanced.yml:61-120
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2014551Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2014826Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[1](Checkout full repository history)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2014985Z 	File: /.github/workflows/security-monitoring-enhanced.yml:127-134
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2015164Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2015437Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[2](TruffleHog historical secret scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2015592Z 	File: /.github/workflows/security-monitoring-enhanced.yml:133-143
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2015766Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2016020Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[3](GitLeaks comprehensive scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2016175Z 	File: /.github/workflows/security-monitoring-enhanced.yml:142-158
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2016437Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2016675Z 	PASSED for resource: jobs(continuous-secret-monitoring).steps[4](Analyze secret findings)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2016830Z 	File: /.github/workflows/security-monitoring-enhanced.yml:157-191
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2017005Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2017242Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2017397Z 	File: /.github/workflows/security-monitoring-enhanced.yml:198-202
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2017571Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2017892Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2018050Z 	File: /.github/workflows/security-monitoring-enhanced.yml:201-209
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2018229Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2018535Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[3](Enhanced vulnerability scanning)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2018691Z 	File: /.github/workflows/security-monitoring-enhanced.yml:208-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2018862Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2019266Z 	PASSED for resource: jobs(dependency-vulnerability-monitoring).steps[4](License compliance monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2019423Z 	File: /.github/workflows/security-monitoring-enhanced.yml:247-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2019598Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2019806Z 	PASSED for resource: jobs(container-security-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2019962Z 	File: /.github/workflows/security-monitoring-enhanced.yml:293-298
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2020134Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2020440Z 	PASSED for resource: jobs(container-security-monitoring).steps[2](Build containers for security monitoring)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2020595Z 	File: /.github/workflows/security-monitoring-enhanced.yml:297-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2020769Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2021049Z 	PASSED for resource: jobs(container-security-monitoring).steps[3](Comprehensive Trivy security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2021209Z 	File: /.github/workflows/security-monitoring-enhanced.yml:320-368
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2021382Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2021673Z 	PASSED for resource: jobs(container-security-monitoring).steps[4](Docker security best practices audit)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2021831Z 	File: /.github/workflows/security-monitoring-enhanced.yml:367-415
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022004Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022217Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022370Z 	File: /.github/workflows/security-monitoring-enhanced.yml:422-427
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022541Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022796Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[2](Generate security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2022957Z 	File: /.github/workflows/security-monitoring-enhanced.yml:426-483
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2023130Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2023369Z 	PASSED for resource: jobs(security-baseline-monitoring).steps[3](Store security baseline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2023525Z 	File: /.github/workflows/security-monitoring-enhanced.yml:482-514
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2023699Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2023902Z 	PASSED for resource: jobs(security-alerting).steps[1](Determine alert severity)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2024175Z 	File: /.github/workflows/security-monitoring-enhanced.yml:521-562
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2024350Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2024532Z 	PASSED for resource: jobs(security-alerting).steps[2](Create security issue)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2024684Z 	File: /.github/workflows/security-monitoring-enhanced.yml:561-611
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2024865Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2025057Z 	PASSED for resource: jobs(security-alerting).steps[3](Send Slack notification)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2025212Z 	File: /.github/workflows/security-monitoring-enhanced.yml:610-652
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2025389Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2025589Z 	PASSED for resource: jobs(security-alerting).steps[4](Generate monitoring report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2025749Z 	File: /.github/workflows/security-monitoring-enhanced.yml:651-692
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026300Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026414Z 	PASSED for resource: on(Optimized CI Pipeline)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026517Z 	File: /.github/workflows/ci-optimized.yml:4-21
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026738Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026829Z 	PASSED for resource: jobs(setup)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2026935Z 	File: /.github/workflows/ci-optimized.yml:42-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027062Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027161Z 	PASSED for resource: jobs(format-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027266Z 	File: /.github/workflows/ci-optimized.yml:89-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027392Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027490Z 	PASSED for resource: jobs(build-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027598Z 	File: /.github/workflows/ci-optimized.yml:124-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027822Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2027913Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028020Z 	File: /.github/workflows/ci-optimized.yml:160-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028139Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028235Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028336Z 	File: /.github/workflows/ci-optimized.yml:259-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028453Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028548Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028651Z 	File: /.github/workflows/ci-optimized.yml:295-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028767Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028869Z 	PASSED for resource: jobs(docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2028971Z 	File: /.github/workflows/ci-optimized.yml:343-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029095Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029186Z 	PASSED for resource: jobs(validate)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029289Z 	File: /.github/workflows/ci-optimized.yml:394-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029527Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029614Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029724Z 	File: /.github/workflows/ci-optimized.yml:41-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2029954Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030045Z 	PASSED for resource: jobs(setup)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030148Z 	File: /.github/workflows/ci-optimized.yml:42-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030376Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030476Z 	PASSED for resource: jobs(format-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030576Z 	File: /.github/workflows/ci-optimized.yml:89-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2030923Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031025Z 	PASSED for resource: jobs(build-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031128Z 	File: /.github/workflows/ci-optimized.yml:124-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031353Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031446Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031549Z 	File: /.github/workflows/ci-optimized.yml:160-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031774Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031862Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2031962Z 	File: /.github/workflows/ci-optimized.yml:259-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032184Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032275Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032382Z 	File: /.github/workflows/ci-optimized.yml:295-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032606Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032701Z 	PASSED for resource: jobs(docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2032800Z 	File: /.github/workflows/ci-optimized.yml:343-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033022Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033220Z 	PASSED for resource: jobs(validate)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033319Z 	File: /.github/workflows/ci-optimized.yml:394-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033459Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033550Z 	PASSED for resource: jobs(setup)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033652Z 	File: /.github/workflows/ci-optimized.yml:42-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033785Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033884Z 	PASSED for resource: jobs(format-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2033991Z 	File: /.github/workflows/ci-optimized.yml:89-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034122Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034219Z 	PASSED for resource: jobs(build-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034322Z 	File: /.github/workflows/ci-optimized.yml:124-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034450Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034542Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034644Z 	File: /.github/workflows/ci-optimized.yml:160-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034771Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034854Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2034965Z 	File: /.github/workflows/ci-optimized.yml:259-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035094Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035182Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035292Z 	File: /.github/workflows/ci-optimized.yml:295-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035426Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035518Z 	PASSED for resource: jobs(docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035625Z 	File: /.github/workflows/ci-optimized.yml:343-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035754Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035839Z 	PASSED for resource: jobs(validate)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2035946Z 	File: /.github/workflows/ci-optimized.yml:394-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036131Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036217Z 	PASSED for resource: jobs(setup)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036326Z 	File: /.github/workflows/ci-optimized.yml:42-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036499Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036590Z 	PASSED for resource: jobs(format-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036697Z 	File: /.github/workflows/ci-optimized.yml:89-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2036870Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037086Z 	PASSED for resource: jobs(build-check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037191Z 	File: /.github/workflows/ci-optimized.yml:124-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037364Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037446Z 	PASSED for resource: jobs(test)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037552Z 	File: /.github/workflows/ci-optimized.yml:160-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037831Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2037916Z 	PASSED for resource: jobs(lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038022Z 	File: /.github/workflows/ci-optimized.yml:259-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038196Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038280Z 	PASSED for resource: jobs(security)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038387Z 	File: /.github/workflows/ci-optimized.yml:295-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038556Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038654Z 	PASSED for resource: jobs(docker-build)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038762Z 	File: /.github/workflows/ci-optimized.yml:343-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2038931Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039016Z 	PASSED for resource: jobs(validate)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039122Z 	File: /.github/workflows/ci-optimized.yml:394-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039485Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039567Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039674Z 	File: /.github/workflows/ci-optimized.yml:41-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039802Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2039936Z 	PASSED for resource: jobs(setup).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040039Z 	File: /.github/workflows/ci-optimized.yml:50-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040157Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040278Z 	PASSED for resource: jobs(setup).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040384Z 	File: /.github/workflows/ci-optimized.yml:55-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040503Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040684Z 	PASSED for resource: jobs(setup).steps[3](Enhanced Go dependency caching)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040784Z 	File: /.github/workflows/ci-optimized.yml:63-77
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2040914Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041066Z 	PASSED for resource: jobs(setup).steps[4](Download dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041173Z 	File: /.github/workflows/ci-optimized.yml:76-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041300Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041448Z 	PASSED for resource: jobs(format-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041551Z 	File: /.github/workflows/ci-optimized.yml:95-101
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041675Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041814Z 	PASSED for resource: jobs(format-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2041918Z 	File: /.github/workflows/ci-optimized.yml:100-107
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042040Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042188Z 	PASSED for resource: jobs(format-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042291Z 	File: /.github/workflows/ci-optimized.yml:106-115
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042419Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042580Z 	PASSED for resource: jobs(format-check).steps[4](Check formatting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042683Z 	File: /.github/workflows/ci-optimized.yml:114-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042803Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2042948Z 	PASSED for resource: jobs(build-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043049Z 	File: /.github/workflows/ci-optimized.yml:130-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043173Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043424Z 	PASSED for resource: jobs(build-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043530Z 	File: /.github/workflows/ci-optimized.yml:135-142
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043659Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043805Z 	PASSED for resource: jobs(build-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2043909Z 	File: /.github/workflows/ci-optimized.yml:141-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044042Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044186Z 	PASSED for resource: jobs(build-check).steps[4](Build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044290Z 	File: /.github/workflows/ci-optimized.yml:149-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044412Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044538Z 	PASSED for resource: jobs(test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044640Z 	File: /.github/workflows/ci-optimized.yml:194-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044762Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044880Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2044983Z 	File: /.github/workflows/ci-optimized.yml:199-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045106Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045226Z 	PASSED for resource: jobs(test).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045328Z 	File: /.github/workflows/ci-optimized.yml:205-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045535Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045674Z 	PASSED for resource: jobs(test).steps[4](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045778Z 	File: /.github/workflows/ci-optimized.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2045905Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046021Z 	PASSED for resource: jobs(test).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046121Z 	File: /.github/workflows/ci-optimized.yml:219-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046242Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046377Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046478Z 	File: /.github/workflows/ci-optimized.yml:247-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046598Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046725Z 	PASSED for resource: jobs(lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046824Z 	File: /.github/workflows/ci-optimized.yml:265-271
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2046943Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047057Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047158Z 	File: /.github/workflows/ci-optimized.yml:270-277
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047274Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047400Z 	PASSED for resource: jobs(lint).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047503Z 	File: /.github/workflows/ci-optimized.yml:276-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047722Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047868Z 	PASSED for resource: jobs(lint).steps[4](Run golangci-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2047975Z 	File: /.github/workflows/ci-optimized.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048094Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048235Z 	PASSED for resource: jobs(security).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048336Z 	File: /.github/workflows/ci-optimized.yml:301-307
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048459Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048580Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048681Z 	File: /.github/workflows/ci-optimized.yml:306-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048797Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2048932Z 	PASSED for resource: jobs(security).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049034Z 	File: /.github/workflows/ci-optimized.yml:312-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049149Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049317Z 	PASSED for resource: jobs(security).steps[4](Install security tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049537Z 	File: /.github/workflows/ci-optimized.yml:320-326
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049656Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049803Z 	PASSED for resource: jobs(security).steps[5](Run security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2049911Z 	File: /.github/workflows/ci-optimized.yml:325-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050032Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050203Z 	PASSED for resource: jobs(security).steps[6](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050306Z 	File: /.github/workflows/ci-optimized.yml:332-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050425Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050573Z 	PASSED for resource: jobs(docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050674Z 	File: /.github/workflows/ci-optimized.yml:353-359
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050792Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2050968Z 	PASSED for resource: jobs(docker-build).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051069Z 	File: /.github/workflows/ci-optimized.yml:358-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051185Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051347Z 	PASSED for resource: jobs(docker-build).steps[3](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051447Z 	File: /.github/workflows/ci-optimized.yml:365-385
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051679Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051840Z 	PASSED for resource: jobs(docker-build).steps[4](Test Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2051939Z 	File: /.github/workflows/ci-optimized.yml:384-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052055Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052218Z 	PASSED for resource: jobs(validate).steps[1](Check pipeline results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052318Z 	File: /.github/workflows/ci-optimized.yml:401-422
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052433Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052611Z 	PASSED for resource: jobs(validate).steps[2](Generate pipeline summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052710Z 	File: /.github/workflows/ci-optimized.yml:421-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2052944Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053079Z 	PASSED for resource: jobs(setup).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053186Z 	File: /.github/workflows/ci-optimized.yml:50-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053413Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053531Z 	PASSED for resource: jobs(setup).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053631Z 	File: /.github/workflows/ci-optimized.yml:55-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2053859Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054042Z 	PASSED for resource: jobs(setup).steps[3](Enhanced Go dependency caching)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054145Z 	File: /.github/workflows/ci-optimized.yml:63-77
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054373Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054529Z 	PASSED for resource: jobs(setup).steps[4](Download dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054629Z 	File: /.github/workflows/ci-optimized.yml:76-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2054852Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055004Z 	PASSED for resource: jobs(format-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055107Z 	File: /.github/workflows/ci-optimized.yml:95-101
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055330Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055468Z 	PASSED for resource: jobs(format-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055570Z 	File: /.github/workflows/ci-optimized.yml:100-107
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2055794Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056030Z 	PASSED for resource: jobs(format-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056131Z 	File: /.github/workflows/ci-optimized.yml:106-115
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056356Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056516Z 	PASSED for resource: jobs(format-check).steps[4](Check formatting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056623Z 	File: /.github/workflows/ci-optimized.yml:114-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056847Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2056999Z 	PASSED for resource: jobs(build-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2057098Z 	File: /.github/workflows/ci-optimized.yml:130-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2057321Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2057457Z 	PASSED for resource: jobs(build-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2057562Z 	File: /.github/workflows/ci-optimized.yml:135-142
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2057913Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058062Z 	PASSED for resource: jobs(build-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058163Z 	File: /.github/workflows/ci-optimized.yml:141-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058391Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058659Z 	PASSED for resource: jobs(build-check).steps[4](Build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058761Z 	File: /.github/workflows/ci-optimized.yml:149-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2058987Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059112Z 	PASSED for resource: jobs(test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059215Z 	File: /.github/workflows/ci-optimized.yml:194-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059442Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059560Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059661Z 	File: /.github/workflows/ci-optimized.yml:199-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2059883Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060011Z 	PASSED for resource: jobs(test).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060118Z 	File: /.github/workflows/ci-optimized.yml:205-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060339Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060483Z 	PASSED for resource: jobs(test).steps[4](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060585Z 	File: /.github/workflows/ci-optimized.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060808Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2060925Z 	PASSED for resource: jobs(test).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061029Z 	File: /.github/workflows/ci-optimized.yml:219-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061252Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061384Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061483Z 	File: /.github/workflows/ci-optimized.yml:247-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061705Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061834Z 	PASSED for resource: jobs(lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2061937Z 	File: /.github/workflows/ci-optimized.yml:265-271
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062160Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062276Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062376Z 	File: /.github/workflows/ci-optimized.yml:270-277
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062598Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062842Z 	PASSED for resource: jobs(lint).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2062944Z 	File: /.github/workflows/ci-optimized.yml:276-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063168Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063309Z 	PASSED for resource: jobs(lint).steps[4](Run golangci-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063413Z 	File: /.github/workflows/ci-optimized.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063639Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063775Z 	PASSED for resource: jobs(security).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2063874Z 	File: /.github/workflows/ci-optimized.yml:301-307
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064096Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064219Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064325Z 	File: /.github/workflows/ci-optimized.yml:306-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064546Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064682Z 	PASSED for resource: jobs(security).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2064784Z 	File: /.github/workflows/ci-optimized.yml:312-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065007Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065255Z 	PASSED for resource: jobs(security).steps[4](Install security tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065358Z 	File: /.github/workflows/ci-optimized.yml:320-326
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065584Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065732Z 	PASSED for resource: jobs(security).steps[5](Run security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2065836Z 	File: /.github/workflows/ci-optimized.yml:325-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066065Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066242Z 	PASSED for resource: jobs(security).steps[6](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066350Z 	File: /.github/workflows/ci-optimized.yml:332-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066577Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066729Z 	PASSED for resource: jobs(docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2066834Z 	File: /.github/workflows/ci-optimized.yml:353-359
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067058Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067236Z 	PASSED for resource: jobs(docker-build).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067341Z 	File: /.github/workflows/ci-optimized.yml:358-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067567Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067838Z 	PASSED for resource: jobs(docker-build).steps[3](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2067947Z 	File: /.github/workflows/ci-optimized.yml:365-385
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068175Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068339Z 	PASSED for resource: jobs(docker-build).steps[4](Test Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068439Z 	File: /.github/workflows/ci-optimized.yml:384-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068669Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068832Z 	PASSED for resource: jobs(validate).steps[1](Check pipeline results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2068932Z 	File: /.github/workflows/ci-optimized.yml:401-422
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069160Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069334Z 	PASSED for resource: jobs(validate).steps[2](Generate pipeline summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069435Z 	File: /.github/workflows/ci-optimized.yml:421-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069696Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069829Z 	PASSED for resource: jobs(setup).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2069934Z 	File: /.github/workflows/ci-optimized.yml:50-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070064Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070183Z 	PASSED for resource: jobs(setup).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070288Z 	File: /.github/workflows/ci-optimized.yml:55-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070414Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070594Z 	PASSED for resource: jobs(setup).steps[3](Enhanced Go dependency caching)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070693Z 	File: /.github/workflows/ci-optimized.yml:63-77
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070820Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2070978Z 	PASSED for resource: jobs(setup).steps[4](Download dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071079Z 	File: /.github/workflows/ci-optimized.yml:76-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071212Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071365Z 	PASSED for resource: jobs(format-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071470Z 	File: /.github/workflows/ci-optimized.yml:95-101
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071596Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071731Z 	PASSED for resource: jobs(format-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2071947Z 	File: /.github/workflows/ci-optimized.yml:100-107
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072078Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072225Z 	PASSED for resource: jobs(format-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072326Z 	File: /.github/workflows/ci-optimized.yml:106-115
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072455Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072615Z 	PASSED for resource: jobs(format-check).steps[4](Check formatting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072716Z 	File: /.github/workflows/ci-optimized.yml:114-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072849Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2072997Z 	PASSED for resource: jobs(build-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073097Z 	File: /.github/workflows/ci-optimized.yml:130-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073224Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073368Z 	PASSED for resource: jobs(build-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073469Z 	File: /.github/workflows/ci-optimized.yml:135-142
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073596Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073743Z 	PASSED for resource: jobs(build-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073848Z 	File: /.github/workflows/ci-optimized.yml:141-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2073976Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074124Z 	PASSED for resource: jobs(build-check).steps[4](Build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074229Z 	File: /.github/workflows/ci-optimized.yml:149-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074356Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074486Z 	PASSED for resource: jobs(test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074587Z 	File: /.github/workflows/ci-optimized.yml:194-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074714Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074833Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2074936Z 	File: /.github/workflows/ci-optimized.yml:199-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075063Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075194Z 	PASSED for resource: jobs(test).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075294Z 	File: /.github/workflows/ci-optimized.yml:205-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075422Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075560Z 	PASSED for resource: jobs(test).steps[4](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075747Z 	File: /.github/workflows/ci-optimized.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075874Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2075989Z 	PASSED for resource: jobs(test).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076089Z 	File: /.github/workflows/ci-optimized.yml:219-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076217Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076351Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076451Z 	File: /.github/workflows/ci-optimized.yml:247-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076578Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076700Z 	PASSED for resource: jobs(lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076804Z 	File: /.github/workflows/ci-optimized.yml:265-271
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2076930Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077038Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077146Z 	File: /.github/workflows/ci-optimized.yml:270-277
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077275Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077397Z 	PASSED for resource: jobs(lint).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077500Z 	File: /.github/workflows/ci-optimized.yml:276-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077726Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2077984Z 	PASSED for resource: jobs(lint).steps[4](Run golangci-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078090Z 	File: /.github/workflows/ci-optimized.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078222Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078355Z 	PASSED for resource: jobs(security).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078460Z 	File: /.github/workflows/ci-optimized.yml:301-307
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078587Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078710Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078821Z 	File: /.github/workflows/ci-optimized.yml:306-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2078947Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079079Z 	PASSED for resource: jobs(security).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079189Z 	File: /.github/workflows/ci-optimized.yml:312-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079316Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079481Z 	PASSED for resource: jobs(security).steps[4](Install security tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079586Z 	File: /.github/workflows/ci-optimized.yml:320-326
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079716Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079861Z 	PASSED for resource: jobs(security).steps[5](Run security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2079970Z 	File: /.github/workflows/ci-optimized.yml:325-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080096Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080260Z 	PASSED for resource: jobs(security).steps[6](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080369Z 	File: /.github/workflows/ci-optimized.yml:332-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080495Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080640Z 	PASSED for resource: jobs(docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080742Z 	File: /.github/workflows/ci-optimized.yml:353-359
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2080868Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081043Z 	PASSED for resource: jobs(docker-build).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081148Z 	File: /.github/workflows/ci-optimized.yml:358-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081273Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081430Z 	PASSED for resource: jobs(docker-build).steps[3](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081536Z 	File: /.github/workflows/ci-optimized.yml:365-385
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081665Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2081937Z 	PASSED for resource: jobs(docker-build).steps[4](Test Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082043Z 	File: /.github/workflows/ci-optimized.yml:384-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082172Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082332Z 	PASSED for resource: jobs(validate).steps[1](Check pipeline results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082437Z 	File: /.github/workflows/ci-optimized.yml:401-422
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082570Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082740Z 	PASSED for resource: jobs(validate).steps[2](Generate pipeline summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2082844Z 	File: /.github/workflows/ci-optimized.yml:421-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083025Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083153Z 	PASSED for resource: jobs(setup).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083257Z 	File: /.github/workflows/ci-optimized.yml:50-56
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083430Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083548Z 	PASSED for resource: jobs(setup).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083650Z 	File: /.github/workflows/ci-optimized.yml:55-64
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2083823Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084000Z 	PASSED for resource: jobs(setup).steps[3](Enhanced Go dependency caching)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084187Z 	File: /.github/workflows/ci-optimized.yml:63-77
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084360Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084512Z 	PASSED for resource: jobs(setup).steps[4](Download dependencies)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084618Z 	File: /.github/workflows/ci-optimized.yml:76-89
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084789Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2084935Z 	PASSED for resource: jobs(format-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085043Z 	File: /.github/workflows/ci-optimized.yml:95-101
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085220Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085353Z 	PASSED for resource: jobs(format-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085458Z 	File: /.github/workflows/ci-optimized.yml:100-107
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085632Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085778Z 	PASSED for resource: jobs(format-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2085882Z 	File: /.github/workflows/ci-optimized.yml:106-115
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086052Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086210Z 	PASSED for resource: jobs(format-check).steps[4](Check formatting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086312Z 	File: /.github/workflows/ci-optimized.yml:114-124
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086480Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086623Z 	PASSED for resource: jobs(build-check).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086731Z 	File: /.github/workflows/ci-optimized.yml:130-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2086900Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087029Z 	PASSED for resource: jobs(build-check).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087134Z 	File: /.github/workflows/ci-optimized.yml:135-142
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087307Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087448Z 	PASSED for resource: jobs(build-check).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087551Z 	File: /.github/workflows/ci-optimized.yml:141-150
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087832Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2087974Z 	PASSED for resource: jobs(build-check).steps[4](Build check)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088073Z 	File: /.github/workflows/ci-optimized.yml:149-160
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088246Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088532Z 	PASSED for resource: jobs(test).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088634Z 	File: /.github/workflows/ci-optimized.yml:194-200
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088811Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2088920Z 	PASSED for resource: jobs(test).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089027Z 	File: /.github/workflows/ci-optimized.yml:199-206
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089200Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089324Z 	PASSED for resource: jobs(test).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089425Z 	File: /.github/workflows/ci-optimized.yml:205-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089599Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089732Z 	PASSED for resource: jobs(test).steps[4](Wait for registry)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2089835Z 	File: /.github/workflows/ci-optimized.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090018Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090130Z 	PASSED for resource: jobs(test).steps[5](Run tests)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090229Z 	File: /.github/workflows/ci-optimized.yml:219-248
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090404Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090532Z 	PASSED for resource: jobs(test).steps[6](Upload coverage)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090743Z 	File: /.github/workflows/ci-optimized.yml:247-259
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2090921Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091043Z 	PASSED for resource: jobs(lint).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091144Z 	File: /.github/workflows/ci-optimized.yml:265-271
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091317Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091424Z 	PASSED for resource: jobs(lint).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091530Z 	File: /.github/workflows/ci-optimized.yml:270-277
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091701Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091819Z 	PASSED for resource: jobs(lint).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2091919Z 	File: /.github/workflows/ci-optimized.yml:276-286
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092089Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092227Z 	PASSED for resource: jobs(lint).steps[4](Run golangci-lint)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092327Z 	File: /.github/workflows/ci-optimized.yml:285-295
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092501Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092633Z 	PASSED for resource: jobs(security).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092736Z 	File: /.github/workflows/ci-optimized.yml:301-307
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2092908Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093033Z 	PASSED for resource: jobs(security).steps[2](Setup Go)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093135Z 	File: /.github/workflows/ci-optimized.yml:306-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093310Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093442Z 	PASSED for resource: jobs(security).steps[3](Restore cache)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093544Z 	File: /.github/workflows/ci-optimized.yml:312-321
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093722Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093881Z 	PASSED for resource: jobs(security).steps[4](Install security tools)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2093984Z 	File: /.github/workflows/ci-optimized.yml:320-326
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094156Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094299Z 	PASSED for resource: jobs(security).steps[5](Run security scan)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094409Z 	File: /.github/workflows/ci-optimized.yml:325-333
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094588Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094842Z 	PASSED for resource: jobs(security).steps[6](Upload security results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2094948Z 	File: /.github/workflows/ci-optimized.yml:332-343
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095127Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095274Z 	PASSED for resource: jobs(docker-build).steps[1](Checkout code)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095384Z 	File: /.github/workflows/ci-optimized.yml:353-359
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095561Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095733Z 	PASSED for resource: jobs(docker-build).steps[2](Set up Docker Buildx)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2095835Z 	File: /.github/workflows/ci-optimized.yml:358-366
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096010Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096169Z 	PASSED for resource: jobs(docker-build).steps[3](Build Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096275Z 	File: /.github/workflows/ci-optimized.yml:365-385
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096450Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096607Z 	PASSED for resource: jobs(docker-build).steps[4](Test Docker image)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096708Z 	File: /.github/workflows/ci-optimized.yml:384-394
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2096880Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097128Z 	PASSED for resource: jobs(validate).steps[1](Check pipeline results)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097229Z 	File: /.github/workflows/ci-optimized.yml:401-422
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097404Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097574Z 	PASSED for resource: jobs(validate).steps[2](Generate pipeline summary)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097775Z 	File: /.github/workflows/ci-optimized.yml:421-449
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2097906Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098010Z 	PASSED for resource: jobs(aws-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098137Z 	File: /.github/workflows/oidc-authentication.yml:54-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098260Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098355Z 	PASSED for resource: jobs(gcp-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098480Z 	File: /.github/workflows/oidc-authentication.yml:148-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098609Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098707Z 	PASSED for resource: jobs(azure-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098830Z 	File: /.github/workflows/oidc-authentication.yml:255-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2098951Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099071Z 	PASSED for resource: jobs(oidc-security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099194Z 	File: /.github/workflows/oidc-authentication.yml:338-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099438Z Check: CKV_GHA_5: "Found artifact build without evidence of cosign sign execution in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099526Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099650Z 	File: /.github/workflows/oidc-authentication.yml:53-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099888Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2099982Z 	PASSED for resource: jobs(aws-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100101Z 	File: /.github/workflows/oidc-authentication.yml:54-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100334Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100426Z 	PASSED for resource: jobs(gcp-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100547Z 	File: /.github/workflows/oidc-authentication.yml:148-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100779Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100877Z 	PASSED for resource: jobs(azure-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2100998Z 	File: /.github/workflows/oidc-authentication.yml:255-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101224Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101470Z 	PASSED for resource: jobs(oidc-security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101592Z 	File: /.github/workflows/oidc-authentication.yml:338-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101730Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101823Z 	PASSED for resource: jobs(aws-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2101948Z 	File: /.github/workflows/oidc-authentication.yml:54-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102084Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102174Z 	PASSED for resource: jobs(gcp-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102293Z 	File: /.github/workflows/oidc-authentication.yml:148-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102428Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102523Z 	PASSED for resource: jobs(azure-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102644Z 	File: /.github/workflows/oidc-authentication.yml:255-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102782Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2102898Z 	PASSED for resource: jobs(oidc-security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103019Z 	File: /.github/workflows/oidc-authentication.yml:338-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103200Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103291Z 	PASSED for resource: jobs(aws-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103520Z 	File: /.github/workflows/oidc-authentication.yml:54-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103704Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103795Z 	PASSED for resource: jobs(gcp-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2103917Z 	File: /.github/workflows/oidc-authentication.yml:148-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104123Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104220Z 	PASSED for resource: jobs(azure-oidc-auth)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104338Z 	File: /.github/workflows/oidc-authentication.yml:255-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104518Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104635Z 	PASSED for resource: jobs(oidc-security-validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2104756Z 	File: /.github/workflows/oidc-authentication.yml:338-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105010Z Check: CKV_GHA_6: "Found artifact build without evidence of cosign sbom attestation in pipeline"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105097Z 	PASSED for resource: jobs
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105219Z 	File: /.github/workflows/oidc-authentication.yml:53-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105344Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105514Z 	PASSED for resource: jobs(aws-oidc-auth).steps[1](Configure AWS OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105638Z 	File: /.github/workflows/oidc-authentication.yml:64-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105762Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2105974Z 	PASSED for resource: jobs(aws-oidc-auth).steps[2](Configure AWS credentials via OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106099Z 	File: /.github/workflows/oidc-authentication.yml:99-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106222Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106429Z 	PASSED for resource: jobs(aws-oidc-auth).steps[3](Validate AWS OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106554Z 	File: /.github/workflows/oidc-authentication.yml:110-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106671Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106851Z 	PASSED for resource: jobs(aws-oidc-auth).steps[4](Login to Amazon ECR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2106975Z 	File: /.github/workflows/oidc-authentication.yml:135-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2107094Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2107263Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[1](Configure GCP OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2107384Z 	File: /.github/workflows/oidc-authentication.yml:158-203
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2107501Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2107889Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[2](Authenticate to Google Cloud)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108015Z 	File: /.github/workflows/oidc-authentication.yml:202-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108137Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108315Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[3](Setup Google Cloud SDK)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108438Z 	File: /.github/workflows/oidc-authentication.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108562Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108769Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[4](Validate GCP OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2108892Z 	File: /.github/workflows/oidc-authentication.yml:219-245
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109010Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109248Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[5](Configure Artifact Registry authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109371Z 	File: /.github/workflows/oidc-authentication.yml:244-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109498Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109682Z 	PASSED for resource: jobs(azure-oidc-auth).steps[1](Configure Azure OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109806Z 	File: /.github/workflows/oidc-authentication.yml:265-305
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2109925Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110090Z 	PASSED for resource: jobs(azure-oidc-auth).steps[2](Login to Azure)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110331Z 	File: /.github/workflows/oidc-authentication.yml:304-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110453Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110675Z 	PASSED for resource: jobs(azure-oidc-auth).steps[3](Validate Azure OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110798Z 	File: /.github/workflows/oidc-authentication.yml:312-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2110918Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111166Z 	PASSED for resource: jobs(oidc-security-validation).steps[1](OIDC Token Security Validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111294Z 	File: /.github/workflows/oidc-authentication.yml:346-400
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111412Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111659Z 	PASSED for resource: jobs(oidc-security-validation).steps[2](Token Security Best Practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111780Z 	File: /.github/workflows/oidc-authentication.yml:399-447
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2111905Z Check: CKV_GHA_3: "Suspicious use of curl with secrets"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2112165Z 	PASSED for resource: jobs(oidc-security-validation).steps[3](Generate OIDC Configuration Report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2112284Z 	File: /.github/workflows/oidc-authentication.yml:446-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2112516Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2112685Z 	PASSED for resource: jobs(aws-oidc-auth).steps[1](Configure AWS OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2112810Z 	File: /.github/workflows/oidc-authentication.yml:64-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113044Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113255Z 	PASSED for resource: jobs(aws-oidc-auth).steps[2](Configure AWS credentials via OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113377Z 	File: /.github/workflows/oidc-authentication.yml:99-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113606Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113816Z 	PASSED for resource: jobs(aws-oidc-auth).steps[3](Validate AWS OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2113937Z 	File: /.github/workflows/oidc-authentication.yml:110-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2114165Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2114336Z 	PASSED for resource: jobs(aws-oidc-auth).steps[4](Login to Amazon ECR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2114459Z 	File: /.github/workflows/oidc-authentication.yml:135-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2114685Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2114974Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[1](Configure GCP OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2115097Z 	File: /.github/workflows/oidc-authentication.yml:158-203
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2115332Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2115533Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[2](Authenticate to Google Cloud)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2115660Z 	File: /.github/workflows/oidc-authentication.yml:202-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2115892Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116068Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[3](Setup Google Cloud SDK)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116190Z 	File: /.github/workflows/oidc-authentication.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116418Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116626Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[4](Validate GCP OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116748Z 	File: /.github/workflows/oidc-authentication.yml:219-245
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2116980Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2117210Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[5](Configure Artifact Registry authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2117411Z 	File: /.github/workflows/oidc-authentication.yml:244-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2117745Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2117928Z 	PASSED for resource: jobs(azure-oidc-auth).steps[1](Configure Azure OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2118051Z 	File: /.github/workflows/oidc-authentication.yml:265-305
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2118281Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2118442Z 	PASSED for resource: jobs(azure-oidc-auth).steps[2](Login to Azure)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2118569Z 	File: /.github/workflows/oidc-authentication.yml:304-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2118802Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119021Z 	PASSED for resource: jobs(azure-oidc-auth).steps[3](Validate Azure OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119144Z 	File: /.github/workflows/oidc-authentication.yml:312-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119382Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119623Z 	PASSED for resource: jobs(oidc-security-validation).steps[1](OIDC Token Security Validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119747Z 	File: /.github/workflows/oidc-authentication.yml:346-400
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2119979Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2120220Z 	PASSED for resource: jobs(oidc-security-validation).steps[2](Token Security Best Practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2120347Z 	File: /.github/workflows/oidc-authentication.yml:399-447
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2120576Z Check: CKV_GHA_1: "Ensure ACTIONS_ALLOW_UNSECURE_COMMANDS isn't true on environment variables"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2120836Z 	PASSED for resource: jobs(oidc-security-validation).steps[3](Generate OIDC Configuration Report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2120961Z 	File: /.github/workflows/oidc-authentication.yml:446-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121111Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121275Z 	PASSED for resource: jobs(aws-oidc-auth).steps[1](Configure AWS OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121398Z 	File: /.github/workflows/oidc-authentication.yml:64-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121534Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121742Z 	PASSED for resource: jobs(aws-oidc-auth).steps[2](Configure AWS credentials via OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2121865Z 	File: /.github/workflows/oidc-authentication.yml:99-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122124Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122328Z 	PASSED for resource: jobs(aws-oidc-auth).steps[3](Validate AWS OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122451Z 	File: /.github/workflows/oidc-authentication.yml:110-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122584Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122750Z 	PASSED for resource: jobs(aws-oidc-auth).steps[4](Login to Amazon ECR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2122878Z 	File: /.github/workflows/oidc-authentication.yml:135-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123012Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123175Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[1](Configure GCP OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123297Z 	File: /.github/workflows/oidc-authentication.yml:158-203
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123429Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123622Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[2](Authenticate to Google Cloud)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123756Z 	File: /.github/workflows/oidc-authentication.yml:202-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2123896Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124074Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[3](Setup Google Cloud SDK)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124200Z 	File: /.github/workflows/oidc-authentication.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124334Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124650Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[4](Validate GCP OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124778Z 	File: /.github/workflows/oidc-authentication.yml:219-245
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2124913Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125149Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[5](Configure Artifact Registry authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125276Z 	File: /.github/workflows/oidc-authentication.yml:244-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125410Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125592Z 	PASSED for resource: jobs(azure-oidc-auth).steps[1](Configure Azure OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125713Z 	File: /.github/workflows/oidc-authentication.yml:265-305
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2125844Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126008Z 	PASSED for resource: jobs(azure-oidc-auth).steps[2](Login to Azure)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126134Z 	File: /.github/workflows/oidc-authentication.yml:304-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126267Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126486Z 	PASSED for resource: jobs(azure-oidc-auth).steps[3](Validate Azure OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126608Z 	File: /.github/workflows/oidc-authentication.yml:312-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126741Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2126983Z 	PASSED for resource: jobs(oidc-security-validation).steps[1](OIDC Token Security Validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2127104Z 	File: /.github/workflows/oidc-authentication.yml:346-400
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2127239Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2127475Z 	PASSED for resource: jobs(oidc-security-validation).steps[2](Token Security Best Practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2127597Z 	File: /.github/workflows/oidc-authentication.yml:399-447
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2127835Z Check: CKV_GHA_4: "Suspicious use of netcat with IP address"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128103Z 	PASSED for resource: jobs(oidc-security-validation).steps[3](Generate OIDC Configuration Report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128227Z 	File: /.github/workflows/oidc-authentication.yml:446-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128413Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128580Z 	PASSED for resource: jobs(aws-oidc-auth).steps[1](Configure AWS OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128705Z 	File: /.github/workflows/oidc-authentication.yml:64-100
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2128887Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2129219Z 	PASSED for resource: jobs(aws-oidc-auth).steps[2](Configure AWS credentials via OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2129342Z 	File: /.github/workflows/oidc-authentication.yml:99-111
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2129521Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2129726Z 	PASSED for resource: jobs(aws-oidc-auth).steps[3](Validate AWS OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2129855Z 	File: /.github/workflows/oidc-authentication.yml:110-136
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130033Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130206Z 	PASSED for resource: jobs(aws-oidc-auth).steps[4](Login to Amazon ECR)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130327Z 	File: /.github/workflows/oidc-authentication.yml:135-148
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130504Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130671Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[1](Configure GCP OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130798Z 	File: /.github/workflows/oidc-authentication.yml:158-203
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2130974Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2131168Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[2](Authenticate to Google Cloud)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2131291Z 	File: /.github/workflows/oidc-authentication.yml:202-214
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2131574Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2131751Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[3](Setup Google Cloud SDK)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2131875Z 	File: /.github/workflows/oidc-authentication.yml:213-220
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132052Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132256Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[4](Validate GCP OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132381Z 	File: /.github/workflows/oidc-authentication.yml:219-245
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132563Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132799Z 	PASSED for resource: jobs(gcp-oidc-auth).steps[5](Configure Artifact Registry authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2132921Z 	File: /.github/workflows/oidc-authentication.yml:244-255
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133097Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133278Z 	PASSED for resource: jobs(azure-oidc-auth).steps[1](Configure Azure OIDC)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133404Z 	File: /.github/workflows/oidc-authentication.yml:265-305
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133580Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133738Z 	PASSED for resource: jobs(azure-oidc-auth).steps[2](Login to Azure)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2133859Z 	File: /.github/workflows/oidc-authentication.yml:304-313
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134039Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134260Z 	PASSED for resource: jobs(azure-oidc-auth).steps[3](Validate Azure OIDC authentication)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134382Z 	File: /.github/workflows/oidc-authentication.yml:312-338
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134560Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134805Z 	PASSED for resource: jobs(oidc-security-validation).steps[1](OIDC Token Security Validation)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2134932Z 	File: /.github/workflows/oidc-authentication.yml:346-400
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2135108Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2135345Z 	PASSED for resource: jobs(oidc-security-validation).steps[2](Token Security Best Practices)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2135465Z 	File: /.github/workflows/oidc-authentication.yml:399-447
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2135638Z Check: CKV_GHA_2: "Ensure run commands are not vulnerable to shell injection"
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2135894Z 	PASSED for resource: jobs(oidc-security-validation).steps[3](Generate OIDC Configuration Report)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2136103Z 	File: /.github/workflows/oidc-authentication.yml:446-496
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2136653Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2136786Z 	FAILED for resource: on(Security Monitoring & Alerting)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2138209Z ##[error]	File: /.github/workflows/security-monitoring.yml:9-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2138895Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2138986Z 		9  |       scan_type:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139109Z 		10 |         description: 'Type of security scan to run'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139190Z 		11 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139268Z 		12 |         default: 'full'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139343Z 		13 |         type: choice
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139413Z 		14 |         options:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139492Z 		15 |           - 'full'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139579Z 		16 |           - 'critical-only'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139678Z 		17 |           - 'dependencies-only'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139760Z 		18 |           - 'containers-only'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139839Z 		19 |       alert_level:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2139947Z 		20 |         description: 'Alert severity threshold'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140022Z 		21 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140096Z 		22 |         default: 'high'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140356Z 		23 |         type: choice
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140426Z 		24 |         options:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140502Z 		25 |           - 'critical'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140577Z 		26 |           - 'high'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140647Z 		27 |           - 'medium'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140717Z 		28 |           - 'low'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140784Z 		29 | 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140905Z 		30 | # Security: Minimal permissions for monitoring
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2140978Z 		31 | permissions:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2141050Z 		32 |   contents: read
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2141058Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2141634Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2141739Z 	FAILED for resource: on(Comprehensive)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2142529Z ##[error]	File: /.github/workflows/scheduled-comprehensive.yml:11-29
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2142943Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143037Z 		11 |       run_external_deps:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143212Z 		12 |         description: 'Run tests requiring external dependencies (AWS/GCP)'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143290Z 		13 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143365Z 		14 |         default: 'true'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143444Z 		15 |         type: boolean
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143523Z 		16 |       run_flaky_detection:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143664Z 		17 |         description: 'Run flaky test detection (multiple runs)'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143743Z 		18 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143816Z 		19 |         default: 'true'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143894Z 		20 |         type: boolean
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2143969Z 		21 |       test_iterations:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144127Z 		22 |         description: 'Number of test iterations for flaky detection'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144204Z 		23 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144277Z 		24 |         default: '3'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144354Z 		25 |         type: string
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144423Z 		26 | 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144536Z 		27 | # Separate concurrency for comprehensive tests
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144611Z 		28 | concurrency:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144721Z 		29 |   group: comprehensive-${{ github.run_id }}
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2144729Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2145278Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2145417Z 	FAILED for resource: on(Security Monitoring Enhanced)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2146330Z ##[error]	File: /.github/workflows/security-monitoring-enhanced.yml:13-32
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2146896Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2146977Z 		13 |       scan_type:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147094Z 		14 |         description: 'Type of security scan to run'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147176Z 		15 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147250Z 		16 |         default: 'full'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147323Z 		17 |         type: choice
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147398Z 		18 |         options:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147477Z 		19 |           - 'full'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147550Z 		20 |           - 'quick'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147759Z 		21 |           - 'dependencies'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147847Z 		22 |           - 'containers'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147922Z 		23 |           - 'secrets'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2147999Z 		24 |       notify_on_success:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148153Z 		25 |         description: 'Send notifications on successful scans'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148233Z 		26 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148310Z 		27 |         default: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148391Z 		28 |         type: boolean
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148463Z 		29 | 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148571Z 		30 | # SECURITY: Minimal required permissions
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148645Z 		31 | permissions:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148723Z 		32 |   contents: read
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2148730Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2149293Z Check: CKV_GHA_7: "The build output cannot be affected by user parameters other than the build entry point and the top-level source location. GitHub Actions workflow_dispatch inputs MUST be empty. "
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2149561Z 	FAILED for resource: on(OIDC Authentication Setup)
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2150349Z ##[error]	File: /.github/workflows/oidc-authentication.yml:22-43
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2150788Z 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2150878Z 		22 |       cloud_provider:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2150987Z 		23 |         description: 'Cloud provider for OIDC'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151064Z 		24 |         required: true
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151143Z 		25 |         type: choice
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151219Z 		26 |         options:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151291Z 		27 |           - 'aws'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151363Z 		28 |           - 'gcp'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151439Z 		29 |           - 'azure'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151508Z 		30 |           - 'all'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151581Z 		31 |       environment:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151690Z 		32 |         description: 'Deployment environment'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151771Z 		33 |         required: false
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151855Z 		34 |         default: 'development'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151931Z 		35 |         type: choice
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2151999Z 		36 |         options:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152076Z 		37 |           - 'development'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152147Z 		38 |           - 'staging'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152225Z 		39 |           - 'production'
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152288Z 		40 | 
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152433Z 		41 | # SECURITY: Minimal permissions with OIDC token capability
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152512Z 		42 | permissions:
Infrastructure Security Scanning	Run Checkov IaC scan	2025-08-03T18:58:08.2152590Z 		43 |   contents: read
Dependency Scanning	Run govulncheck	﻿2025-08-03T18:58:00.8149331Z ##[group]Run echo "🔍 Running Go vulnerability database check"
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8149829Z [36;1mecho "🔍 Running Go vulnerability database check"[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8150129Z [36;1m[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8150334Z [36;1m# SECURITY: Install govulncheck[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8150658Z [36;1mgo install golang.org/x/vuln/cmd/govulncheck@latest[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8213478Z [36;1m[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8213805Z [36;1m# SECURITY: Scan for vulnerabilities[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8214201Z [36;1mif govulncheck -json ./... > govulncheck-results.json; then[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8214676Z [36;1m  echo "✅ govulncheck completed successfully"[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8215026Z [36;1melse[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8215296Z [36;1m  echo "❌ govulncheck found vulnerabilities"[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8215874Z [36;1m  cat govulncheck-results.json[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8216128Z [36;1m  exit 1[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8216306Z [36;1mfi[0m
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8251485Z shell: /usr/bin/bash -e {0}
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8251851Z env:
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8252078Z   SEVERITY_THRESHOLD: HIGH
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8252301Z   SCAN_TIMEOUT: 600
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8252499Z   MAX_VULNERABILITIES_HIGH: 0
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8252726Z   MAX_VULNERABILITIES_CRITICAL: 0
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8252967Z ##[endgroup]
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.8325247Z 🔍 Running Go vulnerability database check
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.9023809Z go: downloading golang.org/x/vuln v1.1.4
Dependency Scanning	Run govulncheck	2025-08-03T18:58:00.9786072Z go: downloading golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7
Dependency Scanning	Run govulncheck	2025-08-03T18:58:01.0196314Z go: downloading golang.org/x/mod v0.22.0
Dependency Scanning	Run govulncheck	2025-08-03T18:58:01.0199664Z go: downloading golang.org/x/tools v0.29.0
Dependency Scanning	Run govulncheck	2025-08-03T18:58:01.0976807Z go: downloading golang.org/x/sync v0.10.0
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4650537Z govulncheck: loading packages: 
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4652582Z There are errors with the provided package patterns:
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4652973Z 
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4679437Z ##[error]/home/runner/go/pkg/mod/golang.org/x/sys@v0.31.0/unix/vgetrandom_linux.go:7:9: file requires newer Go version go1.24 (application built with go1.23)
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4688610Z ##[error]/home/runner/go/pkg/mod/golang.org/x/net@v0.38.0/http2/config_go124.go:7:9: file requires newer Go version go1.24 (application built with go1.23)
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4689582Z 
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4689902Z For details on package patterns, see https://pkg.go.dev/cmd/go#hdr-Package_lists_and_patterns.
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.4690247Z 
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5101865Z ❌ govulncheck found vulnerabilities
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5113024Z {
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5113381Z   "config": {
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5113792Z     "protocol_version": "v1.0.0",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5114198Z     "scanner_name": "govulncheck",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5114527Z     "scanner_version": "v1.1.4",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5114863Z     "db": "https://vuln.go.dev",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5115235Z     "db_last_modified": "2025-07-30T21:37:48Z",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5115854Z     "go_version": "go1.24.5",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5116150Z     "scan_level": "symbol",
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5116466Z     "scan_mode": "source"
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5116692Z   }
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5116930Z }
Dependency Scanning	Run govulncheck	2025-08-03T18:58:19.5124111Z ##[error]Process completed with exit code 1.
Security Policy Compliance	Enforce security gates	﻿2025-08-03T18:58:28.7892651Z ##[group]Run echo "🚨 Enforcing final security gates"
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7893967Z [36;1mecho "🚨 Enforcing final security gates"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7894984Z [36;1m[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7895758Z [36;1mcompliance_status="FAILED"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7896712Z [36;1msecurity_level=""[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7897563Z [36;1m[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7898318Z [36;1mcase "$compliance_status" in[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7899271Z [36;1m  "PASSED")[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7900158Z [36;1m    echo "✅ SECURITY GATES PASSED"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7901457Z [36;1m    echo "🎉 All security requirements met - deployment authorized"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7902915Z [36;1m    ;;[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7903692Z [36;1m  "CONDITIONAL_PASS")[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7904728Z [36;1m    if [[ "$security_level" == "production" ]]; then[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7905866Z [36;1m      echo "❌ SECURITY GATES FAILED"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7907310Z [36;1m      echo "🚨 Conditional pass not allowed for production - fix all security issues"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7908726Z [36;1m      exit 1[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7909504Z [36;1m    else[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7910406Z [36;1m      echo "⚠️ SECURITY GATES CONDITIONALLY PASSED"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7911735Z [36;1m      echo "🔍 Development environment - monitoring required"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7913121Z [36;1m    fi[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7913847Z [36;1m    ;;[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7914579Z [36;1m  "FAILED")[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7915416Z [36;1m    echo "❌ SECURITY GATES FAILED"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7916707Z [36;1m    echo "🚨 Critical security issues detected - deployment blocked"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7918214Z [36;1m    echo "🔧 Fix all security vulnerabilities before proceeding"[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7919389Z [36;1m    exit 1[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7920412Z [36;1m    ;;[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7921563Z [36;1mesac[0m
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7954610Z shell: /usr/bin/bash -e {0}
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7955489Z env:
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7956311Z   SEVERITY_THRESHOLD: HIGH
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7957189Z   SCAN_TIMEOUT: 600
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7958002Z   MAX_VULNERABILITIES_HIGH: 0
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7958967Z   MAX_VULNERABILITIES_CRITICAL: 0
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.7959951Z ##[endgroup]
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.8010631Z 🚨 Enforcing final security gates
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.8012321Z ❌ SECURITY GATES FAILED
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.8014458Z 🚨 Critical security issues detected - deployment blocked
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.8016016Z 🔧 Fix all security vulnerabilities before proceeding
Security Policy Compliance	Enforce security gates	2025-08-03T18:58:28.8025697Z ##[error]Process completed with exit code 1.
