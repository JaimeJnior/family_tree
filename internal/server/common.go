package server

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
)

func WriteErrorMessage(w http.ResponseWriter, r *http.Request, status int, err error) error {
	w.WriteHeader(status)
	body, err := json.Marshal(Error{
		Message: err.Error(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.Write(body)
	return nil

}

func WriteError(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
}

func WriteJsonBody(w http.ResponseWriter, r *http.Request, status int, body interface{}) error {

	bodyResponse, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(status)
	w.Write(bodyResponse)
	return nil
}

func WriteXMLBody(w http.ResponseWriter, r *http.Request, status int, body interface{}) error {

	bodyResponse, err := xml.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(status)
	w.Write(bodyResponse)
	return nil
}
func WriteGOBBody(w http.ResponseWriter, r *http.Request, status int, body interface{}) error {
	var byteBody bytes.Buffer
	gobEncoder := gob.NewEncoder(&byteBody)

	if err := gobEncoder.Encode(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(status)
	w.Write(byteBody.Bytes())
	return nil
}

func WriteErrorValidation(w http.ResponseWriter, r *http.Request, err error) error {

	for mapError, status := range ErrorStatusResponseMap {
		if errors.Is(err, mapError) {
			return WriteErrorMessage(w, r, status, err)

		}
	}
	return WriteErrorMessage(w, r, http.StatusInternalServerError, err)
}

func ResponseStrategy(acceptHeader []string, status int) func(w http.ResponseWriter, r *http.Request, body any) error {
	for _, accept := range acceptHeader {
		switch accept {
		case AcceptApplicationBinary, AcceptOctetStream:
			return func(w http.ResponseWriter, r *http.Request, body any) error {
				return WriteGOBBody(w, r, status, body)
			}
		case AcceptApplicationXML:
			return func(w http.ResponseWriter, r *http.Request, body any) error {
				return WriteXMLBody(w, r, status, body)
			}
		case AcceptApplicationJson:
			return func(w http.ResponseWriter, r *http.Request, body any) error {
				return WriteJsonBody(w, r, status, body)
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request, body any) error {
		return WriteJsonBody(w, r, status, body)
	}
}
