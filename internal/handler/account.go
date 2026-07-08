package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lancer/log/internal/auth"
	"github.com/lancer/log/internal/mailer"
	"github.com/lancer/log/internal/model"
	"github.com/lancer/log/internal/repo"
)

const passwordResetPublicMessage = "如果账号已配置邮箱，验证码已发送"

type accountResp struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	DisplayName       string    `json:"display_name"`
	RecoveryEmail     string    `json:"recovery_email"`
	HasRecoveryEmail  bool      `json:"has_recovery_email"`
	PasswordUpdatedAt time.Time `json:"password_updated_at"`
}

func (h *APIHandler) Account(c *gin.Context) {
	u, err := repo.GetUserByID(c.Request.Context(), h.DB.Pool, currentUserID(c))
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 404, "账号不存在")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, accountFromUser(u))
}

type recoveryEmailReq struct {
	Email string `json:"email"`
}

func (h *APIHandler) UpdateRecoveryEmail(c *gin.Context) {
	var r recoveryEmailReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, "bad request")
		return
	}
	email, err := normalizeEmail(r.Email)
	if err != nil {
		fail(c, 400, "邮箱格式不正确")
		return
	}
	if err := repo.UpdateUserRecoveryEmail(c.Request.Context(), h.DB.Pool, currentUserID(c), email); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			fail(c, 404, "账号不存在")
			return
		}
		fail(c, 500, err.Error())
		return
	}
	h.Account(c)
}

type updatePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *APIHandler) UpdateAccountPassword(c *gin.Context) {
	var r updatePasswordReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, "bad request")
		return
	}
	if err := validateNewPassword(r.NewPassword); err != nil {
		fail(c, 400, err.Error())
		return
	}
	u, err := repo.GetUserByID(c.Request.Context(), h.DB.Pool, currentUserID(c))
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if !auth.CheckPassword(u.PasswordHash, r.CurrentPassword) {
		fail(c, 400, "当前密码不正确")
		return
	}
	hash, err := auth.HashPassword(r.NewPassword)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := repo.UpdateUserPassword(c.Request.Context(), h.DB.Pool, u.ID, hash); err != nil {
		fail(c, 500, err.Error())
		return
	}
	_ = repo.InvalidatePasswordResetCodes(c.Request.Context(), h.DB.Pool, u.ID)
	ok(c, gin.H{"saved": true})
}

type passwordResetRequestReq struct {
	Username string `json:"username"`
}

func (h *APIHandler) RequestPasswordReset(c *gin.Context) {
	if !h.Config.SMTP.Ready() {
		fail(c, 503, "邮件服务未配置")
		return
	}
	var r passwordResetRequestReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, "bad request")
		return
	}
	username := strings.TrimSpace(r.Username)
	if username == "" {
		fail(c, 400, "请输入用户名")
		return
	}
	u, err := repo.GetUserByUsername(c.Request.Context(), h.DB.Pool, username)
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 404, "用户不存在")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if strings.TrimSpace(u.RecoveryEmail) == "" {
		fail(c, 400, "该用户未配置找回邮箱")
		return
	}
	recent, err := repo.CountRecentPasswordResetCodes(c.Request.Context(), h.DB.Pool, u.ID, time.Now().Add(-15*time.Minute))
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if recent >= 3 {
		fail(c, 429, "请求过于频繁，请 15 分钟后再试")
		return
	}
	code, err := randomResetCode()
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	codeHash, err := auth.HashPassword(code)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := repo.CreatePasswordResetCode(c.Request.Context(), h.DB.Pool, u.ID, codeHash, time.Now().Add(10*time.Minute), c.ClientIP()); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := mailer.SendPasswordResetCode(h.Config.SMTP, u.RecoveryEmail, h.brandName(c), code); err != nil {
		fmt.Printf("smtp send error: %v\n", err)
		fail(c, 500, "验证码发送失败: "+err.Error())
		return
	}
	ok(c, gin.H{"message": passwordResetPublicMessage})
}

type passwordResetConfirmReq struct {
	Username    string `json:"username"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *APIHandler) ConfirmPasswordReset(c *gin.Context) {
	var r passwordResetConfirmReq
	if err := c.ShouldBindJSON(&r); err != nil {
		fail(c, 400, "bad request")
		return
	}
	if err := validateNewPassword(r.NewPassword); err != nil {
		fail(c, 400, err.Error())
		return
	}
	u, err := repo.GetUserByUsername(c.Request.Context(), h.DB.Pool, strings.TrimSpace(r.Username))
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 400, "验证码无效或已过期")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	reset, err := repo.LatestActivePasswordResetCode(c.Request.Context(), h.DB.Pool, u.ID)
	if errors.Is(err, repo.ErrNotFound) {
		fail(c, 400, "验证码无效或已过期")
		return
	}
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	code := strings.TrimSpace(r.Code)
	if code == "" || !auth.CheckPassword(reset.CodeHash, code) {
		_ = repo.IncrementPasswordResetAttempts(c.Request.Context(), h.DB.Pool, reset.ID)
		fail(c, 400, "验证码无效或已过期")
		return
	}
	hash, err := auth.HashPassword(r.NewPassword)
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	tx, err := h.DB.Pool.Begin(c.Request.Context())
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	defer tx.Rollback(c.Request.Context())
	if err := repo.UpdateUserPassword(c.Request.Context(), tx, u.ID, hash); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := repo.UsePasswordResetCode(c.Request.Context(), tx, reset.ID); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := repo.InvalidatePasswordResetCodes(c.Request.Context(), tx, u.ID); err != nil {
		fail(c, 500, err.Error())
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, gin.H{"saved": true})
}

func currentUserID(c *gin.Context) uuid.UUID {
	return c.MustGet("uid").(uuid.UUID)
}

func accountFromUser(u model.User) accountResp {
	return accountResp{
		ID:                u.ID.String(),
		Username:          u.Username,
		DisplayName:       u.DisplayName,
		RecoveryEmail:     u.RecoveryEmail,
		HasRecoveryEmail:  strings.TrimSpace(u.RecoveryEmail) != "",
		PasswordUpdatedAt: u.PasswordUpdatedAt,
	}
}

func normalizeEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Address == "" || strings.Contains(addr.Address, " ") {
		return "", fmt.Errorf("invalid email")
	}
	return strings.ToLower(addr.Address), nil
}

func validateNewPassword(pw string) error {
	if len([]rune(pw)) < 8 {
		return fmt.Errorf("新密码至少需要 8 位")
	}
	return nil
}

func randomResetCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (h *APIHandler) brandName(c *gin.Context) string {
	var raw []byte
	err := h.DB.Pool.QueryRow(c.Request.Context(), `SELECT value::text FROM settings WHERE section_key='branding'`).Scan(&raw)
	if err != nil {
		return "lancer.log"
	}
	var v struct {
		Brand string `json:"brand"`
	}
	if json.Unmarshal(raw, &v) != nil || strings.TrimSpace(v.Brand) == "" {
		return "lancer.log"
	}
	return v.Brand
}
