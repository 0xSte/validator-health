package validator

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0xste/validator-stats/pkg/beacon"
	"github.com/0xste/validator-stats/pkg/prom"
)

type Client struct {
	requests     atomic.Int64
	promClient   *prom.Client
	beaconClient *beacon.Client
}

func NewClient(beaconClient *beacon.Client, promClient *prom.Client) *Client {
	if promClient != nil {
		return &Client{
			promClient:   promClient,
			beaconClient: beaconClient,
		}
	}
	return &Client{
		beaconClient: beaconClient,
	}
}

type Health struct {
	Info       beacon.Validator
	Conditions map[string][]Condition
}

func (c *Client) GetEstimatedDuration(items int) time.Duration {
	return c.beaconClient.GetEstimatedDuration(items)
}

func (c *Client) logStatus() {
	time.Sleep(time.Second * 30)
}
func (c *Client) GetValidatorHealth(pubkey string, lookback time.Duration) (*Health, error) {
	validator, err := c.beaconClient.GetValidator(context.Background(), pubkey)
	if err != nil {
		return &Health{
			Info: beacon.Validator{ // make a mock health object
				Status: "UNKNOWN",
				Data:   beacon.ValidatorData{},
			},
			Conditions: map[string][]Condition{
				pubkey: {{
					Day:       time.Now(),
					Count:     1,
					IssueType: "NOT_EXISTS",
				}},
			},
		}, err
	}

	stats, err := c.beaconClient.GetValidatorStats(context.Background(), 90, validator.Data.Validatorindex)
	if err != nil {
		return &Health{
			Info: *validator,
			Conditions: map[string][]Condition{
				pubkey: {Condition{
					Day:       time.Now(),
					Count:     1,
					IssueType: IssueType(err.Error()),
				}},
			},
		}, err
	}

	pkErrors := make(map[string][]Condition)

	timeThreshold := time.Now().Add(-lookback)

	for _, stat := range stats.Data {
		if stat.MissedBlocks != 0 && stat.DayEnd.After(timeThreshold) {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedBlocks,
				IssueType: missedBlock,
			})
		}
		if stat.MissedSync != 0 && stat.DayEnd.After(timeThreshold) {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedSync,
				IssueType: missedSync,
			})
		}
		if stat.MissedAttestations != 0 && stat.DayEnd.After(timeThreshold) {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedAttestations,
				IssueType: missedAttestation,
			})
		}
		if stat.ProposerSlashings != 0 && stat.DayEnd.After(timeThreshold) {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.ProposerSlashings,
				IssueType: slashingProposer,
			})
		}
		if stat.AttesterSlashings != 0 && stat.DayEnd.After(timeThreshold) {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.AttesterSlashings,
				IssueType: slashingAttester,
			})
		}

		// real-time stats
		if validator.Data.Status != "active_online" {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       time.Now(),
				Count:     1,
				IssueType: IssueType(fmt.Sprintf("status_%s", validator.Data.Status)),
			})
		}
		if validator.Data.Slashed {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       time.Now(),
				Count:     1,
				IssueType: "slashed",
			})
		}
		if validator.Data.Exitepoch < 180600 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       time.Now(),
				Count:     int(validator.Data.Exitepoch),
				IssueType: "exit_epoch",
			})
		}
	}
	return &Health{
		Info:       *validator,
		Conditions: pkErrors,
	}, nil
}

type IssueType string

const (
	missedBlock       IssueType = "missed_block"
	missedAttestation IssueType = "missed_attestation"
	missedSync        IssueType = "missed_sync"
	slashingAttester  IssueType = "slashing_attester"
	slashingProposer  IssueType = "slashing_propoer"
)

type Condition struct {
	Day       time.Time
	Count     int
	IssueType IssueType
}
