// Package devicehandler provides the device handler for the server.
package devicehandler

import "aibuddy/pkg/ahttp"

// RtcHandler is the handler for the RTC service.
type RtcHandler struct{}

// NewRtcHandler creates a new RtcHandler.
func NewRtcHandler() *RtcHandler {
	return &RtcHandler{}
}

// GenerateAIAgentCall is the handler for the GenerateAIAgentCall service.
func (h *RtcHandler) GenerateAIAgentCall(state *ahttp.State, req *GenerateAIAgentCallRequest) error {
	_ = state
	_ = req
	return nil
}
