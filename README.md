# ci-health

![ci-health-tests](https://github.com/fgimenez/ci-health/workflows/ci-health-tests/badge.svg)

This repo contains code to calculate metrics about the performance of CI systems
based on Prow.

## Definitions

* Merge queue: list of Pull Requests that are ready to be merged at any given
date. For being ready to be merged they must:

  * Have the `lgtm` label.
  * Have the `approved` label.
  * Not have any label matching `do-not-merge/*`, i.e. `do-not-merge/hold`,  `do-not-merge/work-in-progress` etc. .
  * Not have any label matching `needs-*`, i.e. `needs-rebase`, `needs-ok-to-test` etc. .

* Merge queue length: number of PRs in the merge queue at a given time.
* Time to merge: for each merged PR, the time in days it took since it entered the merge
queue for the last time until it got finally merged.
* Retests to merge: for each merged PR, how many `/test` and `/retest` comments
were issued after the last code push.

## Status
This status is updated every 3 hours. The average values are calculated with
data from the previous 7 days since the execution time.

### kubevirt/kubevirt

![avg-merge-queue-lenght](https://fgimenez.github.io/ci-health/output/kubevirt/kubevirt/merge-queue-length.svg)
![avg-time-to-merge](https://fgimenez.github.io/ci-health/output/kubevirt/kubevirt/time-to-merge.svg)
![avg-retests-to-merge](https://fgimenez.github.io/ci-health/output/kubevirt/kubevirt/retests-to-merge.svg)

[Latest execution data](https://fgimenez.github.io/ci-health/output/kubevirt/kubevirt/results.json)

[Latest weeks data](https://grafana.ci.kubevirt.io/d/WZU1-LPGz/merge-queue)

## Historical data evolution

These plots will be updated every week.

* kubevirt/kubevirt merge queue length:
  ![kubevirt-kubevirt-merge-queue-length](./output/kubevirt/kubevirt/batch/merge-queue-length/plot/plot.png)

  Data avaialable [here](./output/kubevirt/kubevirt/batch/merge-queue-length/data).

* kubevirt/kubevirt time to merge:
  ![kubevirt-kubevirt-time-to-merge](./output/kubevirt/kubevirt/batch/time-to-merge/plot/plot.png)

  Data avaialable [here](./output/kubevirt/kubevirt/batch/time-to-merge/data).

* kubevirt/kubevirt retests to merge:
  ![kubevirt-kubevirt-retests-to-merge](./output/kubevirt/kubevirt/batch/retests-to-merge/plot/plot.png)

  Data avaialable [here](./output/kubevirt/kubevirt/batch/retests-to-merge/data).

## Local execution
You can execute the tool locally to grab the stats of an specific repo that uses
Prow, these are the requirements:

* [Bazelisk](https://github.com/bazelbuild/bazelisk)
* A GitHub token with `public_repo` permission, it is required because the tool
queries GitHub's API

A generic command execution from the repo's root looks like:
```
$ bazelisk run //cmd/stats -- --gh-token /path/to/token --source <org/repo> --path /path/to/output/dir --data-days <days-to-query>
```
where:
* `--gh-token`: should contain the path of the file where you saved your GitHub
token.
* `--source`: is the organization and repo to query information from.
* `--path`: is the path to store output data.
* `--data-days`: is the number of days to query.

You can check all the available options with:
```
$ bazelisk run //cmd/stats -- --help
```
So, for instance, if you have stored the path of your GitHub token file in a
`GITHUB_TOKEN` environment variable, a query for the last four days of
kubevirt/kubevirt can look like:
```
$ bazelisk run //cmd/stats -- --gh-token ${GITHUB_TOKEN} --source kubevirt/kubevirt --path /tmp/ci-health --data-days 7
```
