package beacon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	beaconBaseUrl = "https://beaconcha.in"
)

type Client struct {
	rc *resty.Client
}

func NewClient(hc *http.Client) *Client {
	return &Client{resty.NewWithClient(hc).SetBaseURL(beaconBaseUrl)}
}

type Validator struct {
	Status string `json:"status"`
	Data   struct {
		Activationeligibilityepoch int     `json:"activationeligibilityepoch"`
		Activationepoch            int     `json:"activationepoch"`
		Balance                    int64   `json:"balance"`
		Effectivebalance           int64   `json:"effectivebalance"`
		Exitepoch                  float64 `json:"exitepoch"`
		Lastattestationslot        int     `json:"lastattestationslot"`
		Name                       string  `json:"name"`
		Pubkey                     string  `json:"pubkey"`
		Slashed                    bool    `json:"slashed"`
		Status                     string  `json:"status"`
		Validatorindex             int     `json:"validatorindex"`
		Withdrawableepoch          float64 `json:"withdrawableepoch"`
		Withdrawalcredentials      string  `json:"withdrawalcredentials"`
		TotalWithdrawals           int     `json:"total_withdrawals"`
	} `json:"data"`
}

func (c *Client) GetValidator(ctx context.Context, pubkeys ...string) (*Validator, error) {
	delimited := delimit(pubkeys, ",")
	resp, err := c.rc.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/api/v1/validator/%s", delimited))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response was %d", resp.StatusCode())
	}
	var validator Validator
	if err := json.Unmarshal(resp.Body(), &validator); err != nil {
		return nil, err
	}
	return &validator, nil
}

type Proposals struct {
	Data []struct {
		Attestationscount          int    `json:"attestationscount"`
		Attesterslashingscount     int    `json:"attesterslashingscount"`
		Blockroot                  string `json:"blockroot"`
		Depositscount              int    `json:"depositscount"`
		Epoch                      int    `json:"epoch"`
		Eth1DataBlockhash          string `json:"eth1data_blockhash"`
		Eth1DataDepositcount       int    `json:"eth1data_depositcount"`
		Eth1DataDepositroot        string `json:"eth1data_depositroot"`
		ExecBaseFeePerGas          int    `json:"exec_base_fee_per_gas"`
		ExecBlockHash              string `json:"exec_block_hash"`
		ExecBlockNumber            int    `json:"exec_block_number"`
		ExecExtraData              string `json:"exec_extra_data"`
		ExecFeeRecipient           string `json:"exec_fee_recipient"`
		ExecGasLimit               int    `json:"exec_gas_limit"`
		ExecGasUsed                int    `json:"exec_gas_used"`
		ExecLogsBloom              string `json:"exec_logs_bloom"`
		ExecParentHash             string `json:"exec_parent_hash"`
		ExecRandom                 string `json:"exec_random"`
		ExecReceiptsRoot           string `json:"exec_receipts_root"`
		ExecStateRoot              string `json:"exec_state_root"`
		ExecTimestamp              int    `json:"exec_timestamp"`
		ExecTransactionsCount      int    `json:"exec_transactions_count"`
		Graffiti                   string `json:"graffiti"`
		GraffitiText               string `json:"graffiti_text"`
		Parentroot                 string `json:"parentroot"`
		Proposer                   int    `json:"proposer"`
		Proposerslashingscount     int    `json:"proposerslashingscount"`
		Randaoreveal               string `json:"randaoreveal"`
		Signature                  string `json:"signature"`
		Slot                       int    `json:"slot"`
		Stateroot                  string `json:"stateroot"`
		Status                     string `json:"status"`
		SyncaggregateBits          string `json:"syncaggregate_bits"`
		SyncaggregateParticipation int    `json:"syncaggregate_participation"`
		SyncaggregateSignature     string `json:"syncaggregate_signature"`
		Voluntaryexitscount        int    `json:"voluntaryexitscount"`
	} `json:"data"`
	Status string `json:"status"`
}

func (c *Client) GetValidatorProposals(ctx context.Context, epoch string, pubkeys ...string) (*Proposals, error) {
	delimited := delimit(pubkeys, ",")
	resp, err := c.rc.R().
		SetContext(ctx).
		//SetQueryParam("epoch", epoch).
		Get(fmt.Sprintf("/api/v1/validator/%s/proposals", delimited))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response was %d", resp.StatusCode())
	}
	var proposals Proposals
	if err := json.Unmarshal(resp.Body(), &proposals); err != nil {
		return nil, err
	}
	return &proposals, nil
}

type Stats struct {
	Data []struct {
		AttesterSlashings     int    `json:"attester_slashings"`
		Day                   int    `json:"day"`
		DayEnd                string `json:"day_end"`
		DayStart              string `json:"day_start"`
		Deposits              int    `json:"deposits"`
		DepositsAmount        int    `json:"deposits_amount"`
		EndBalance            int    `json:"end_balance"`
		EndEffectiveBalance   int    `json:"end_effective_balance"`
		MaxBalance            int    `json:"max_balance"`
		MaxEffectiveBalance   int    `json:"max_effective_balance"`
		MinBalance            int    `json:"min_balance"`
		MinEffectiveBalance   int    `json:"min_effective_balance"`
		MissedAttestations    int    `json:"missed_attestations"`
		MissedBlocks          int    `json:"missed_blocks"`
		MissedSync            int    `json:"missed_sync"`
		OrphanedAttestations  int    `json:"orphaned_attestations"`
		OrphanedBlocks        int    `json:"orphaned_blocks"`
		OrphanedSync          int    `json:"orphaned_sync"`
		ParticipatedSync      int    `json:"participated_sync"`
		ProposedBlocks        int    `json:"proposed_blocks"`
		ProposerSlashings     int    `json:"proposer_slashings"`
		StartBalance          int    `json:"start_balance"`
		StartEffectiveBalance int    `json:"start_effective_balance"`
		Validatorindex        int    `json:"validatorindex"`
		Withdrawals           int    `json:"withdrawals"`
		WithdrawalsAmount     int    `json:"withdrawals_amount"`
	} `json:"data"`
	Status string `json:"status"`
}

func (c *Client) GetValidatorStats(ctx context.Context, days int, index int) (*Stats, error) {
	resp, err := c.rc.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/api/v1/validator/stats/%d", index))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("response was %d", resp.StatusCode())
	}
	var proposals Stats
	if err := json.Unmarshal(resp.Body(), &proposals); err != nil {
		return nil, err
	}
	return &proposals, nil
}

func delimit(s []string, delimiter string) string {
	var sb strings.Builder
	if len(s) == 1 {
		return s[0]
	}
	for i, val := range s {
		if i > 0 {
			sb.WriteString(delimiter)
		}
		sb.WriteString(val)
	}
	return sb.String()
}
