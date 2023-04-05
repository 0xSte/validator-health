package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"validator-stats/pkg/beacon"
)

func main() {

	// read in from file
	pubkey := "0x8d49260f56c41cb36d2a9efb4033b4249b98138fd50968f9338d55302266f117007ae3a043f3ffa5fc1a9344dc1c0e35"

	client := beacon.NewClient(http.DefaultClient)

	pkErrors, err := getValidatorhealth(client, pubkey)
	if err != nil {
		log.Fatal(err)
	}

	str, _ := json.MarshalIndent(pkErrors, "", "  ")
	fmt.Println(string(str))
}

type Validatorhealth struct {
	Info       beacon.Validator
	Conditions map[string][]Condition
}

func getValidatorhealth(client *beacon.Client, pubkey string) (*Validatorhealth, error) {
	validator, err := client.GetValidator(context.Background(), pubkey)
	if err != nil {
		return nil, err
	}

	stats, err := client.GetValidatorStats(context.Background(), 90, validator.Data.Validatorindex)
	if err != nil {
		return nil, err
	}

	pkErrors := make(map[string][]Condition)

	for _, stat := range stats.Data {
		if stat.MissedBlocks != 0 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedBlocks,
				IssueType: missedBlock,
			})
		}
		if stat.MissedSync != 0 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedSync,
				IssueType: missedSync,
			})
		}
		if stat.MissedAttestations != 0 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.MissedAttestations,
				IssueType: missedAttestation,
			})
		}
		if stat.ProposerSlashings != 0 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.ProposerSlashings,
				IssueType: slashingProposer,
			})
		}
		if stat.AttesterSlashings != 0 {
			pkErrors[validator.Data.Pubkey] = append(pkErrors[validator.Data.Pubkey], Condition{
				Day:       stat.DayEnd,
				Count:     stat.AttesterSlashings,
				IssueType: slashingAttester,
			})
		}
	}
	return &Validatorhealth{
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
	Day       string
	Count     int
	IssueType IssueType
}
