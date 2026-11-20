package ctxdata

import "strings"

// UserType 用户类型
type UserType string

const (
	UserTypePlatform UserType = "platform" // 总网
	UserTypeOperator UserType = "operator" // 分站
	UserTypeAgent    UserType = "agent"    // 代理
)

// UserTypeFromString 从字符串解析 UserType
func UserTypeFromString(s string) UserType {
	return UserType(strings.ToLower(s))
}

// String 实现 UserType 的 String 方法
func (ut UserType) String() string {
	return string(ut)
}

// IsValid 判断是否是有效的用户类型
func (ut UserType) IsValid() bool {
	return ut == UserTypePlatform || ut == UserTypeOperator || ut == UserTypeAgent
}

// IsPlatform 判断是否是总网
func (ut UserType) IsPlatform() bool {
	return ut == UserTypePlatform
}

// IsOperator 判断是否是分站
func (ut UserType) IsOperator() bool {
	return ut == UserTypeOperator
}

// IsAgent 判断是否是代理
func (ut UserType) IsAgent() bool {
	return ut == UserTypeAgent
}

// TokenType 令牌类型
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"  // 访问令牌
	TokenTypeRefresh TokenType = "refresh" // 刷新令牌
	TokenTypePreview TokenType = "preview" // 预览令牌
)

// String 实现 TokenType 的 String 方法
func (tt TokenType) String() string {
	return string(tt)
}

// IsValid 判断是否是有效的令牌类型
func (tt TokenType) IsValid() bool {
	return tt == TokenTypeAccess || tt == TokenTypeRefresh || tt == TokenTypePreview
}

// IsWant 判断是否是期望的令牌类型
func (tt TokenType) IsWant(want TokenType) bool {
	return tt == want
}

// IsAccess 判断是否是访问令牌
func (tt TokenType) IsAccess() bool {
	return tt == TokenTypeAccess
}

// IsRefresh 判断是否是刷新令牌
func (tt TokenType) IsRefresh() bool {
	return tt == TokenTypeRefresh
}

// IsPreview 判断是否是预览令牌
func (tt TokenType) IsPreview() bool {
	return tt == TokenTypePreview
}
