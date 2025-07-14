package utils

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/fiamma-chain/fiamma-go-sdk/log"
)

type JWTConfig struct {
	SigningAlgorithm string        `yaml:"sa" json:"sa" default:"HS256"`
	Key              string        `yaml:"key" json:"key" default:"Fiamma.20241111"`
	PrivKeyFile      string        `yaml:"privKeyFile" json:"privKeyFile"`
	PubKeyFile       string        `yaml:"pubKeyFile" json:"pubKeyFile"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout" default:"30m"`
	MaxRefresh       time.Duration `yaml:"maxRefresh" json:"maxRefresh" default:"1h"`
	TokenLookup      string        `yaml:"tokenLookup" json:"tokenLookup" default:"header: jwt, header: Authorization, query: jwt, param: jwt, cookie: jwt"`
}

type JWTInfo struct {
	Token      string    `json:"token"`
	Expire     time.Time `json:"expire"`
	MaxRefresh time.Time `json:"maxRefresh"`
}

const (
	JWTKey        = "jwt"
	JWTExpire     = "jwt_expire"
	JWTTimeOrigin = "jwt_origin"
)

var (
	// ErrMissingSecretKey indicates Secret key is required
	ErrMissingSecretKey = errors.New("filed to get secret key")

	// ErrFailedTokenCreation indicates JWT Token failed to create, reason unknown
	ErrFailedTokenCreation = errors.New("failed to create JWT Token")

	// ErrExpiredToken indicates JWT token has expired. Can't refresh.
	ErrExpiredToken = errors.New("failed to get valid token")

	// ErrEmptyAuthHeader can be thrown if authing with a HTTP header, the Auth header needs to be set
	ErrEmptyAuthHeader = errors.New("failed to get auth header")

	// ErrEmptyQueryToken can be thrown if authing with URL Query, the query token variable is empty
	ErrEmptyQueryToken = errors.New("failed to get query token")

	// ErrEmptyCookieToken can be thrown if authing with a cookie, the token cookie is empty
	ErrEmptyCookieToken = errors.New("failed to get cookie token")

	// ErrEmptyParamToken can be thrown if authing with parameter in path, the parameter in path is empty
	ErrEmptyParamToken = errors.New("failed to get parameter token")

	// ErrInvalidSigningAlgorithm indicates signing algorithm is invalid, needs to be HS256, HS384, HS512, RS256, RS384 or RS512
	ErrInvalidSigningAlgorithm = errors.New("failed to get get signing algorithm")

	// ErrNoPrivKeyFile indicates that the given private key is unreadable
	ErrNoPrivKeyFile = errors.New("failed to get private key file")

	// ErrNoPubKeyFile indicates that the given public key is unreadable
	ErrNoPubKeyFile = errors.New("failed to get public key file")

	// ErrInvalidPrivKey indicates that the given private key is invalid
	ErrInvalidPrivKey = errors.New("failed to get private key")

	// ErrInvalidPubKey indicates the the given public key is invalid
	ErrInvalidPubKey = errors.New("failed to get public key")
)

type JWTHelper struct {
	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration
	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is TokenTime + MaxRefresh.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// signing algorithm - possible values are HS256, HS384, HS512, RS256, RS384 or RS512
	// Optional, default is HS256.
	SigningAlgorithm string
	// Secret key used for signing. Required.
	Key []byte
	// Private key file for asymmetric algorithms
	PrivKeyFile string
	// Public key file for asymmetric algorithms
	PubKeyFile string

	// Private key
	privKey *rsa.PrivateKey
	// Public key
	pubKey *rsa.PublicKey

	log *log.Logger
}

func NewJWTHelper(cfg JWTConfig) (*JWTHelper, error) {
	if cfg.Key == "" {
		cfg.Key = os.Getenv("JWT_HS256_KEY")
	}
	helper := &JWTHelper{
		Timeout:          cfg.Timeout,
		MaxRefresh:       cfg.MaxRefresh,
		TokenLookup:      cfg.TokenLookup,
		SigningAlgorithm: cfg.SigningAlgorithm,
		Key:              []byte(cfg.Key),
		PrivKeyFile:      cfg.PrivKeyFile,
		PubKeyFile:       cfg.PubKeyFile,
		log:              log.L().With(log.Any("plugin", "jwt_helper")),
	}
	if helper.usingPublicKeyAlgo() {
		err := helper.readKeys()
		if err != nil {
			return nil, err
		}
	} else if cfg.Key == "" {
		return nil, ErrMissingSecretKey
	}
	return helper, nil
}

func (j *JWTHelper) jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)
	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}
	return authHeader, nil
}

func (j *JWTHelper) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)
	if token == "" {
		return "", ErrEmptyQueryToken
	}
	return token, nil
}

