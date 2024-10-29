package v1

var (
	// ErrSuccess common errors
	ErrSuccess             = newError(0, "ok")
	ErrBadRequest          = newError(400, "Bad Request")
	ErrUnauthorized        = newError(401, "Unauthorized")
	ErrNotFound            = newError(404, "Not Found")
	ErrInternalServerError = newError(500, "Internal Server Error")

	// ErrEmailAlreadyUse more biz errors
	ErrEmailAlreadyUse = newError(1001, "The email is already in use.")

	// ErrUserAlreadyExist 已存在用户 不允许注册
	ErrUserAlreadyExist = newError(1002, "User already exists")
)
