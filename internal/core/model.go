// Package core contains pure (non-Windows) rules used by Aray Minimap.
package core

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const SchemaVersion = 1

type Layer string

const (
	Surface     Layer = "M00"
	Underground Layer = "M01"
	Shadow      Layer = "M10"
)

func (l Layer) Valid() bool { return l == Surface || l == Underground || l == Shadow }
func (l Layer) Label() string {
	switch l {
	case Surface:
		return "Superfície"
	case Underground:
		return "Subterrâneo"
	case Shadow:
		return "Terra das Sombras"
	}
	return "Desconhecido"
}

type Marker struct {
	ID          string  `json:"id"`
	Category    string  `json:"cat"`
	Name        string  `json:"name"`
	Master      Layer   `json:"master"`
	X           float64 `json:"px"`
	Y           float64 `json:"py"`
	Description string  `json:"desc,omitempty"`
}

func (m Marker) Validate() error {
	if strings.TrimSpace(m.ID) == "" || strings.TrimSpace(m.Name) == "" || !m.Master.Valid() || math.IsNaN(m.X) || math.IsNaN(m.Y) || m.X < 0 || m.Y < 0 || m.X > 10496 || m.Y > 10496 {
		return fmt.Errorf("marcador inválido %q", m.ID)
	}
	return nil
}

type Point struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Layer   Layer   `json:"layer"`
	Heading float64 `json:"heading"`
	Source  string  `json:"source"`
}

func (p Point) Validate() error {
	if !p.Layer.Valid() || !finite(p.X) || !finite(p.Y) || p.X < 0 || p.Y < 0 || p.X > 10496 || p.Y > 10496 || !finite(p.Heading) {
		return errors.New("posição fora da área ou camada inválida")
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

type Calibration struct {
	// The game-reported point is transformed as masterX = ScaleX*x + OffsetX.
	// Calibration is NOT enabled until independently verified in this game build.
	Enabled bool    `json:"enabled"`
	ScaleX  float64 `json:"scale_x"`
	ScaleY  float64 `json:"scale_y"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
}

func (c Calibration) Apply(x, y float64, layer Layer, heading float64) (Point, error) {
	if !c.Enabled {
		return Point{}, errors.New("calibração de coordenadas pendente")
	}
	if !finite(c.ScaleX) || !finite(c.ScaleY) || c.ScaleX == 0 || c.ScaleY == 0 {
		return Point{}, errors.New("escala inválida")
	}
	p := Point{X: c.ScaleX*x + c.OffsetX, Y: c.ScaleY*y + c.OffsetY, Layer: layer, Heading: heading, Source: "live"}
	return p, p.Validate()
}

type Flags struct {
	Pinned bool `json:"pinned"`
	Found  bool `json:"found"`
}
type Settings struct {
	Schema           int              `json:"schema"`
	Size             int              `json:"size"`
	Opacity          int              `json:"opacity"`
	Zoom             float64          `json:"zoom"`
	Position         string           `json:"position"`
	NorthFixed       bool             `json:"north_fixed"`
	Layer            Layer            `json:"layer"`
	F6Key            int              `json:"f6_key"`
	HiddenCategories map[string]bool  `json:"hidden_categories"`
	Markers          map[string]Flags `json:"markers"`
	Calibration      Calibration      `json:"calibration"`
}

func Defaults() Settings {
	return Settings{Schema: SchemaVersion, Size: 280, Opacity: 95, Zoom: 1, Position: "top-right", NorthFixed: true, Layer: Surface, F6Key: 0x75, Markers: map[string]Flags{}, HiddenCategories: map[string]bool{}}
}
func (s Settings) Validate() error {
	if s.Schema != SchemaVersion || s.Size < 180 || s.Size > 500 || s.Opacity < 30 || s.Opacity > 100 || s.Zoom < 0.4 || s.Zoom > 8 || !s.Layer.Valid() {
		return errors.New("configuração inválida ou incompatível")
	}
	switch s.Position {
	case "top-left", "top-right", "bottom-left", "bottom-right":
	default:
		return errors.New("posição do HUD inválida")
	}
	if s.F6Key < 0x70 || s.F6Key > 0x87 {
		return errors.New("atalho inválido")
	}
	if len(s.Markers) > 50000 {
		return errors.New("quantidade excessiva de marcadores")
	}
	return nil
}
func Visible(m Marker, s Settings) bool {
	return m.Master == s.Layer && !s.HiddenCategories[m.Category]
}
func TogglePin(s *Settings, id string) {
	if s.Markers == nil {
		s.Markers = map[string]Flags{}
	}
	v := s.Markers[id]
	v.Pinned = !v.Pinned
	s.Markers[id] = v
}
func ToggleFound(s *Settings, id string) {
	if s.Markers == nil {
		s.Markers = map[string]Flags{}
	}
	v := s.Markers[id]
	v.Found = !v.Found
	s.Markers[id] = v
}

// CategoryPTBR translates the reference dataset category identifiers without
// changing the original IDs used for persistence and filtering.
func CategoryPTBR(id string) string {
	names := map[string]string{
		"grace": "Locais de Graça", "boss": "Chefes", "npc": "Personagens e missões",
		"weapon": "Armas", "talisman": "Talismãs", "spirit": "Cinzas Espirituais",
		"ash": "Cinzas de Guerra", "spell": "Feitiços e Encantamentos", "armor": "Armaduras",
		"golden": "Sementes Douradas", "tear": "Lágrimas Sagradas", "crystal": "Lágrimas de Cristal",
		"greatrune": "Grandes Runas", "larval": "Lágrimas Larvais", "memory": "Pedras de Memória",
		"whetblade": "Facas de Amolar", "region": "Regiões",
	}
	if text, ok := names[id]; ok {
		return text
	}
	return id
}
