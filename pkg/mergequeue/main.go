// Package mergequeue provides primitives for querying data about GitHub merge
// queues managed by Prow.

// For a GitHub project which CI is managed by Prow, we define the merge queue
// as the list of Pull Requests that are ready to be merged at any given date.
// They have been already reviewed, approved, don't have any needs-* label and
// don't have any do-not-merge/* label. The package provides several functions
// to query data about the merge queue.
package mergequeue

import (
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fgimenez/ci-health/pkg/constants"
	"github.com/fgimenez/ci-health/pkg/gh"
	"github.com/fgimenez/ci-health/pkg/types"
)

var (
	zeroDate = time.Time{}
)

type Handler struct {
	client *gh.Client
}

func NewHandler(client *gh.Client) *Handler {
	return &Handler{
		client,
	}
}

// LengthAt returns the merge queue size for a given date and the PR numbers
// that were part of the queue at that date.
func (mq *Handler) LengthAt(date time.Time) (int, []types.PR, error) {
	prList, err := mq.client.OpenPRsAt(date)
	if err != nil {
		return 0, nil, err
	}

	result := 0
	prs := []types.PR{}
	for _, prItem := range prList {
		log.Debugf("LengthAt: calculating mq date entered for PR num %d from %s", prItem.Number, date)
		if DatePREntered(&prItem.MergeQueuePullRequestFragment, date) != zeroDate {
			result++
			pr := types.PR{
				Number:   prItem.Number,
				MergedAt: prItem.MergedAt.Format(constants.DateFormat),
			}
			prs = append(prs, pr)
		}
	}

	return result, prs, nil
}

