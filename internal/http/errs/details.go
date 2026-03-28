package errs

const (
	MsgExpectedStringDetailErr  = "ожидалась строка"
	MsgExpectedIntegerDetailErr = "ожидалось целое число"
	MsgExpectedNumberDetailErr  = "ожидалось число"
	MsgExpectedBooleanDetailErr = "ожидалось true/false"
	MsgExpectedArrayDetailErr   = "ожидался список"
	MsgExpectedObjectDetailErr  = "ожидался объект"

	MsgNotNullableDetailErr      = "поле не может быть пустым"
	MsgRequiredDetailErr         = "обязательное поле отсутствует"
	MsgEmptyNotAllowedDetailErr  = "значение не может быть пустым"
	MsgInvalidFormatDetailErr    = "некорректный формат данных"
	MsgInvalidValueDetailErr     = "выбрано недопустимое значение"
	MsgTooSmallDetailErr         = "слишком маленькое значение"
	MsgTooLargeDetailErr         = "слишком большое значение"
	MsgGenericInvalidDetailErr   = "некорректные данные"
	MsgInvalidStructureDetailErr = "некорректная структура запроса"
)
