package handler

const (
	// Общие
	msgInternalError = "Внутренняя ошибка сервера"
	msgInvalidFormat = "Неверный формат запроса"
	msgValidationErr = "Ошибка валидации"
	msgUnauthorized  = "Пользователь не авторизован или токен недействителен"

	// Профиль и Пользователи
	msgUserExists         = "Пользователь с таким email уже зарегистрирован"
	msgDeleteErr          = "Ошибка удаления"
	msgCannotDeleteOwner  = "Невозможно удалить профиль владельца организации"
	msgInvalidCredentials = "Неверный email или пароль"

	// Компании
	msgMissingTenantErr = "Отсутствует идентификатор компании (X-Tenant-ID)"
)
