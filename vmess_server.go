package publisher

import (
	"encoding/base64"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// VMessServer represents a VMess server supported by v2rayN.
type VMessServer struct {
	// ConfigVersion represents the version of the current configuration file,
	// which is used by v2rayN to identify the current configuration.
	ConfigVersion string `cfg:"configVersion" json:"v"`

	// Remarks is the server alias, which is used for user-defined remarks.
	Remarks string `cfg:"remarks" json:"ps"`

	// Address is the server address.
	Address string `cfg:"address" json:"add"`

	// Port is the server port.
	Port string `cfg:"port" json:"port"`

	// ID is the user's id.
	ID string `cfg:"id" json:"id"`

	// AlterID is the user's alterId.
	AlterID string `cfg:"alterId" json:"aid"`

	// Network represents the transmission protocol of the underlying transport.
	// e.g. tcp, kcp, ws, h2, quic
	Network string `cfg:"network" json:"net"`

	// HeaderType represents the masquerade header of the underlying transport.
	// e.g. none, http, srtp, utp, wechat-video
	HeaderType string `cfg:"headerType" json:"type"`

	// RequestHost represents the masquerade domain name of the underlying transport.
	RequestHost string `cfg:"requestHost" json:"host"`

	// Path represents the ws path, h2 path or QUIC key/Kcp seed.
	Path string `cfg:"path" json:"path"`

	// StreamSecurity represents the underlying transport layer security.
	StreamSecurity string `cfg:"streamSecurity" json:"tls"`

	// SNI represents the SNI option used when TLS is enabled.
	SNI string `cfg:"sni" json:"sni"`
}

// WriteShareLink writes a shared link to w.
func (v *VMessServer) WriteShareLink(w io.Writer) error {
	b64Writer := base64.NewEncoder(base64.StdEncoding, w)
	defer b64Writer.Close()
	jsonEncoder := jsoniter.ConfigFastest.NewEncoder(b64Writer)
	w.Write([]byte("vmess://"))
	return jsonEncoder.Encode(v)
}

// ShareLink returns the shared link bytes.
func (v *VMessServer) ShareLink() (string, error) {
	buf := &strings.Builder{}
	err := v.WriteShareLink(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
