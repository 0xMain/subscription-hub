package errs

const (
	MsgInternalErr     = "Внутренняя ошибка сервера"
	MsgForbiddenErr    = "Недостаточно прав"
	MsgBadRequestErr   = "Ошибка обработки запроса"
	MsgValidationErr   = "Ошибка валидации"
	MsgUnauthorizedErr = "Требуется авторизация"

	MsgUserNotInTenantErr = "Пользователь не принадлежит данной организации"

	MsgUserExistsErr         = "Пользователь с таким email уже существует"
	MsgUserNotFoundErr       = "Пользователь не найден"
	MsgDeleteErr             = "Ошибка удаления"
	MsgInvalidCredentialsErr = "Неверный email или пароль"

	MsgMissingTenantErr   = "Отсутствует идентификатор компании (X-Tenant-ID)"
	MsgInvalidTenantIDErr = "Некорректный идентификатор компании"
	MsgTenantNotFoundErr  = "Компания не найдена"
)
