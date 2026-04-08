#!/bin/sh
set -e

export VAULT_ADDR="${VAULT_ADDR:-http://vault:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-admin-token}"
MAX_WAIT=10

i=0
until vault status > /dev/null 2>&1 || [ $? -eq 2 ] || [ $i -ge $MAX_WAIT ]; do
  [ $i -eq 0 ] && echo "Vault: подключение"
  sleep 1
  i=$((i + 1))
done

if [ $i -ge $MAX_WAIT ] && ! vault status > /dev/null 2>&1 && [ $? -ne 2 ]; then
  echo "Vault: ошибка тайм-аута ($MAX_WAIT сек)"
  exit 1
fi

vault kv put secret/subscription-hub @/secrets.json > /dev/null
echo "Vault: секреты обновлены"
