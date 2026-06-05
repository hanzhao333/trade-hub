package service

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zhanghanzhao/trade-hub/internal/model"
	"github.com/zhanghanzhao/trade-hub/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already exists")
)

type AuthService struct {
	users     *repository.UserRepo
	jwtSecret []byte
}

func NewAuthService(users *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret)}
}

func (s *AuthService) Register(email, password string) error {
	// []byte(password)：把密码字符窜转换为字节
	// bcrypt.DefaultCost：使用默认的加密成本（10）
	// bcrypt.GenerateFromPassword：生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.Create(&model.User{Email: email, PasswordHash: string(hash)}); err != nil {
		if isDuplicateKey(err) {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

func isDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func (s *AuthService) Login(email, password string) (string, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}
	// CompareHashAndPassword：将密码哈希PasswordHash解密，与前端传入的password进行比较
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	// 生成token
	return s.issueToken(u.ID, u.Email)
}

func (s *AuthService) GetProfile(userID uint) (*model.User, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return u, nil
}

// JWT 是 三段用 . 拼起来的字符串：

// eyJhbGci...   .   eyJzdWI...   .   SflKxwRJ...
//
//	Header          Payload         Signature
//	（头）           （载荷）          （签名）
func (s *AuthService) issueToken(userID uint, email string) (string, error) {
	// 就是一个 map[string]interface{} 的别名，用来装 JWT Payload
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	// 创建一个 待签名的 token 对象 t
	// SigningMethodHS256：用 HMAC-SHA256 对称算法
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// s.jwtSecret： 进程中存的环境变量jwtSecret
	return t.SignedString(s.jwtSecret)
	// Signature是根据Header、Payload算出来的，所以严格意义上
	// Signature是验证token是否合法的关键，
	// 只要Header、Payload被篡改，Signature就会不匹配
}

func (s *AuthService) ParseToken(tokenStr string) (uint, error) {
	// 解析 JWT 字符串；第二个参数是 keyFunc，用来提供验签密钥
	// t是token对象
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// 返回签发 token 时用的同一个密钥（HS256）
		return s.jwtSecret, nil
	})
	// 解析失败，或 token 已过期/签名无效
	if err != nil || !t.Valid {
		return 0, ErrInvalidCredentials
	}
	// 把 Claims 断言成 MapClaims（Login 里 issueToken 用的就是这种）
	// 不断言的话 userID := t.Claims["sub"]  // ❌ 编译不过，Claims 不是 map
	// claims是payload
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidCredentials
	}
	// 取出 sub（subject），Login 里写入的是 userID；JSON 数字默认是 float64
	// .()断言之后返回两个值 一个是转换后的值，
	// 一个是bool值，如果转换成功，则ok为true，否则为false
	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, ErrInvalidCredentials
	}
	// 转成 uint 返回给 middleware，供 handler 用 c.GetUint 取当前用户
	return uint(sub), nil
}
