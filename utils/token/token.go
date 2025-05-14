package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultKeyPath = "./keys"
	privateKeyFile = "ed25519.key"
	publicKeyFile  = "ed25519.pub"
)

// 鍵の初期化と取得
func getKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// 鍵ファイルのパスを環境変数から取得またはデフォルト値を使用
	keyPath := os.Getenv("KEY_PATH")
	if keyPath == "" {
		keyPath = defaultKeyPath
	}

	privateKeyPath := filepath.Join(keyPath, privateKeyFile)
	publicKeyPath := filepath.Join(keyPath, publicKeyFile)

	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(keyPath, 0700); err != nil {
		return nil, nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	// 秘密鍵ファイルが存在するか確認
	_, err := os.Stat(privateKeyPath)
	if os.IsNotExist(err) {
		// 鍵が存在しない場合は新規生成
		return generateAndSaveKeys(privateKeyPath, publicKeyPath)
	} else if err != nil {
		return nil, nil, fmt.Errorf("error checking key file: %w", err)
	}

	// 既存の鍵をファイルから読み込む
	return loadKeysFromFile(privateKeyPath, publicKeyPath)
}

// 新しい鍵ペアを生成して保存
func generateAndSaveKeys(privateKeyPath, publicKeyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	fmt.Println("Generating new ED25519 key pair...")

	// 新しい鍵ペアを生成
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// 秘密鍵をファイルに保存
	err = os.WriteFile(privateKeyPath, []byte(base64.StdEncoding.EncodeToString(privateKey)), 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save private key: %w", err)
	}

	// 公開鍵をファイルに保存
	err = os.WriteFile(publicKeyPath, []byte(base64.StdEncoding.EncodeToString(publicKey)), 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save public key: %w", err)
	}

	fmt.Println("New ED25519 key pair generated and saved successfully.")
	return privateKey, publicKey, nil
}

// ファイルから鍵を読み込む
func loadKeysFromFile(privateKeyPath, publicKeyPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// 秘密鍵を読み込む
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Base64デコード
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(string(privateKeyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// 秘密鍵を作成
	privateKey := ed25519.PrivateKey(privateKeyDecoded)

	// 公開鍵は秘密鍵から取得することもできるが、一応ファイルからも読み込む
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		// 公開鍵ファイルがない場合は秘密鍵から取得
		publicKey := privateKey.Public().(ed25519.PublicKey)
		return privateKey, publicKey, nil
	}

	// Base64デコード
	publicKeyDecoded, err := base64.StdEncoding.DecodeString(string(publicKeyBytes))
	if err != nil {
		// デコードエラーの場合も秘密鍵から取得
		publicKey := privateKey.Public().(ed25519.PublicKey)
		return privateKey, publicKey, nil
	}

	publicKey := ed25519.PublicKey(publicKeyDecoded)
	return privateKey, publicKey, nil
}

// GenerateToken 指定されたユーザーIDに基づいてJWTトークンを生成する
func GenerateToken(id uint) (string, error) {
	tokenLifespan, err := strconv.Atoi(os.Getenv("TOKEN_HOUR_LIFESPAN"))
	if err != nil {
		// デフォルト値を設定
		tokenLifespan = 24
	}

	// 秘密鍵を取得
	privateKey, _, err := getKeys()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = id
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(tokenLifespan)).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	return token.SignedString(privateKey)
}

// 以下は前回と同じ実装
func extractTokenString(c *gin.Context) string {
	bearToken := c.Request.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func parseToken(tokenString string) (*jwt.Token, error) {
	// 公開鍵を取得
	_, publicKey, err := getKeys()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 署名方法がEdDSAであることを確認
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// TokenValid トークンが有効かどうかを検証
func Valid(c *gin.Context) error {
	tokenString := extractTokenString(c)
	if tokenString == "" {
		return fmt.Errorf("no token provided")
	}

	token, err := parseToken(tokenString)
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

// ExtractTokenId トークンからユーザーIDを取得
func ExtractTokenId(c *gin.Context) (uint, error) {
	tokenString := extractTokenString(c)
	if tokenString == "" {
		return 0, fmt.Errorf("no token provided")
	}

	token, err := parseToken(tokenString)
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		userId, ok := claims["user_id"].(float64)
		if !ok {
			return 0, fmt.Errorf("invalid user_id in token")
		}
		return uint(userId), nil
	}

	return 0, fmt.Errorf("invalid token claims")
}
