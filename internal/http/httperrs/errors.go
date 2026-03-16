package httperrs

const (
	MsgInternalErr      = "Внутренняя ошибка сервера"
	MsgInvalidFormatErr = "Неверный формат запроса"
	MsgValidationErr    = "Ошибка валидации"
	MsgNotFoundErr      = "Ресурс не найден"

	MsgUnauthorizedErr      = "Требуется авторизация"
	MsgInvalidTokenErr      = "Невалидный или просроченный токен"
	MsgMissingAuthHeaderErr = "Отсутствует заголовок авторизации"
	MsgInvalidUserIDErr     = "Некорректный идентификатор пользователя в токене"

	MsgForbiddenErr       = "У вас недостаточно прав для выполнения этого действия"
	MsgAccessDeniedErr    = "Доступ запрещен"
	MsgUserNotInTenantErr = "Пользователь не принадлежит данной организации"

	MsgUserExistsErr         = "Пользователь с таким email уже зарегистрирован"
	MsgUserNotFoundErr       = "Пользователь не найден"
	MsgDeleteErr             = "Ошибка удаления"
	MsgCannotDeleteOwnerErr  = "Невозможно удалить профиль владельца организации"
	MsgInvalidCredentialsErr = "Неверный email или пароль"

	MsgMissingTenantErr   = "Отсутствует идентификатор компании (X-Tenant-ID)"
	MsgInvalidTenantIDErr = "Некорректный идентификатор компании"
	MsgTenantNotFoundErr  = "Компания не найдена"
)
