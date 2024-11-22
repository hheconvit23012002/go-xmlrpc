package go_xmlrpc

import (
	"encoding/xml"
)

type MethodCall struct {
	XMLName    xml.Name    `xml:"methodCall"`
	MethodName string      `xml:"methodName"`
	Params     ParamsArray `xml:"params"`
}

type ParamsArray struct {
	Params []ParamValue `xml:"param"`
}

type ParamValue struct {
	Value ValueType `xml:"value"`
}

type ValueType struct {
	StringValue string      `xml:"string,omitempty"`
	IntValue    int64       `xml:"int,omitempty"`
	StructValue *StructType `xml:"struct,omitempty"`
}

type StructType struct {
	Members []MemberType `xml:"member"`
}

type MemberType struct {
	Name  string    `xml:"name"`
	Value ValueType `xml:"value"`
}

type MethodResponse struct {
	XMLName xml.Name        `xml:"methodResponse"`
	Params  *ParamsArray    `xml:"params,omitempty"`
	Fault   *FaultContainer `xml:"fault,omitempty"`
}

type FaultContainer struct {
	Value ValueType `xml:"value"`
}

type MethodHandler interface {
	Handle(params []ParamValue) (interface{}, error)
}

type ServerConfig struct {
	Logger LoggerInterface
}

type LoggerInterface interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
