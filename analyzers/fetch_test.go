package analyzers_test

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gostaticanalysis/vetgen/analyzers"
)

func TestFetcher_List(t *testing.T) {

	cases := map[string]struct {
		body   string
		want   []*analyzers.Info
		hasErr bool
	}{
		"empty": {html(t, []string{}), nil, false},
		"single": {
			html(t, []string{"github.com/gostaticanalysis/dupimport"}),
			[]*analyzers.Info{{
				Import: "github.com/gostaticanalysis/dupimport",
				Name:   "Analyzer",
			}},
			false,
		},
		"noanalyzer": {
			html(t, []string{"github.com/gostaticanalysis/analysisutil"}),
			nil,
			false,
		},
		"invalidpath": {
			html(t, []string{"__invalid_path__"}),
			nil,
			false,
		},
	}

	for name, tt := range cases {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tt.body)
			}))
			defer sv.Close()

			f := analyzers.Fetcher{
				HTTPClient: sv.Client(),
				BaseURL:    sv.URL + "/",
			}
			ctx := context.Background()
			got, err := f.List(ctx)
			switch {
			case tt.hasErr && err == nil:
				t.Fatal("Expected error has not occurred")
			case !tt.hasErr && err != nil:
				t.Fatal("Unexpected error:", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("want %#v but got %#v", tt.want, got)
			}

			for i := range got {
				if *got[i] != *tt.want[i] {
					t.Errorf("%dth element want %#v but got %#v", i, tt.want[i], got[i])
				}
			}
		})
	}
}

var htmlTempl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html><body><ul class="ImportedBy-list">{{range .}}
	<li class="Details-indent">
		<a class="u-breakWord" href="/{{.}}">{{.}}</a>
	</li>
{{end}}</ul></div></details></ul></body></html>`))

func html(t *testing.T, imports []string) string {
	var buf bytes.Buffer
	if err := htmlTempl.Execute(&buf, imports); err != nil {
		t.Fatal("Unexpected error:", err)
	}
	return buf.String()
}
