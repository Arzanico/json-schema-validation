package server

type PayloadValidationRequest struct {
	Payload *string `json:"payload"`
}

func NewPayloadValidationRequest() *PayloadValidationRequest {
	return &PayloadValidationRequest{}
}
