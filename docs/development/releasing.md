# Release Process

This document outlines the steps required to release the operator. This document assumes that you
have achieved the "Approver"/"Maintainer" status, and have permission to manually trigger GitHub
Actions on this repo.

To release operator updates to the [k8s community operators](https://github.com/k8s-operatorhub/community-operators),
you must be listed as an approver in our [operator CI configuration](https://github.com/k8s-operatorhub/community-operators/blob/main/operators/shipwright-operator/ci.yaml)
or request approval from a listed GitHub user.

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

### Step 3 (if needed): Set up your machine to run the OperatorHub release script

The OperatorHub release script requires the following:

1. Fork the [k8s community-operators](https://github.com/k8s-operatorhub/community-operators)
   repository.
2. Clone your fork with the `origin` remote name. Be sure to set your name and email in your local
   `git` configuration.
3. Add the community operators repository as the `upstream` remote:

   ```sh
   $ git remote add upstream https://github.com/k8s-operatorhub/community-operators.git
   ```

4. Install the [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md)
   tool locally.

### Step 4: Update the Operator on OperatorHub.io

[OperatorHub.io](https://operatorhub.io) publishes a catalog of operators sourced from git. To add
a new operator version, we must submit a pull request to the appropriate git repository.

Run the script `./hack/release-operatorhub.sh` to create a new release branch in your fork. The
script defaults to submitting pull requests to the k8s-operatorhub/community-operators catalog, but
other catalogs with the same format are supported.

The script accepts the following environment variables:

- `OPERATORHUB_DIR`: directory where the operator catalog repository was cloned. 
- `VERSION`: version of the operator to submit for update. Do not include the leading `v` in the
   version name.
- `HUB_REPO`: Regular expression to match for the operator catalog. Defaults to
  `k8s-operatorhub\/community-operators` - be sure to escape special characters when overriding
  this value.

Once the script completes and pushes the branch to your fork, open a pull request against the
[community operators](https://github.com/k8s-operatorhub/community-operators) repository.

### Step 5 (optional): Update other operator catalogs

OperatorHub.io is not the only catalog that can be used to publish operators on Kubernetes
clusters. Community members can use the `release-operatorhub.sh` script to update any other catalog
that uses the OperatorHub file structure by providing appropriate environment variable overrides.

Maintainers are not required to submit updates to other operator catalogs.
