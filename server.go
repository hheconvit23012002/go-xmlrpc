package xmlrpc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	handlers map[string]MethodHandler
	logger   LoggerInterface
}

func NewServer(config ServerConfig) *Server {
	return &Server{
		handlers: make(map[string]MethodHandler),
		logger:   config.Logger,
	}
}

func (s *Server) RegisterHandler(method string, handler MethodHandler) {
	s.handlers[method] = handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, 405, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendError(w, 500, "Failed to read request body")
		return
	}

	s.logger.Debug("received request", "body", string(body))

	var methodCall MethodCall
	if err := xml.Unmarshal(body, &methodCall); err != nil {
		s.logger.Error("failed to unmarshal request", "error", err)
		s.sendError(w, 400, "Invalid XML-RPC request")
		return
	}

	handler, exists := s.handlers[methodCall.MethodName]
	if !exists {
		s.sendError(w, 400, fmt.Sprintf("Unknown method: %s", methodCall.MethodName))
		return
	}

	result, err := handler.Handle(methodCall.Params.Params)
	if err != nil {
		s.sendError(w, 500, err.Error())
		return
	}

	s.sendResponse(w, result)
}

func (s *Server) sendResponse(w http.ResponseWriter, result interface{}) {
	jsonResponse, err := json.Marshal(result)
	if err != nil {
		s.sendError(w, 500, "Failed to create response")
		return
	}

	xmlResponse := MethodResponse{
		Params: &ParamsArray{
			Params: []ParamValue{
				{
					Value: ValueType{
						StructValue: &StructType{
							Members: []MemberType{
								{
									Name: "JsonResult",
									Value: ValueType{
										StringValue: string(jsonResponse),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "text/xml")
	if err := xml.NewEncoder(w).Encode(&xmlResponse); err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) sendError(w http.ResponseWriter, code int, message string) {
	response := MethodResponse{
		Fault: &FaultContainer{
			Value: ValueType{
				StructValue: &StructType{
					Members: []MemberType{
						{Name: "faultCode", Value: ValueType{IntValue: int64(code)}},
						{Name: "faultString", Value: ValueType{StringValue: message}},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "text/xml")
	if err := xml.NewEncoder(w).Encode(&response); err != nil {
		s.logger.Error("failed to encode error response", "error", err)
	}
}
