# Renovate automation design

## Goal

Run Renovate on `main` as a manual pilot, create candidate branches, and open
pull requests only after those branches pass repository validation.

## Workflows

The automation uses two GitHub Actions workflows:

1. `renovate.yaml` runs Renovate manually and creates candidate branches.
2. `renovate-pr-opener.yaml` validates candidate branches, opens pull requests after successful validation, and deletes rejected branches when no pull request exists.

## Branch lifecycle

1. The Renovate workflow is triggered manually on `main`.
2. Renovate creates `renovate/**` branches but does not open pull requests;
   PR creation is held for Dependency Dashboard approval, while the opener
   workflow creates the validated pull request directly.
3. The PR opener workflow validates the pushed `renovate/**` branch and checks
   whether an open pull request already exists.
4. If validation succeeds and no open pull request already exists, the opener
   workflow opens the pull request.
5. If validation fails, the opener workflow deletes the branch unless an open
   pull request already exists.

The state of an existing pull request is handled as follows:

- If an open pull request already exists for a Renovate branch, the PR opener
  still validates pushes to that branch but does not create a duplicate pull
  request.
- If a security pull request is manually closed and its branch was deleted,
  the next Renovate run creates a fresh branch, which restarts validation.
- If a security pull request is manually closed but its branch remains and is
  conflicted, `rebaseWhen: conflicted` causes Renovate to refresh that branch,
  which restarts validation.
- `vulnerabilityAlerts.recreateWhen: always` allows Renovate to propose the
  same closed security update again instead of treating it as permanently
  dismissed.

## Validation

All Renovate branches are checked with:

- `make vendor`
- `make check-go-version`
- a clean Git tree after vendoring

Security branches use the `renovate/security/**` prefix and additionally run
`govulncheck` against both:

- the merge base with `main`
- the candidate branch

`renovate-security-gate` parses both raw `govulncheck` reports and reports two separate checks:

1. **Govulncheck sanity check**
   - fails if the baseline does not have a reachable vulnerability that the candidate fixes, logging that the Renovate alert is a false alarm
   - fails if the candidate introduces a reachable vulnerability
   - passes only when the candidate removes at least one reachable vulnerability and introduces none
   - Renovate does not expose the triggering CVE ID in branch metadata, so the gate verifies removal of at least one reachable baseline vulnerability
2. **CVSS threshold check**
   - passes when at least one fixed CVE is at or above the threshold, or when its CVSS score is unclear
   - fails when every fixed CVE has a known score below the threshold
   - records every fixed CVE and its CVSS result, including unclear results, in the pull-request description

The threshold follows Red Hat Product Security’s guidance for vulnerabilities with CVSS scores of `7.0` and above.

If either check fails, the PR opener workflow deletes the rejected branch when no pull request is already open.

## Debug artifacts

The Renovate workflow uploads a debug log and the effective configuration file.
The security validation uploads its console log, the raw baseline and candidate
`govulncheck` JSON, and the Markdown CVSS summary. Artifacts are retained for
seven days.

## Renovate configuration

`renovate.json` targets only `main`, enables the Go module manager, and uses
`prCreation: approval`. Vulnerability alerts use the `security/` branch prefix
and the same PR-creation policy. This prevents Renovate from racing the opener
workflow: Renovate only creates candidate branches, while the opener creates
pull requests after validation succeeds. Branch topics include the dependency
and target version. Renovate rebases an existing branch only when it is
conflicted, and closed security pull requests are recreated on later Renovate
runs.

## Authentication

The workflow requires a `RENOVATE_TOKEN` repository secret backed by a
personal access token with repository write access. GitHub App installation
tokens are not supported because they expire after one hour. The default
`GITHUB_TOKEN` cannot be used because pushes created with it do not trigger the
PR opener workflow.

The workflow runs only through GitHub's `workflow_dispatch` trigger during the
pilot phase. After the workflow is available on `main`, it can be run from the
Actions tab with the same token requirement. A weekly schedule can be
re-enabled later by adding a `schedule` trigger once the pilot is validated.