// TimesToMerge returns a map with the duration each merged PR number took to
// land in the time frame between the given start and end dates.
func (mq *Handler) TimesToMerge(startDate, endDate time.Time) (map[types.PR]time.Duration, error) {
	prs, err := mq.client.MergedPRsBetween(startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := map[types.PR]time.Duration{}
	for _, prItem := range prs {
		log.Debugf("TimesToMerge: calculating mq date entered for PR num %d merged at %s", prItem.Number, prItem.MergedAt)
		mqStart := DatePREntered(&prItem.MergeQueuePullRequestFragment, prItem.MergedAt)
		if mqStart.Equal(zeroDate) {
			log.Debugf("TimesToMerge: Merge queue enter date not found for PR %d", prItem.Number)
		} else {
			pr := types.PR{
				Number:   prItem.Number,
				MergedAt: prItem.MergedAt.Format(constants.DateFormat),
			}
			result[pr] = prItem.MergedAt.Sub(mqStart)
		}
	}

	return result, nil
}

// MergedPRsBetween returns an slice of type.PR merged between the given dates.
func (mq *Handler) MergedPRsBetween(startDate, endDate time.Time) ([]types.PR, error) {
	mergedPRs, err := mq.client.MergedPRsBetween(startDate, endDate)
	if err != nil {
		return nil, err
	}

	prs := []types.PR{}

	for _, mergedPR := range mergedPRs {
		pr := types.PR{
			Number:   mergedPR.Number,
			MergedAt: mergedPR.MergedAt.Format(constants.DateFormat),
		}
		prs = append(prs, pr)
	}
	return prs, nil
}

// DatePREntered returns when a PR entered the merge queue before a
// given date, zero value date if it was not in the merge queue by that date.
func DatePREntered(pr *types.MergeQueuePullRequestFragment, date time.Time) time.Time {
	labelsAdded, labelsRemoved := createMapsFromEvents(pr, date)

	if !hasAllLabelsRequiredForMerge(labelsAdded, labelsRemoved) {
		return zeroDate
	}

	if hasAnyDoNotMergeLabel(labelsAdded, labelsRemoved) {
		return zeroDate
	}

	doNotMergeLabelRemoval := latestMomentWhenAnyDoNotMergeLabelWasRemoved(labelsRemoved)
	requiredForMergeLabelAddition := latestMomentWhenAnyLabelRequiredForMergeWasAdded(labelsAdded)

	if doNotMergeLabelRemoval.After(requiredForMergeLabelAddition) {
		return doNotMergeLabelRemoval
	}
	return requiredForMergeLabelAddition
}

func createMapsFromEvents(pr *types.MergeQueuePullRequestFragment, date time.Time) (labelsAdded map[string]time.Time, labelsRemoved map[string]time.Time) {
	labelsAdded = make(map[string]time.Time)
	labelsRemoved = make(map[string]time.Time)

	for _, timelineItem := range pr.TimelineItems.Nodes {
		if isLabeledEvent(timelineItem, date) {

			labelsAdded[timelineItem.LabeledEventFragment.AddedLabel.Name] = timelineItem.LabeledEventFragment.CreatedAt
			labelsRemoved[timelineItem.LabeledEventFragment.AddedLabel.Name] = zeroDate

		} else if isUnlabeledEvent(timelineItem, date) {

			labelsAdded[timelineItem.UnlabeledEventFragment.RemovedLabel.Name] = zeroDate
			labelsRemoved[timelineItem.UnlabeledEventFragment.RemovedLabel.Name] = timelineItem.UnlabeledEventFragment.CreatedAt

		}
	}
	return
}

func isLabeledEvent(timelineItem types.TimelineItem, date time.Time) bool {
	return timelineItem.LabeledEventFragment != types.LabeledEventFragment{} &&
		timelineItem.LabeledEventFragment.AddedLabel.Name != "" &&
		date.After(timelineItem.LabeledEventFragment.CreatedAt)
}

func isUnlabeledEvent(timelineItem types.TimelineItem, date time.Time) bool {
	return timelineItem.UnlabeledEventFragment != types.UnlabeledEventFragment{} &&
		timelineItem.UnlabeledEventFragment.RemovedLabel.Name != "" &&
		date.After(timelineItem.UnlabeledEventFragment.CreatedAt)
}

func hasAnyDoNotMergeLabel(labelsAdded map[string]time.Time, labelsRemoved map[string]time.Time) bool {
	for _, doNotMergeLabel := range constants.DoNotMergeLabels() {
		if !strings.Contains(doNotMergeLabel, "*") {
			if labelsAdded[doNotMergeLabel] != zeroDate && !labelsRemoved[doNotMergeLabel].After(labelsAdded[doNotMergeLabel]) {
				return true
			}
		} else {
			foundLabel, foundTime := firstMapEntryWithKeyMatchingPatternAndNonZeroDate(labelsAdded, doNotMergeLabel)
			if foundTime != zeroDate && !labelsRemoved[foundLabel].After(foundTime) {
				return true
			}
		}
	}
	return false
}

func latestMomentWhenAnyDoNotMergeLabelWasRemoved(labelsRemoved map[string]time.Time) time.Time {
	result := zeroDate
	for _, doNotMergeLabel := range constants.DoNotMergeLabels() {
		if !strings.Contains(doNotMergeLabel, "*") {
			if labelsRemoved[doNotMergeLabel] != zeroDate && labelsRemoved[doNotMergeLabel].After(result) {
				result = labelsRemoved[doNotMergeLabel]
			}
		} else {
			foundLabel, foundTime := firstMapEntryWithKeyMatchingPatternAndNonZeroDate(labelsRemoved, doNotMergeLabel)
			if foundTime != zeroDate && labelsRemoved[foundLabel].After(result) {
				result = foundTime
			}
		}
	}
	return result
}

func firstMapEntryWithKeyMatchingPatternAndNonZeroDate(labelsToTimes map[string]time.Time, pattern string) (foundLabel string, foundTime time.Time) {
	regex, err := regexp.Compile(strings.ReplaceAll(pattern, "*", ".*"))
	if err != nil {
		panic(err)
	}
	for key, value := range labelsToTimes {
		if !regex.MatchString(key) {
			continue
		}
		if value != zeroDate {
			return key, value
		}
	}
	return "", zeroDate
}

func hasAllLabelsRequiredForMerge(labelsAdded map[string]time.Time, labelsRemoved map[string]time.Time) bool {
	for _, requiredForMergeLabel := range constants.RequiredForMergeLabels() {
		if labelsAdded[requiredForMergeLabel] == zeroDate || labelsRemoved[requiredForMergeLabel] != zeroDate {
			return false
		}
	}
	return true
}

func latestMomentWhenAnyLabelRequiredForMergeWasAdded(labelsAdded map[string]time.Time) time.Time {
	result := zeroDate
	for _, requiredForMergeLabel := range constants.RequiredForMergeLabels() {
		if labelsAdded[requiredForMergeLabel] != zeroDate && labelsAdded[requiredForMergeLabel].After(result) {
			result = labelsAdded[requiredForMergeLabel]
		}
	}
	return result
}
