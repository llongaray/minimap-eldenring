package core

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	jsMarker = regexp.MustCompile(`\{\s*id:\s*"((?:[^"\\]|\\.)+)"\s*,\s*cat:\s*"([a-zA-Z0-9_-]+)"\s*,\s*name:\s*"((?:[^"\\]|\\.)+)"\s*,\s*master:\s*"(M00|M01|M10)"\s*,\s*px:\s*(-?\d+(?:\.\d+)?)\s*,\s*py:\s*(-?\d+(?:\.\d+)?)`)
	jsDesc   = regexp.MustCompile(`\bdesc:\s*"((?:[^"\\]|\\.)+)"`)
)

func FromReferenceJS(raw []byte) ([]Marker, error) {
	// Parse only simple literal records; NEVER evaluate downloaded JavaScript.
	matches := jsMarker.FindAllSubmatchIndex(raw, -1)
	if len(matches) == 0 {
		return nil, errors.New("nenhum marcador encontrado; formato JS não reconhecido")
	}
	out := make([]Marker, 0, len(matches))
	ids := map[string]bool{}
	for _, v := range matches {
		get := func(n int) string { return string(raw[v[2*n]:v[2*n+1]]) }
		id, err := strconv.Unquote(`"` + get(1) + `"`)
		if err != nil {
			continue
		}
		name, err := strconv.Unquote(`"` + get(3) + `"`)
		if err != nil {
			continue
		}
		x, errX := strconv.ParseFloat(get(5), 64)
		y, errY := strconv.ParseFloat(get(6), 64)
		if errX != nil || errY != nil {
			continue
		}
		m := Marker{ID: id, Category: get(2), Name: name, Master: Layer(get(4)), X: x, Y: y}
		// Desc is read from the same record line, and is never evaluated.
		end := v[1]
		lineEnd := strings.IndexByte(string(raw[end:]), '\n')
		if lineEnd >= 0 {
			end += lineEnd
		}
		if desc := jsDesc.FindSubmatch(raw[v[1]:end]); desc != nil {
			if d, e := strconv.Unquote(string(append(append([]byte{'"'}, desc[1]...), '"'))); e == nil {
				m.Description = d
			}
		}
		if m.Validate() == nil && !ids[m.ID] {
			ids[m.ID] = true
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("marcadores presentes mas inválidos")
	}
	return out, nil
}
func LoadMarkers(path string) ([]Marker, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(strings.ToLower(path), ".js") {
		return FromReferenceJS(raw)
	}
	var out []Marker
	if err = json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, m := range out {
		if err = m.Validate(); err != nil {
			return nil, err
		}
		if ids[m.ID] {
			return nil, errors.New("ID repetido")
		}
		ids[m.ID] = true
	}
	return out, nil
}
