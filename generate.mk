ROOT_DIR=$(shell pwd)
SWAGGER_FILE=$(ROOT_DIR)/api/spec.yaml
CONFIG_DIR=$(ROOT_DIR)/api/config
GEN_DIR=$(ROOT_DIR)/internal/http/gen
OAPI_CODEGEN=oapi-codegen

.PHONY: generate generate-all generate-auth generate-profile generate-user generate-tenants generate-customers generate-invoices generate-invoice-items clean

# Основная цель для генерации всего
generate: generate-all

generate-all: generate-auth generate-profile generate-user generate-tenants generate-customers generate-invoices generate-invoice-items
	@echo "✅ Весь код сгенерирован"

generate-auth:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/auth.yaml $(SWAGGER_FILE)
	@echo "✅ Код аутентификации сгенерирован"

generate-profile:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/profile.yaml $(SWAGGER_FILE)
	@echo "✅ Код профиля сгенерирован"

generate-user:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/user.yaml $(SWAGGER_FILE)
	@echo "✅ Код пользователей сгенерирован"

generate-tenants:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/tenant.yaml $(SWAGGER_FILE)
	@echo "✅ Код компаний сгенерирован"

generate-customers:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/customer.yaml $(SWAGGER_FILE)
	@echo "✅ Код клиентов сгенерирован"

generate-invoices:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/invoice.yaml $(SWAGGER_FILE)
	@echo "✅ Код счетов сгенерирован"

generate-invoice-items:
	@$(OAPI_CODEGEN) -config $(CONFIG_DIR)/invoice_item.yaml $(SWAGGER_FILE)
	@echo "✅ Код позиций счетов сгенерирован"

clean:
	@rm -rf $(GEN_DIR)
	@echo "✅ Очистка завершена"