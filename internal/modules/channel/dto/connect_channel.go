package dto

// ConnectRequest selects the connection method.
type ConnectRequest struct {
	Method string `json:"method" binding:"omitempty,oneof=qr pairing"`
	Number string `json:"number" binding:"omitempty"`
}

// ConnectResponse carries QR / pairing data.
type ConnectResponse struct {
	QRCode      string `json:"qr_code,omitempty"`
	PairingCode string `json:"pairing_code,omitempty"`
	State       string `json:"state"`
}
