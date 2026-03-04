package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

// NotificationService 通知服务
// 支持企业微信通知（后续可扩展邮件、钉钉等）
type NotificationService struct {
	cfg    *config.Config
	log    *logger.Logger
	token  string
	expire time.Time
}

// NewNotificationService 创建通知服务
func NewNotificationService(cfg *config.Config, log *logger.Logger) *NotificationService {
	return &NotificationService{cfg: cfg, log: log}
}

// WeChatMessage 企业微信消息格式
type WeChatMessage struct {
	ToUser  string      `json:"touser"`
	MsgType string      `json:"msgtype"`
	AgentID string      `json:"agentid"`
	Text    *TextMsg    `json:"text,omitempty"`
	Markdown *MarkdownMsg `json:"markdown,omitempty"`
}

// TextMsg 文本消息
type TextMsg struct {
	Content string `json:"content"`
}

// MarkdownMsg Markdown 消息
type MarkdownMsg struct {
	Content string `json:"content"`
}

// SendInstanceReadyNotification 发送实例就绪通知
func (s *NotificationService) SendInstanceReadyNotification(username, instanceName, accessURL string) error {
	content := fmt.Sprintf(`## 🎉 你的 OpenClaw 实例已就绪！

> **实例名称**：%s  
> **访问地址**：[%s](%s)  
> **状态**：✅ 运行中

快去使用吧！如有问题请联系管理员。`, instanceName, accessURL, accessURL)

	return s.sendWeChatMarkdown(username, content)
}

// SendApprovalNotification 发送审批通知（通知管理员有新申请）
func (s *NotificationService) SendApprovalNotification(adminUsername, applicant, instanceName, spec, reason string) error {
	content := fmt.Sprintf(`## 📋 有新的 OpenClaw 申请待审批

> **申请人**：%s  
> **实例名称**：%s  
> **规格**：%s  
> **申请理由**：%s

请前往管理控制台审批。`, applicant, instanceName, spec, reason)

	return s.sendWeChatMarkdown(adminUsername, content)
}

// SendInstanceFailedNotification 发送实例创建失败通知
func (s *NotificationService) SendInstanceFailedNotification(username, instanceName, errMsg string) error {
	content := fmt.Sprintf(`## ⚠️ 你的 OpenClaw 实例创建失败

> **实例名称**：%s  
> **失败原因**：%s

请联系管理员排查问题。`, instanceName, errMsg)

	return s.sendWeChatMarkdown(username, content)
}

// SendRejectNotification 发送拒绝通知
func (s *NotificationService) SendRejectNotification(username, instanceName, reason string) error {
	content := fmt.Sprintf(`## ❌ 你的 OpenClaw 申请未通过

> **申请实例**：%s  
> **拒绝原因**：%s

如有疑问，请联系管理员重新申请。`, instanceName, reason)

	return s.sendWeChatMarkdown(username, content)
}

// getAccessToken 获取企业微信 access_token（带缓存）
func (s *NotificationService) getAccessToken() (string, error) {
	// 检查 token 是否还有效
	if s.token != "" && time.Now().Before(s.expire) {
		return s.token, nil
	}

	// 重新获取 token
	url := fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		s.cfg.WeChat.CorpID,
		s.cfg.WeChat.Secret,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取企业微信 token 失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("企业微信返回错误: %s", result.ErrMsg)
	}

	// 缓存 token（提前 5 分钟过期）
	s.token = result.AccessToken
	s.expire = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)

	return s.token, nil
}

// sendWeChatMarkdown 发送企业微信 Markdown 消息
func (s *NotificationService) sendWeChatMarkdown(toUser, content string) error {
	// 检查企业微信配置是否完整
	if s.cfg.WeChat.CorpID == "" || s.cfg.WeChat.Secret == "" {
		s.log.Warn("企业微信未配置，跳过通知")
		return nil
	}

	token, err := s.getAccessToken()
	if err != nil {
		return fmt.Errorf("获取企业微信 token 失败: %w", err)
	}

	msg := WeChatMessage{
		ToUser:  toUser,
		MsgType: "markdown",
		AgentID: s.cfg.WeChat.AgentID,
		Markdown: &MarkdownMsg{
			Content: content,
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("发送企业微信通知失败: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("企业微信发送失败: %s", result.ErrMsg)
	}

	return nil
}
