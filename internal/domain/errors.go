package domain

import "errors"

var (
	ErrInvalidToken       = errors.New("невалидный токен")
	ErrAccessDenied       = errors.New("недостаточно прав")
	ErrInvalidCredentials = errors.New("неверный email или пароль")

	ErrIllegalOwnerAssignment = errors.New("назначение владельца возможно только через процедуру передачи прав")
	ErrCannotRemoveOwner      = errors.New("нельзя удалить владельца через управление персоналом")
	ErrUserAlreadyRegistered  = errors.New("пользователь с таким email уже зарегистрирован")
	ErrUserNotFound           = errors.New("пользователь не найден")
	ErrUserNotInTenant        = errors.New("пользователь не состоит в компании")
	ErrUserAlreadyInTenant    = errors.New("пользователь уже состоит в компании")

	ErrOwnerAlreadyExists = errors.New("у компании уже есть владелец")

	ErrTenantNotFound      = errors.New("компания не найдена")
	ErrTenantAlreadyExists = errors.New("компания с таким названием уже существует")

	ErrCannotDeleteOwner = errors.New("невозможно удалить учетную запись: вы являетесь владельцем одной или нескольких компаний. Сначала передайте права владения или удалите компании")
)
