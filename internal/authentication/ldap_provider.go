package authentication

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xiehqing/hiauth/internal/configx"
	"github.com/xiehqing/hiauth/internal/db/entity"
)

const (
	ldapDefaultPort       = 389
	ldapDefaultTLSPort    = 636
	ldapDefaultTimeoutSec = 5
)

type LDAPConfig struct {
	URL                string `json:"url"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	UseTLS             bool   `json:"useTLS"`
	SkipTLSVerify      bool   `json:"skipTLSVerify"`
	UserDNTemplate     string `json:"userDNTemplate"`
	TimeoutSeconds     int    `json:"timeoutSeconds"`
	AllowEmptyPassword bool   `json:"allowEmptyPassword"`
}

type ldapAuthProvider struct {
	config *configx.Reader
}

func newLDAPAuthProvider(config *configx.Reader) *ldapAuthProvider {
	return &ldapAuthProvider{config: config}
}

func (p *ldapAuthProvider) Name() string {
	return authProviderLDAP
}

func (p *ldapAuthProvider) Authenticate(ctx context.Context, user *entity.User, password string) error {
	if user == nil {
		return ErrInvalidLogin
	}
	if strings.TrimSpace(password) == "" && !p.ldapConfigAllowsEmptyPassword(ctx) {
		return ErrInvalidLogin
	}

	cfg, err := p.readConfig(ctx)
	if err != nil {
		return err
	}
	bindDN := cfg.userDN(user.Username)
	if bindDN == "" {
		return fmt.Errorf("%w: LDAP 用户 DN 模板不能为空", ErrAuthConfig)
	}
	if err := ldapSimpleBind(ctx, cfg, bindDN, password); err != nil {
		return ErrInvalidLogin
	}
	return nil
}

func (p *ldapAuthProvider) ldapConfigAllowsEmptyPassword(ctx context.Context) bool {
	var cfg LDAPConfig
	if err := p.config.JSON(ctx, entity.SecurityAuthLDAPConfig, &cfg); err != nil {
		return false
	}
	return cfg.AllowEmptyPassword
}

func (p *ldapAuthProvider) readConfig(ctx context.Context) (*LDAPConfig, error) {
	var cfg LDAPConfig
	if err := p.config.JSON(ctx, entity.SecurityAuthLDAPConfig, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAuthConfig, err)
	}
	if err := cfg.normalize(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *LDAPConfig) normalize() error {
	if strings.TrimSpace(c.URL) != "" {
		parsed, err := url.Parse(strings.TrimSpace(c.URL))
		if err != nil {
			return fmt.Errorf("%w: LDAP URL 不合法", ErrAuthConfig)
		}
		c.Host = parsed.Hostname()
		if parsed.Port() != "" {
			port, err := strconv.Atoi(parsed.Port())
			if err != nil {
				return fmt.Errorf("%w: LDAP 端口不合法", ErrAuthConfig)
			}
			c.Port = port
		}
		if strings.EqualFold(parsed.Scheme, "ldaps") {
			c.UseTLS = true
		}
	}
	c.Host = strings.TrimSpace(c.Host)
	if c.Host == "" {
		return fmt.Errorf("%w: LDAP 地址不能为空", ErrAuthConfig)
	}
	if c.Port <= 0 {
		if c.UseTLS {
			c.Port = ldapDefaultTLSPort
		} else {
			c.Port = ldapDefaultPort
		}
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = ldapDefaultTimeoutSec
	}
	c.UserDNTemplate = strings.TrimSpace(c.UserDNTemplate)
	if c.UserDNTemplate == "" {
		return fmt.Errorf("%w: LDAP 用户 DN 模板不能为空", ErrAuthConfig)
	}
	return nil
}

func (c LDAPConfig) address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c LDAPConfig) timeout() time.Duration {
	return time.Duration(c.TimeoutSeconds) * time.Second
}

func (c LDAPConfig) userDN(username string) string {
	username = strings.TrimSpace(username)
	template := strings.TrimSpace(c.UserDNTemplate)
	switch {
	case strings.Contains(template, "{{username}}"):
		return strings.ReplaceAll(template, "{{username}}", username)
	case strings.Contains(template, "%s"):
		return fmt.Sprintf(template, username)
	default:
		return template
	}
}

func ldapSimpleBind(ctx context.Context, cfg *LDAPConfig, bindDN, password string) error {
	dialer := &net.Dialer{Timeout: cfg.timeout()}
	conn, err := dialer.DialContext(ctx, "tcp", cfg.address())
	if err != nil {
		return err
	}
	defer conn.Close()

	if cfg.UseTLS {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: cfg.SkipTLSVerify,
		})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		conn = tlsConn
	}

	deadline := time.Now().Add(cfg.timeout())
	_ = conn.SetDeadline(deadline)
	if err := ctxDeadline(ctx, conn); err != nil {
		return err
	}
	if _, err := conn.Write(ldapBindRequest(1, bindDN, password)); err != nil {
		return err
	}

	response, err := readLDAPPacket(conn)
	if err != nil {
		return err
	}
	if !ldapBindSuccess(response) {
		return ErrInvalidLogin
	}
	return nil
}

func ctxDeadline(ctx context.Context, conn net.Conn) error {
	if deadline, ok := ctx.Deadline(); ok {
		return conn.SetDeadline(deadline)
	}
	return nil
}

func ldapBindRequest(messageID int, bindDN, password string) []byte {
	body := append(berInteger(messageID), berApplication(0, append(append(berInteger(3), berOctetString(bindDN)...), berContextPrimitive(0, []byte(password))...))...)
	return berSequence(body)
}

func ldapBindSuccess(packet []byte) bool {
	idx := bytes.IndexByte(packet, 0x61)
	if idx < 0 || idx+2 >= len(packet) {
		return false
	}
	content, ok := berContent(packet[idx:])
	if !ok || len(content) < 3 || content[0] != 0x0a {
		return false
	}
	enumContent, ok := berContent(content)
	return ok && len(enumContent) == 1 && enumContent[0] == 0
}

func readLDAPPacket(r io.Reader) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	if header[0] != 0x30 {
		return nil, fmt.Errorf("invalid ldap packet")
	}
	length, lengthBytes, err := readBERLength(r, header[1])
	if err != nil {
		return nil, err
	}
	packet := append([]byte{header[0], header[1]}, lengthBytes...)
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return append(packet, body...), nil
}

func readBERLength(r io.Reader, first byte) (int, []byte, error) {
	if first&0x80 == 0 {
		return int(first), nil, nil
	}
	count := int(first & 0x7f)
	if count == 0 || count > 4 {
		return 0, nil, fmt.Errorf("invalid ber length")
	}
	buf := make([]byte, count)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, nil, err
	}
	length := 0
	for _, b := range buf {
		length = length<<8 + int(b)
	}
	return length, buf, nil
}

func berSequence(content []byte) []byte {
	return berTLV(0x30, content)
}

func berApplication(tag byte, content []byte) []byte {
	return berTLV(0x60+tag, content)
}

func berInteger(value int) []byte {
	return berTLV(0x02, []byte{byte(value)})
}

func berOctetString(value string) []byte {
	return berTLV(0x04, []byte(value))
}

func berContextPrimitive(tag byte, value []byte) []byte {
	return berTLV(0x80+tag, value)
}

func berTLV(tag byte, content []byte) []byte {
	result := []byte{tag}
	result = append(result, berLength(len(content))...)
	return append(result, content...)
}

func berLength(length int) []byte {
	if length < 0x80 {
		return []byte{byte(length)}
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(length))
	for len(buf) > 0 && buf[0] == 0 {
		buf = buf[1:]
	}
	return append([]byte{0x80 | byte(len(buf))}, buf...)
}

func berContent(packet []byte) ([]byte, bool) {
	if len(packet) < 2 {
		return nil, false
	}
	offset := 2
	length := int(packet[1])
	if packet[1]&0x80 != 0 {
		count := int(packet[1] & 0x7f)
		if count == 0 || len(packet) < 2+count {
			return nil, false
		}
		offset = 2 + count
		length = 0
		for _, b := range packet[2:offset] {
			length = length<<8 + int(b)
		}
	}
	if length < 0 || len(packet) < offset+length {
		return nil, false
	}
	return packet[offset : offset+length], true
}
