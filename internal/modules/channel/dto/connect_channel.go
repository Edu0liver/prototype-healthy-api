package dto

// ConnectRequest carries the optional phone number used to also request a
// pairing code. When empty, the channel's stored account number is used.
type ConnectRequest struct {
	Number string `json:"number" binding:"omitempty"`
}

// ConnectResponse carries QR / pairing data. Both may be present so the panel
// can show the QR code and the pairing code simultaneously.
type ConnectResponse struct {
	QRCode      string `json:"qr_code,omitempty"`
	PairingCode string `json:"pairing_code,omitempty"`
	State       string `json:"state"`
}
