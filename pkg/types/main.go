package types

import (
	"errors"
	"fmt"
	"image/color"
	"time"

	"github.com/kubevirt/ci-health/pkg/constants"
)

const (
	StatsAction Action = "stats"
	BatchAction Action = "batch"

	FetchMode = "fetch"
	PlotMode  = "plot"

	MergeQueueLengthMetric  Metric = "merge-queue-length"
	TimeToMergeMetric       Metric = "time-to-merge"
	RetestsToMergeMetric    Metric = "retests-to-merge"
	MergedPRsMetric         Metric = "merged-prs"
	MergedPRsNoRetestMetric Metric = "merged-prs-no-retest"
)

type Metric string

func (m Metric) IsValid() error {
	switch m {
	case MergeQueueLengthMetric, TimeToMergeMetric, RetestsToMergeMetric, MergedPRsMetric:
		return nil
	}
	return errors.New("Invalid MetricType value")
}

func (m Metric) ResultsName() string {
	switch m {
	case MergeQueueLengthMetric:
		return constants.MergeQueueLengthName
	case TimeToMergeMetric:
		return constants.TimeToMergeName
	case RetestsToMergeMetric:
		return constants.RetestsToMergeName
	case MergedPRsMetric:
		return constants.MergedPRsName
	default:
		return ""
	}
}

type Action string

func (a Action) IsValid() error {
	switch a {
	case StatsAction, BatchAction:
		return nil
	}
	return errors.New("Invalid Action value")
}

type Mode string

func (m Mode) IsValid() error {
	switch m {
	case FetchMode, PlotMode:
		return nil
	}
	return errors.New("Invalid BatchMode value")
}

type Options struct {
	Action

	Path            string
	TokenPath       string
	Source          string
	DataDays        int
	LogLevel        string
	RequestedAction Action

	// stats options
	TimeToMergeRedLevel          float64
	TimeToMergeYellowLevel       float64
	MergeQueueLengthRedLevel     float64
	MergeQueueLengthYellowLevel  float64
	RetestsToMergeYellowLevel    float64
	RetestsToMergeRedLevel       float64
	MergedPRsYellowLevel         float64
	MergedPRsRedLevel            float64
	MergedPRsNoRetestYellowLevel float64
	MergedPRsNoRetestRedLevel    float64

	// batch options
	Mode         string
	TargetMetric string
	StartDate    string
}

type Label struct {
	Name string
}

type LabeledEventFragment struct {
	CreatedAt  time.Time
	AddedLabel Label `graphql:"addedLabel:label"`
}

type UnlabeledEventFragment struct {
	CreatedAt    time.Time
	RemovedLabel Label `graphql:"removedLabel:label"`
}

type IssueCommentFragment struct {
	CreatedAt time.Time
	BodyText  string
}

type Commit struct {
	CommittedDate time.Time
}

type PullRequestCommitFragment struct {
	Commit Commit
}

type Actor struct {
	Login string
}

type BaseRefForcePushFragment struct {
	Actor     Actor
	CreatedAt time.Time
}

type HeadRefForcePushFragment struct {
	Actor     Actor
	CreatedAt time.Time
}

type TimelineItem struct {
	LabeledEventFragment      `graphql:"... on LabeledEvent"`
	UnlabeledEventFragment    `graphql:"... on UnlabeledEvent"`
	IssueCommentFragment      `graphql:"... on IssueComment"`
	PullRequestCommitFragment `graphql:"... on PullRequestCommit"`
	BaseRefForcePushFragment  `graphql:"... on BaseRefForcePushedEvent"`
	HeadRefForcePushFragment  `graphql:"... on HeadRefForcePushedEvent"`
}

type TimelineItems struct {
	Nodes []TimelineItem
}

type ChatopsPullRequestFragment struct {
	Number        int
	CreatedAt     time.Time
	MergedAt      time.Time
	TimelineItems `graphql:"timelineItems(first:100, itemTypes:[PULL_REQUEST_COMMIT, BASE_REF_FORCE_PUSHED_EVENT, HEAD_REF_FORCE_PUSHED_EVENT, ISSUE_COMMENT])"`
}

type MergeQueuePullRequestFragment struct {
	Number        int
	CreatedAt     time.Time
	MergedAt      time.Time
	TimelineItems `graphql:"timelineItems(first:100, itemTypes:[LABELED_EVENT, UNLABELED_EVENT])"`
}

type BarePullRequestFragment struct {
	Number    int
	CreatedAt time.Time
	MergedAt  time.Time
}

type BarePRList []struct {
	BarePullRequestFragment `graphql:"... on PullRequest"`
}

type MergeQueuePRList []struct {
	MergeQueuePullRequestFragment `graphql:"... on PullRequest"`
}

type ChatopsPRList []struct {
	ChatopsPullRequestFragment `graphql:"... on PullRequest"`
}

type Curve struct {
	X     []string
	Y     []float64
	Color color.RGBA
	Title string
}

type PlotData struct {
	Title      string
	XAxisLabel string
	YAxisLabel string
	Curves     []Curve
}

type PR struct {
	Number   int
	MergedAt string
}

type DataPoint struct {
	Value float64
	Date  *time.Time `json:",omitempty"`
	PRs   []PR       `json:",omitempty"`
}

// RunningAverageDataItem contains data information in the form of a running average.
// It contains the actual average value and the data points used to obtain it.
type RunningAverageDataItem struct {
	Avg        float64
	Std        float64
	Number     float64
	NoRetest   float64
	DataPoints []DataPoint
}

func (d *RunningAverageDataItem) String() string {
	return fmt.Sprintf(constants.BadgeDataFormat, d.Avg, d.Std)
}

func (d *RunningAverageDataItem) SimpleBadgeString() string {
	return fmt.Sprintf(constants.NoRetestBadgeDataFormat, d.NoRetest, d.Number)
}

// Results represents the data obtained from GitHub. It includes the source repo from which
// the data was obtained and the number of days back from the execution time included in the
// data.
type Results struct {
	EndDate  string
	DataDays int
	Source   string
	Data     map[string]RunningAverageDataItem
}
