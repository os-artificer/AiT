/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package geocode_test

import (
	"context"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/geocode"
	"github.com/os-artificer/AiT/pkg/testhttp"
)

func TestReverse_MockOK(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioOK})
	c := &geocode.Client{BaseURL: srv.ReverseURL(), HTTP: srv.HTTPClient(), UserAgent: "test"}
	addr, err := c.Reverse(context.Background(), 31.24, 121.49)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(addr, "外滩") {
		t.Fatalf("addr=%q", addr)
	}
}

func TestReverse_MockServerError(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioServerError})
	c := &geocode.Client{BaseURL: srv.ReverseURL(), HTTP: srv.HTTPClient()}
	_, err := c.Reverse(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReverse_MockUnableToGeocode(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioGeocodeUnable})
	c := &geocode.Client{BaseURL: srv.ReverseURL(), HTTP: srv.HTTPClient()}
	_, err := c.Reverse(context.Background(), 0, 0)
	if err == nil || !strings.Contains(err.Error(), "Unable to geocode") {
		t.Fatalf("err=%v", err)
	}
}