func (j *JWTHelper) jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)
	if cookie == "" {
		return "", ErrEmptyCookieToken
	}
	return cookie, nil
}

func (j *JWTHelper) jwtFromParam(c *gin.Context, key string) (string, error) {
	token := c.Param(key)
	if token == "" {
		return "", ErrEmptyParamToken
	}
	return token, nil
}

// ParseToken parse jwt token from gin context
func (j *JWTHelper) parseToken(c *gin.Context) (*jwt.Token, error) {
	token, err := j.GetTokenString(c)
	if err != nil {
		return nil, err
	}

	jtoken, err := jwt.Parse(token, j.keyFunc)
	if err != nil {
		return nil, err
	}
	c.Set(JWTKey, token)
	return jtoken, nil
}

func (j *JWTHelper) keyFunc(t *jwt.Token) (interface{}, error) {
	if jwt.GetSigningMethod(j.SigningAlgorithm) != t.Method {
		return nil, ErrInvalidSigningAlgorithm
	}
	if j.usingPublicKeyAlgo() {
		return j.pubKey, nil
	}
	// save token string if valid
	return j.Key, nil
}

func (j *JWTHelper) readKeys() error {
	err := j.privateKey()
	if err != nil {
		return err
	}
	err = j.publicKey()
	if err != nil {
		return err
	}
	return nil
}

func (j *JWTHelper) privateKey() error {
	keyData, err := os.ReadFile(j.PrivKeyFile)
	if err != nil {
		return ErrNoPrivKeyFile
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPrivKey
	}
	j.privKey = key
	return nil
}

func (j *JWTHelper) publicKey() error {
	keyData, err := ioutil.ReadFile(j.PubKeyFile)
	if err != nil {
		return ErrNoPubKeyFile
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPubKey
	}
	j.pubKey = key
	return nil
}

func (j *JWTHelper) usingPublicKeyAlgo() bool {
	switch j.SigningAlgorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JWTHelper) GetTokenString(c *gin.Context) (string, error) {
	var token string
	var err error

	methods := strings.Split(j.TokenLookup, ",")
	for _, method := range methods {
		if len(token) > 0 {
			break
		}
		parts := strings.Split(strings.TrimSpace(method), ":")
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "header":
			token, err = j.jwtFromHeader(c, v)
			if err != nil {
				j.log.Debug("failed to get jwt from header")
			}
		case "query":
			token, err = j.jwtFromQuery(c, v)
			if err != nil {
				j.log.Debug("failed to get jwt from query")
			}
		case "cookie":
			token, err = j.jwtFromCookie(c, v)
			if err != nil {
				j.log.Debug("failed to get jwt from cookie")
			}
		case "param":
			token, err = j.jwtFromParam(c, v)
			if err != nil {
				j.log.Debug("failed to get jwt from param")
			}
		}
	}

	if err != nil {
		return "", err
	}
	return token, nil
}

func (j *JWTHelper) Generate(claims map[string]interface{}) (token string, expire time.Time, err error) {
	now := time.Now()
	expire = now.Add(j.Timeout)
	claims[JWTExpire] = expire.Unix()
	claims[JWTTimeOrigin] = now.Unix()

	builder := jwt.New(jwt.GetSigningMethod(j.SigningAlgorithm))
	data := builder.Claims.(jwt.MapClaims)
	for k, v := range claims {
		data[k] = v
	}
	if j.usingPublicKeyAlgo() {
		token, err = builder.SignedString(j.privKey)
	} else {
		token, err = builder.SignedString(j.Key)
	}
	return
}

func (j *JWTHelper) Refresh(c *gin.Context) (string, time.Time, error) {
	claims, err := j.CheckMaxRefreshAndParse(c)
	if err != nil {
		return "", time.Now(), err
	}
	return j.Generate(claims)
}

func (j *JWTHelper) CheckExpireAndParse(c *gin.Context) (map[string]interface{}, error) {
	return j.checkAndParse(c, j.Timeout)
}

func (j *JWTHelper) CheckMaxRefreshAndParse(c *gin.Context) (map[string]interface{}, error) {
	return j.checkAndParse(c, j.MaxRefresh)
}

func (j *JWTHelper) checkAndParse(c *gin.Context, offset time.Duration) (map[string]interface{}, error) {
	jtoken, err := j.parseToken(c)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, err
	}
	if jtoken == nil {
		return nil, ErrFailedTokenCreation
	}

	claims := jtoken.Claims.(jwt.MapClaims)
	origin := int64(claims[JWTTimeOrigin].(float64))
	if time.Since(time.Unix(origin, 0)) > offset {
		return nil, ErrExpiredToken
	}
	return claims, nil
}
