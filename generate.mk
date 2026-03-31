SCHEMA_DIR   = ./api/schemas
CONFIG_DIR   = ./api/config
GEN_DIR      = ./internal/http/gen
SWAGGER_FILE = ./api/spec.yaml
OAPI_CODEGEN = oapi-codegen

define gen_api
	@mkdir -p $(GEN_DIR)/$(1)api
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/$(1)/models.yaml $(SCHEMA_DIR)/$(1).yaml
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/$(1)/server.yaml $(SWAGGER_FILE)
	@echo "🚀 API $(1) готово"
endef

define gen_models
	@mkdir -p $(GEN_DIR)/$(1)
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/$(1)/models.yaml $(SCHEMA_DIR)/$(1).yaml
	@echo "📦 Модели $(1) готовы"
endef

.PHONY: generate clean

generate:
	@mkdir -p $(GEN_DIR)
	@$(call gen_models,common)
	@$(call gen_models,user)
	@$(call gen_api,auth)
	@$(call gen_api,profile)
	@$(call gen_api,tenant)
	@echo "✨ Все сгенерировано"

clean:
	@rm -rf $(GEN_DIR)
	@echo "🧹 Сгенерированный код удален"