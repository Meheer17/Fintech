#!/bin/sh
# revenueiq_infrastructure/vault/init.sh
#
# This script initializes the HashiCorp Vault KV secret store with default development secrets
# and demonstrates how microservices in RevenueIQ can query secrets via Vault's HTTP API.

set -e

# Configuration
VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-my-root-token}"

echo "========================================================="
echo " RevenueIQ - HashiCorp Vault Initializer & Guide"
echo "========================================================="
echo "Vault Address: $VAULT_ADDR"
echo "Vault Token  : $VAULT_TOKEN"
echo "---------------------------------------------------------"

# Helper function to check if vault container is ready
wait_for_vault() {
  echo "Checking if Vault is reachable..."
  until curl -s -f -o /dev/null "$VAULT_ADDR/v1/sys/health"; do
    echo "Waiting for Vault to start..."
    sleep 2
  done
  echo "Vault is up and healthy!"
}

# 1. Wait for Vault
wait_for_vault

# 2. Write Dev Secrets to KV (v2) Engine
# In development mode, the 'secret/' KV engine is enabled by default.
# We store the secrets under the path 'secret/data/revenueiq-dynamics'
echo "Writing development secrets to Vault..."
curl -s \
  --header "X-Vault-Token: $VAULT_TOKEN" \
  --request POST \
  --data '{
    "data": {
      "jwt_secret": "super-secret-jwt-key-change-this-in-production",
      "aws_access_key_id": "mock-access-key-id",
      "aws_secret_access_key": "mock-secret-access-key",
      "mongodb_uri": "mongodb://mongodb:27017"
    }
  }' \
  "$VAULT_ADDR/v1/secret/data/revenueiq-dynamics"

echo "Secrets stored successfully!"

# 3. Read secrets to verify
echo "---------------------------------------------------------"
echo "Verifying stored secrets:"
curl -s \
  --header "X-Vault-Token: $VAULT_TOKEN" \
  "$VAULT_ADDR/v1/secret/data/revenueiq-dynamics" | jq .data.data || curl -s --header "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/secret/data/revenueiq-dynamics"

echo ""
echo "========================================================="
echo " HOW SERVICES SHOULD QUERY SECRETS FROM VAULT"
echo "========================================================="
echo "Services can retrieve these secrets using Vault's HTTP API."
echo ""
echo "Example HTTP GET request:"
echo "  GET $VAULT_ADDR/v1/secret/data/revenueiq-dynamics"
echo "  Headers: X-Vault-Token: <vault-token>"
echo ""
echo "Example curl command:"
echo "  curl -H \"X-Vault-Token: \$VAULT_TOKEN\" \$VAULT_ADDR/v1/secret/data/revenueiq-dynamics"
echo ""
echo "Example Go implementation snippet:"
echo '
package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type VaultSecretResponse struct {
	Data struct {
		Data map[string]interface{} `json:"data"`
	} `json:"data"`
}

func GetSecrets(vaultAddr, token string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/secret/data/revenueiq-dynamics", vaultAddr), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned non-200 status: %d", resp.StatusCode)
	}

	var secretResponse VaultSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&secretResponse); err != nil {
		return nil, err
	}

	return secretResponse.Data.Data, nil
}
'
echo "========================================================="
