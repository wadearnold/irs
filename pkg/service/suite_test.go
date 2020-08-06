// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package service_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/irs/pkg/service"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type ServerTest struct {
	testServer http.Handler
}

var _ = check.Suite(&ServerTest{})

func (t *ServerTest) SetUpSuite(c *check.C) {
	var err error
	t.testServer, err = service.TestConfigureHandlers()
	c.Assert(err, check.IsNil)
}

func (t *ServerTest) TearDownSuite(c *check.C) {}

func (t *ServerTest) SetUpTest(c *check.C) {}

func (t *ServerTest) TearDownTest(c *check.C) {}

func (t *ServerTest) makeRequest(method, url, body string, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	return recorder, request
}

func (t *ServerTest) getWriter(name string, c *check.C) (*multipart.Writer, *bytes.Buffer) {
	path := filepath.Join("..", "..", "test", "testdata", name)
	file, err := os.Open(path)
	c.Assert(err, check.IsNil)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	c.Assert(err, check.IsNil)
	_, err = io.Copy(part, file)
	c.Assert(err, check.IsNil)
	return writer, body
}

func (t *ServerTest) getErrWriter(name string, c *check.C) (*multipart.Writer, *bytes.Buffer) {
	path := filepath.Join("..", "..", "test", "testdata", name)
	file, err := os.Open(path)
	c.Assert(err, check.IsNil)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("err", filepath.Base(path))
	c.Assert(err, check.IsNil)
	_, err = io.Copy(part, file)
	c.Assert(err, check.IsNil)
	return writer, body
}

func (t *ServerTest) TestUnknownRequest(c *check.C) {
	recorder, request := t.makeRequest(http.MethodGet, "/unknown", "", c)
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (t *ServerTest) TestHealth(c *check.C) {
	recorder, request := t.makeRequest(http.MethodGet, "/health", "", c)
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestJsonPrint(c *check.C) {
	writer, body := t.getWriter("oneTransactionFile.json", c)
	err := writer.WriteField("format", "json")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/print", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestIrsPrint(c *check.C) {
	writer, body := t.getWriter("oneTransactionFile.json", c)
	err := writer.WriteField("format", "irs")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/print", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestJsonConvert(c *check.C) {
	writer, body := t.getWriter("oneTransactionFile.json", c)
	err := writer.WriteField("format", "json")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/convert", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestIrsConvert(c *check.C) {
	writer, body := t.getWriter("oneTransactionFile.json", c)
	err := writer.WriteField("format", "irs")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/convert", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestValidator(c *check.C) {
	writer, body := t.getWriter("oneTransactionFile.json", c)
	err := writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/validator", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestWithInvalidForm(c *check.C) {
	writer, body := t.getErrWriter("oneTransactionFile.json", c)
	err := writer.WriteField("format", "json")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/print", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (t *ServerTest) TestPrintWithInvalidData(c *check.C) {
	writer, body := t.getWriter("fileWithInvalidPayment.json", c)
	err := writer.WriteField("format", "json")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/print", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}

func (t *ServerTest) TestValidateWithInvalidData(c *check.C) {
	writer, body := t.getWriter("fileWithInvalidPayment.json", c)
	err := writer.WriteField("format", "json")
	c.Assert(err, check.IsNil)
	err = writer.Close()
	c.Assert(err, check.IsNil)
	recorder, request := t.makeRequest(http.MethodPost, "/validator", body.String(), c)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	t.testServer.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotImplemented)
}
