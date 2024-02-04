package talker_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Arsfiqball/csverse/talker"
)

func TestAttr(t *testing.T) {
	type sampleT struct {
		ID talker.Attr[int] `json:"id,omitzero"`
	}

	t.Run("Omit", func(t *testing.T) {
		var sample sampleT

		if err := json.Unmarshal([]byte(`{}`), &sample); err != nil {
			t.Fatal(err)
		}

		if sample.ID != talker.Omit[int]() {
			t.Fatal("attribute is not null")
		}

		if sample.ID.Present() {
			t.Fatal("attribute is present")
		}

		if sample.ID.Filled() {
			t.Fatal("attribute is filled")
		}

		if fmt.Sprintf("%v", sample.ID) != "0" {
			t.Fatal("attribute is not 0")
		}

		b, err := json.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}

		// This should be `{}` instead of `{"id":null}`
		// But the current json.Marshal implementation doesn't support omitzero tag
		if string(b) != `{"id":null}` {
			t.Fatal("attribute is not omitted")
		}
	})

	t.Run("Null", func(t *testing.T) {
		var sample sampleT

		if err := json.Unmarshal([]byte(`{"id":null}`), &sample); err != nil {
			t.Fatal(err)
		}

		if sample.ID != talker.Null[int]() {
			t.Fatal("attribute is not null")
		}

		if !sample.ID.Present() {
			t.Fatal("attribute is not present")
		}

		if sample.ID.Filled() {
			t.Fatal("attribute is filled")
		}

		if fmt.Sprintf("%v", sample.ID) != "0" {
			t.Fatal("attribute is not 0")
		}

		b, err := json.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != `{"id":null}` {
			t.Fatal("attribute is not null")
		}
	})

	t.Run("Empty", func(t *testing.T) {
		var sample sampleT

		if err := json.Unmarshal([]byte(`{"id":0}`), &sample); err != nil {
			t.Fatal(err)
		}

		if sample.ID != talker.Empty[int]() {
			t.Fatal("attribute is not empty")
		}

		if !sample.ID.Present() {
			t.Fatal("attribute is not present")
		}

		if !sample.ID.Filled() {
			t.Fatal("attribute is not filled")
		}

		if sample.ID.Get() != 0 {
			t.Fatal("attribute is not zero")
		}

		if fmt.Sprintf("%v", sample.ID) != "0" {
			t.Fatal("attribute is not 0")
		}

		b, err := json.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != `{"id":0}` {
			t.Fatal("attribute is not empty")
		}
	})

	t.Run("Value", func(t *testing.T) {
		var sample sampleT

		if err := json.Unmarshal([]byte(`{"id":30}`), &sample); err != nil {
			t.Fatal(err)
		}

		if sample.ID != talker.Value(30) {
			t.Fatal("attribute is not 30")
		}

		if !sample.ID.Present() {
			t.Fatal("attribute is not present")
		}

		if !sample.ID.Filled() {
			t.Fatal("attribute is not filled")
		}

		if sample.ID.Get() != 30 {
			t.Fatal("attribute is not 30")
		}

		if fmt.Sprintf("%v", sample.ID) != "30" {
			t.Fatal("attribute is not 30")
		}

		b, err := json.Marshal(sample)
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != `{"id":30}` {
			t.Fatal("attribute is not 30")
		}
	})

	t.Run("Filled", func(t *testing.T) {
		sample := struct {
			ID   talker.Attr[int]    `json:"id"`
			Name talker.Attr[string] `json:"name"`
		}{
			ID: talker.Value(30),
		}

		if !talker.Filled(sample.ID) {
			t.Fatal("attribute is not filled")
		}

		if talker.Filled(sample.Name) {
			t.Fatal("attribute is filled")
		}

		if talker.Filled(sample.ID, sample.Name) {
			t.Fatal("all attribute detected as filled")
		}
	})

	t.Run("Present", func(t *testing.T) {
		sample := struct {
			ID   talker.Attr[int]    `json:"id"`
			Name talker.Attr[string] `json:"name"`
		}{
			ID: talker.Value(30),
		}

		if !talker.Present(sample.ID) {
			t.Fatal("attribute is not present")
		}

		if talker.Present(sample.Name) {
			t.Fatal("attribute is present")
		}

		if talker.Present(sample.ID, sample.Name) {
			t.Fatal("all attribute detected as present")
		}
	})
}
