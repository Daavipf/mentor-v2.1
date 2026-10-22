package domain

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden: you do not have permission to access this resource")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrExamAlreadyFinished = errors.New("exam is already finished")
	ErrQuestionNotInExam   = errors.New("question is not part of this exam")
	ErrAlternativeNotFound = errors.New("selected alternative does not belong to this question")
	ErrUserEmailConflict   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
)

