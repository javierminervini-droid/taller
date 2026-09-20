package services

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("not found")
	ErrClientNotFound     = errors.New("client not found")
	ErrTechNotLinked      = errors.New("technician not linked")
	ErrTechViewOrders     = errors.New("technician view orders only")
	ErrTechUpdateOrders   = errors.New("technician update orders only")
	ErrFollowupsAdmin     = errors.New("followups admin only")
	ErrSignToken          = errors.New("sign token failed")
	ErrBadRequest         = errors.New("bad request")
	ErrTemplateNotFound   = errors.New("template not found")
	ErrTechContactOrders  = errors.New("technician contact orders only")
)

type badRequestError struct{ msg string }

func (e badRequestError) Error() string { return e.msg }
func (e badRequestError) Unwrap() error { return ErrBadRequest }

func BadRequest(msg string) error { return badRequestError{msg: msg} }
