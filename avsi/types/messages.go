package types

import (
	"io"
	"math"

	"github.com/cosmos/gogoproto/proto"

	"github.com/0xPellNetwork/pelldvs/libs/protoio"
)

const (
	maxMsgSize = math.MaxInt32 // 2GB
)

// WriteMessage writes a varint length-delimited protobuf message.
func WriteMessage(msg proto.Message, w io.Writer) error {
	protoWriter := protoio.NewDelimitedWriter(w)
	_, err := protoWriter.WriteMsg(msg)
	return err
}

// ReadMessage reads a varint length-delimited protobuf message.
func ReadMessage(r io.Reader, msg proto.Message) error {
	_, err := protoio.NewDelimitedReader(r, maxMsgSize).ReadMsg(msg)
	return err
}

//----------------------------------------

func ToRequestEcho(message string) *Request {
	return &Request{
		Value: &Request_Echo{&RequestEcho{Message: message}},
	}
}

func ToRequestFlush() *Request {
	return &Request{
		Value: &Request_Flush{&RequestFlush{}},
	}
}

func ToRequestInfo(req *RequestInfo) *Request {
	return &Request{
		Value: &Request_Info{req},
	}
}

func ToRequestQuery(req *RequestQuery) *Request {
	return &Request{
		Value: &Request_Query{req},
	}
}

func ToRequestProcessDvsRequest(req *RequestProcessDVSRequest) *Request {
	return &Request{
		Value: &Request_ProcessDvsRequest{req},
	}
}

func ToRequestProcessDvsResponse(req *RequestProcessDVSResponse) *Request {
	return &Request{
		Value: &Request_ProcessDvsResponse{req},
	}
}

func ToResponseException(message string) *Response {
	return &Response{
		Value: &Response_Exception{&ResponseException{Error: message}},
	}
}

func ToResponseEcho(message string) *Response {
	return &Response{
		Value: &Response_Echo{&ResponseEcho{Message: message}},
	}
}

func ToResponseFlush() *Response {
	return &Response{
		Value: &Response_Flush{&ResponseFlush{}},
	}
}

func ToResponseInfo(res *ResponseInfo) *Response {
	return &Response{
		Value: &Response_Info{res},
	}
}

func ToResponseQuery(res *ResponseQuery) *Response {
	return &Response{
		Value: &Response_Query{res},
	}
}

func ToResponseProcessRequest(res *ResponseProcessDVSRequest) *Response {
	return &Response{
		Value: &Response_ProcessDvsRequest{res},
	}
}

func ToResponseProcessDVSResponse(res *ResponseProcessDVSResponse) *Response {
	return &Response{
		Value: &Response_ProcessDvsResponse{res},
	}
}
