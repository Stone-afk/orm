package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrPointerOnly 只支持一级指针作为输入
	// 看到这个 error 说明你输入了其它的东西
	// 我们并不希望用户能够直接使用 err == ErrPointerOnly
	// 所以放在我们的 internal 包里
	ErrPointerOnly = errors.New("orm: 只支持一级指针作为输入，例如 &User{}")

	ErrInputNil               = errors.New("orm: 不支持 nil")
	ErrNoRows                 = errors.New("orm: 未找到数据")
	ErrTooManyReturnedColumns = errors.New("orm: 过多列")
)

func NewErrDuplicateAlias(alias string) error {
	return fmt.Errorf("orm: 重复别名 %s", alias)
}

// NewErrUnknownField 返回代表未知字段的错误
// 一般意味着你可能输入的是列名，或者输入了错误的字段名
func NewErrUnknownField(fdName string) error {
	return fmt.Errorf("orm: 未知字段 %s", fdName)
}

// NewErrUnsupportedExpressionType 返回一个不支持该 expression 错误信息
func NewErrUnsupportedExpressionType(exp any) error {
	return fmt.Errorf("orm: 不支持的表达式 %v", exp)
}

// 后面可以考虑支持错误码
// func NewErrUnsupportedExpressionType(exp any) error {
// 	return fmt.Errorf("orm-50001: 不支持的表达式 %v", exp)
// }

// 后面还可以考虑用 AST 分析源码，生成错误排除手册，例如
// @ErrUnsupportedExpressionType 40001
// 发生该错误，主要是因为传入了不支持的 Expression 的实际类型
// 一般来说，这是因为中间件

func NewErrInvalidTagContent(tag string) error {
	return fmt.Errorf("orm: 错误的标签设置: %s", tag)
}

// NewErrUnknownColumn 返回代表未知列的错误
// 一般意味着你使用了错误的列名
// 注意和 NewErrUnknownField 区别
func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm: 未知列 %s", name)
}

func NewErrUnsupportedSelectable(exp any) error {
	return fmt.Errorf("orm: 不支持的目标列 %v", exp)
}
