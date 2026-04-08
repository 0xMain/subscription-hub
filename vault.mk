DC_RUN_VAULT = docker compose run --rm vault-init

sync:
	@echo "Vault: синхронизация"
	@$(DC_RUN_VAULT)