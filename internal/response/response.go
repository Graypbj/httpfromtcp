package response

import (
	"errors"
	"fmt"
	"io"
)

type StatusCode int

type Writer struct {
	Writer           io.Writer
	statusCode       StatusCode
	headers          map[string]string
	body             []byte
	statusCodeWasSet bool
	headersWereSet   bool
	bodyWasWritten   bool
}

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) []byte {
	reasonPhrase := ""
	switch statusCode {
	case StatusCodeSuccess:
		reasonPhrase = "OK"
	case StatusCodeBadRequest:
		reasonPhrase = "Bad Request"
	case StatusCodeInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := w.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.statusCodeWasSet {
		return errors.New("status line already set")
	}
	w.statusCode = statusCode
	w.statusCodeWasSet = true
	return nil
}

func (w *Writer) WriteHeaders(headers map[string]string) error {
	if !w.statusCodeWasSet {
		return errors.New("must call WriteStatusLine before WriteHeaders")
	}
	if w.headersWereSet {
		return errors.New("headers already set")
	}
	w.headers = headers
	w.headersWereSet = true
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if !w.statusCodeWasSet || !w.headersWereSet {
		return 0, errors.New("must call WriteStatusLine and WriteHeaders before WriteBody")
	}
	if w.bodyWasWritten {
		return 0, errors.New("body was already written")
	}
	w.body = make([]byte, len(p))
	copy(w.body, p)
	w.bodyWasWritten = true
	return len(p), nil
}

func (w *Writer) Flush(to io.Writer) error {
	statusLine := getStatusLine(w.statusCode)
	if _, err := to.Write(statusLine); err != nil {
		return err
	}

	for k, v := range w.headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", k, v)
		if _, err := to.Write([]byte(headerLine)); err != nil {
			return err
		}
	}

	if _, err := to.Write([]byte("\r\n")); err != nil {
		return err
	}

	if len(w.body) > 0 {
		if _, err := to.Write(w.body); err != nil {
			return err
		}
	}

	return nil
}
