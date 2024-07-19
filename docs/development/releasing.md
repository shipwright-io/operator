# Release Process

This document outlines the steps required to release the operator. This document assumes that you
have achieved the "Approver"/"Maintainer" status, and have permission to manually trigger GitHub
Actions on this repo.

## Release Candidates (`X.Y.0-rcN`)

### Step 0: Set Up the Release Branch

Run the "Create Release Branch" GitHub Action, providing a valid Semantic Version with a major and
minor revision number. Example: `v0.13`.

### Step 1: Create a Release Candidate

This is ideal for `.0` releases, where there is most risk of "blocking" bugs. Release candidates
can be skipped for z-stream releases.

Run the "Release" GitHub action, with the following parameters:

- New tag: provide a semantic version _without the leading `v`_. Use a trailing `-` to add patch
  suffix. Ex: 0.13.0-rc0.
- Previous tag: provide a semantic version _that matches a tag on the GitHub repo_. Ex: `v0.12.0`.
- Ref: use the branch for the release in question. Ex: `release-v0.13`.

This will draft a release for GitHub - it will not publish a tag or release note.

### Step 2: Publish Draft Release

Find the draft release. Edit the release as follows:

- Add a leading `v` suffix to the generated tag and release name. Ex: `0.13.0-rc0` becomes `v0.13.0-rc0`.
- Change the "Draft release notes" title to "What's Changed".

Once you're happy, publish the draft release note. If an item is missing from the release note,
review the pull requests that should be included. Each desired PR should have:

1. A release note.
2. A `kind/*` label, such as `kind/feature` or `kind/bug`.

### Step 3: Verify the Bundle

Once the release candidate is published, broadcast the candidate build to the community in Slack
and the shipwright-dev mailing list. Refer to the "Testing a Release" section of the
[OLM Development](./olm-development.md#testing-a-release) guide for instructions on how to test the
release candidate.

### Step 4: Triage/Fix bugs

Once the release candidate is published, the community should file any issues as bugs in GitHub. It
is up to maintainers to determine which bugs are "release blockers" and which ones can be addressed
in a future release.

### Step 5: Repeat Release Candidates

Repeat steps 1-4 as needed if a "release blocker" issue is identified.

## Official Releases

Before proceeding with an official release, ensure that the 
[release branch](#step-0-set-up-the-release-branch) has been set up.

### Step 1: Bump versions for release

Once bugs have been triaged and the community feels ready, submit pull requests to bump the
`VERSION` make variable. For a new release, there should be two pull requests:

1. One for the `main` branch to update the minor semantic version. Ex: Update `0.12.0-rc0` to
   `0.13.0-rc0`
2. One for the `release-v*` branch, dropping any release candidate patch suffixes and/or updating
   the patch semantic version itself. Ex: `0.13.0` to `0.13.1`.

In both cases, run `make bundle` and commit any generated changes as part of the pull request.
Work with the community to get these pull requests merged.

### Step 2: Publish the Release

Repeat the process in [Step 1](#step-1-create-a-release-candidate) and
[Step 2](#step-2-publish-draft-release) above to create the release. For an official release, the
version should adhere to the `X.Y.Z` format (not extra dashes).
