package prom

import (
	"context"
	"log"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// QueryInstant is shorthand for querying
func (c *Client) QueryInstant(ctx context.Context, q string) (model.Value, error) {
	result, warnings, err := v1.NewAPI(c.c).Query(ctx, q, time.Now())
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("warnings querying prometheus", zap.Strings("warnings", warnings))
	}

	return result, nil
}

// QueryRange queries a range
func (c *Client) QueryRange(ctx context.Context, q string, start, end time.Time, step time.Duration) (model.Value, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	result, warnings, err := v1.NewAPI(c.c).QueryRange(ctx, q, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("warnings querying prometheus", zap.Strings("warnings", warnings))
	}

	return result, nil
}

func (c *Client) GetValidatorPubkeys(ctx context.Context) ([]string, error) {
	resp, err := c.QueryInstant(ctx, `validator_statuses{pubkey!="", node_network="mainnet"} != 0`) // don't check inactive validators
	if err != nil {
		return nil, err
	}

	var pubkeys []string
	validatorVec := resp.(model.Vector)
	for _, v := range validatorVec {
		for name, value := range v.Metric {
			if string(name) == "pubkey" {
				pubkeys = append(pubkeys, string(value))
			}
		}
	}
	return pubkeys, nil
}
