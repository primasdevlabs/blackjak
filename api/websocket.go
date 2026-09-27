package api

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"blackjak/agent"
	"blackjak/protocol"
	"blackjak/workspace"
)

func strField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// WebSocketConn wraps a hijacked net.Conn for RFC 6455 WebSocket framing.
type WebSocketConn struct {
	conn bufio.ReadWriter
	rwc  net.Conn
	mu   sync.Mutex
}

func computeAcceptKey(challenge string) string {
	h := sha1.New()
	h.Write([]byte(challenge + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// UpgradeHandler upgrades an HTTP connection to WebSocket.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Expected Upgrade: websocket", http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "Missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	rwc, buf, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	acceptKey := computeAcceptKey(key)

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"

	buf.WriteString(resp)
	buf.Flush()

	ws := &WebSocketConn{
		conn: *buf,
		rwc:  rwc,
	}
	defer rwc.Close()

	// Subscribe to all agent events
	eventCh, unsubscribe := s.broker.Subscribe("")
	defer unsubscribe()

	// Writer goroutine streaming events to client
	stopCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case evt, ok := <-eventCh:
				if !ok {
					return
				}
				msg := protocol.ServerMessage{
					ProtocolVersion: protocol.ProtocolVersion,
					Type:            string(evt.Type),
					RunID:           evt.RunID,
					AgentID:         evt.AgentID,
					Timestamp:       evt.Timestamp,
					Data:            evt.Data,
				}
				data, err := json.Marshal(msg)
				if err == nil {
					_ = ws.WriteTextFrame(data)
				}
			}
		}
	}()

	// Reader loop handling client frames
	for {
		payload, err := ws.ReadFrame()
		if err != nil {
			close(stopCh)
			break
		}

		var clientMsg protocol.ClientMessage
		if err := json.Unmarshal(payload, &clientMsg); err != nil {
			continue
		}

		s.processClientMessage(ws, clientMsg)
	}
}

// WriteFrame writes a single RFC 6455 frame with the given opcode.
func (ws *WebSocketConn) WriteFrame(opcode byte, payload []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	w := ws.conn.Writer
	if err := w.WriteByte(0x80 | opcode); err != nil {
		return err
	}

	length := len(payload)
	if length < 126 {
		if err := w.WriteByte(byte(length)); err != nil {
			return err
		}
	} else if length <= 65535 {
		if err := w.WriteByte(126); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint16(length)); err != nil {
			return err
		}
	} else {
		if err := w.WriteByte(127); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, uint64(length)); err != nil {
			return err
		}
	}

	if _, err := w.Write(payload); err != nil {
		return err
	}
	return w.Flush()
}

func (s *Server) processClientMessage(ws *WebSocketConn, msg protocol.ClientMessage) {
	switch msg.Type {
	case protocol.ClientMsgPing:
		pong := protocol.ServerMessage{
			ProtocolVersion: protocol.ProtocolVersion,
			Type:            "pong",
			Timestamp:       time.Now(),
		}
		data, _ := json.Marshal(pong)
		_ = ws.WriteTextFrame(data)

	case protocol.ClientMsgHostHello:
		raw, err := json.Marshal(msg.Data)
		if err != nil {
			break
		}
		var info protocol.HostInfo
		if err := json.Unmarshal(raw, &info); err == nil && info.Name != "" {
			s.SetHostInfo(info)
		}

	case protocol.ClientMsgRunStart:
		prompt, _ := msg.Data["prompt"].(string)
		wsPath, _ := msg.Data["workspace"].(string)
		if wsPath == "" {
			wsPath = s.workspace.RootPath
		}
		var attachments []workspace.Attachment
		if raw, ok := msg.Data["attachments"].([]interface{}); ok {
			for _, item := range raw {
				if m, ok := item.(map[string]interface{}); ok {
					attachments = append(attachments, workspace.Attachment{
						ID:   strField(m, "id"),
						Type: workspace.AttachmentType(strField(m, "type")),
						Path: strField(m, "path"),
						Name: strField(m, "name"),
					})
				}
			}
		}
		if prompt != "" {
			run := s.runManager.CreateRun(prompt, wsPath, attachments...)
			go func() {
				_ = s.agent.ExecuteRun(run.Context(), run, s.runManager)
			}()
		}

	case protocol.ClientMsgRunCancel:
		runID, _ := msg.Data["runId"].(string)
		if runID != "" {
			s.runManager.CancelRun(runID)
		}

	case protocol.ClientMsgApprovalRespond:
		reqID, _ := msg.Data["requestId"].(string)
		runID, _ := msg.Data["runId"].(string)
		granted, _ := msg.Data["granted"].(bool)
		reason, _ := msg.Data["reason"].(string)

		if runID != "" && reqID != "" {
			_ = s.runManager.SubmitApproval(runID, agent.ApprovalResponse{
				RequestID: reqID,
				Granted:   granted,
				Reason:    reason,
			})
		}
	}
}

// ReadFrame reads an RFC 6455 frame. Control frames (ping/pong/close) are
// handled inline — a ping is answered with a pong — and only data frame
// payloads are returned to the caller. Browsers send transport pings; a
// server that never pongs gets disconnected.
func (ws *WebSocketConn) ReadFrame() ([]byte, error) {
	ws.mu.Lock()
	r := ws.conn.Reader
	ws.mu.Unlock()

	for {
		b1, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		fin := b1&0x80 != 0
		opcode := b1 & 0x0F

		payload, err := readFramePayload(r)
		if err != nil {
			return nil, err
		}

		switch opcode {
		case 0x8: // close
			return nil, io.EOF
		case 0x9: // ping → pong
			_ = ws.WriteFrame(0xA, payload)
			continue
		case 0xA: // pong
			continue
		case 0x0: // continuation of a fragmented message — unsupported, skip
			continue
		}

		if !fin {
			// First fragment of a multi-frame message — we don't reassemble
			// fragmented payloads; treat it as a standalone frame.
			continue
		}

		if opcode == 0x1 || opcode == 0x2 {
			return payload, nil
		}
	}
}

func readFramePayload(r *bufio.Reader) ([]byte, error) {

	b2, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	masked := (b2 & 0x80) != 0
	length := uint64(b2 & 0x7F)

	if length == 126 {
		var l uint16
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return nil, err
		}
		length = uint64(l)
	} else if length == 127 {
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return nil, err
		}
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	if masked {
		for i := uint64(0); i < length; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	return payload, nil
}

// WriteTextFrame writes a single RFC 6455 unmasked text frame.
func (ws *WebSocketConn) WriteTextFrame(payload []byte) error {
	return ws.WriteFrame(0x1, payload)
}
