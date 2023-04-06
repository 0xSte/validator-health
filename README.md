# validator-stats

## Background
- Makes use of the beaconcha.in API
- Can read from a YML file of pubkeys, or can be configured to read from prometheus eth-prysm validator metrics if you have them centralised

## Getting Started

### File mode
- Create a file called pubkeys.yml
- Add the pubkeys you care about
- Set the appropriate env vars:
  - `RUN_MODE=file` 
  - `TIME_RANGE` default == 90 days

### Prometheus mode
- Set the appropriate env vars:
    - `RUN_MODE=prom`
    - `TIME_RANGE` default == 90 days
    - `PROM_USER` the user
    - `PROM_PASSWORD` the password
    - `PROM_ENDPOINT` should be the configured datasource fully qualified path e.g. https://prometheus.example.com/api/v1/prom/

### Running
- Run the go application either as a binary:
  - ./validator-stats 
- Or as a go application
  - go run cmd/main.go

### Evaluate out.csv
- This includes the following fields for ONLY validators which have "ISSUES"
  - pubkey
  - issue_type (the type of "issue)
  - count (the number of instances of the "issue")
  - timestamp (of the "issue")
  - status (the validator status)
  - withdrawal_credentials (Withdrawal creds)
- "Issues" are defined as one of the following:
  - missed_block
  - missed_attestation
  - missed_sync
  - slashing_attester
  - slashing_proposer
  - status_ (not active)
  - exited_ (not relevant)

### Evaluate info.csv
- This includes the following fields for ALL validators found
    - pubkey
    - status
    - withdrawal_address
    - slashed
    - name
    - index
    - timestamp (of the state snapshot)

  
- This has some brief summary info about each public key provided


## Gotchas
- blockcha.in has a 10 requests per minute Rate limit, if you have a lot of validators, this can take some time... you can upgrade this
