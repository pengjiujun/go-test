package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 1. 定义载荷结构 (CustomClaims)
// 建议继承 RegisteredClaims，这样包含了标准的 exp, iss, nbf 等字段
type CustomClaims struct {
	UserID int64 `json:"id"`
	jwt.RegisteredClaims
}

// 2. 定义 JWT 工具类
type JWT struct {
	SigningKey []byte // 签名密钥
	Issuer     string // 签发者
	Expire     int64  // 过期时间(秒)
}

var (
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenNotValidYet = errors.New("token not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

// NewJWT 初始化函数
func NewJWT(secret, issuer string, expireSeconds int64) *JWT {
	return &JWT{
		SigningKey: []byte(secret),
		Issuer:     issuer,
		Expire:     expireSeconds,
	}
}

// CreateToken 生成 Token
func (j *JWT) CreateToken(userID int64) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.Expire) * time.Second)), // 过期时间
			Issuer:    j.Issuer,                                                                  // 签发人
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                            // 签发时间
		},
	}

	// 使用 HS256 签名算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 解析 Token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	// 错误处理工程化：将晦涩的库错误转换为业务错误
	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		} else {
			return nil, ErrTokenInvalid
		}
	}

	// 验证 Claims 类型
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}
