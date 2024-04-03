package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/datatypes"
)

type AvailableTemplates string

const (
	LandingPage AvailableTemplates = "landingPage"
)

func (at *AvailableTemplates) Scan(value interface{}) error {
	*at = AvailableTemplates(value.(string))
	return nil
}

func (at AvailableTemplates) Value() (driver.Value, error) {
	return string(at), nil
}

func (at AvailableTemplates) String() string {
	return string(at)
}

func (at AvailableTemplates) IsValid() error {
	switch at {
	case LandingPage:
		return nil
	}

	return fmt.Errorf("invalid dashboard template. Expected one of %s, got %s", LandingPage, at)
}

type AvailableWidgets string

const (
	FavoriteServices    AvailableWidgets = "favoriteServices"
	NotificationsEvents AvailableWidgets = "notificationsEvents"
	LearningResources   AvailableWidgets = "learningResources"
	ExploreCapabilities AvailableWidgets = "exploreCapabilities"
	Edge                AvailableWidgets = "edge"
	Ansible             AvailableWidgets = "ansible"
	Rhel                AvailableWidgets = "rhel"
	Openshift           AvailableWidgets = "openshift"
	RecentlyVisited     AvailableWidgets = "recentlyVisited"
)

func (aw AvailableWidgets) IsValid() error {
	switch aw {
	case FavoriteServices, NotificationsEvents, LearningResources, ExploreCapabilities, Edge, Ansible, Rhel, Openshift, RecentlyVisited:
		return nil
	}

	return fmt.Errorf("invalid widget. Expected one of [%s, %s, %s, %s, %s, %s, %s, %s, %s] got %s", FavoriteServices, NotificationsEvents, LearningResources, ExploreCapabilities, Edge, Ansible, Rhel, Openshift, RecentlyVisited, aw)
}

type BaseWidgetDimensions struct {
	Width     int `json:"w"`
	Height    int `json:"h"`
	MaxHeight int `json:"maxH"`
	MinHeight int `json:"minH"`
}

func (bwd BaseWidgetDimensions) InitDimensions(w, h, maxH, minH int) BaseWidgetDimensions {
	if w < 1 || h < 1 || maxH < 1 || minH < 1 {
		panic("invalid widget dimensions, all values must be greater than 0")
	}
	bwd.Width = w
	bwd.Height = h
	bwd.MaxHeight = maxH
	bwd.MinHeight = minH
	return bwd
}

type GridItem struct {
	BaseWidgetDimensions
	Title  string `json:"title"`
	ID     string `json:"i"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Static bool   `json:"static"`
}

type TemplateConfig struct {
	Sm datatypes.JSONType[[]GridItem] `gorm:"not null;default null" json:"sm"`
	Md datatypes.JSONType[[]GridItem] `gorm:"not null;default null" json:"md"`
	Lg datatypes.JSONType[[]GridItem] `gorm:"not null;default null" json:"lg"`
	Xl datatypes.JSONType[[]GridItem] `gorm:"not null;default null" json:"xl"`
}

type GridSizes string

const (
	Sm GridSizes = "sm"
	Md GridSizes = "md"
	Lg GridSizes = "lg"
	Xl GridSizes = "xl"
)

func (gs GridSizes) IsValid() error {
	switch gs {
	case Sm, Md, Lg, Xl:
		return nil
	default:
		return errors.New(fmt.Errorf("invalid grid size, expected one of %s, %s, %s, %s", Sm, Md, Lg, Xl).Error())
	}
}

func (gs GridSizes) GetMaxWidth() (int, error) {
	err := gs.IsValid()
	if err != nil {
		return 0, err
	}
	switch gs {
	case Sm:
		return 1, nil
	case Md:
		return 2, nil
	case Lg:
		return 3, nil
	case Xl:
		return 4, nil
	default:
		return 0, errors.New("invalid grid size")
	}
}

func (gi GridItem) IsValid(variant GridSizes) error {
	if err := variant.IsValid(); err != nil {
		return err
	}

	if gi.ID == "" {
		return errors.New(`invalid grid item, field id "i" is required`)
	}

	if gi.Width < 1 || gi.Height < 1 || gi.MaxHeight < 1 || gi.MinHeight < 1 {
		return errors.New(`invalid grid item, height "h", width "w", maxHeight "maxH", mixHeight "minH" must be greater than 0`)
	}

	if gi.Height > gi.MaxHeight {
		return errors.New(fmt.Errorf(`invalid grid item, height "h" %d must be less than or equal to max height "maxH" %d`, gi.Height, gi.MaxHeight).Error())
	}

	if gi.Height < gi.MinHeight {
		return errors.New(fmt.Errorf(`invalid grid item, height "h" %d must be greater than or equal to min height "minH" %d`, gi.Height, gi.MinHeight).Error())
	}

	maxGridSize, err := variant.GetMaxWidth()
	if err != nil {
		return err
	}

	if gi.Width > maxGridSize {
		return errors.New(fmt.Errorf("invalid grid item, layout variant %s, width must be less than or equal to %d", variant, maxGridSize).Error())
	}

	if gi.X > maxGridSize {
		return errors.New(fmt.Errorf("invalid grid item, layout variant %s, coordinate X must be less than or equal to %d", variant, maxGridSize).Error())
	}

	return nil
}

func (tc *TemplateConfig) SetLayoutSizeItems(layoutSize string, items []GridItem) *TemplateConfig {
	jsonItems := datatypes.NewJSONType[[]GridItem](items)
	reflect.ValueOf(tc).Elem().FieldByName(layoutSize).Set(reflect.ValueOf(jsonItems))
	return tc
}

func (tc TemplateConfig) IsValid() error {
	configs := reflect.ValueOf(tc)
	typeOfS := configs.Type()

	for i := 0; i < configs.NumField(); i++ {
		dgi := configs.Field(i).Interface().(datatypes.JSONType[[]GridItem])
		items := dgi.Data()
		layoutSize := typeOfS.Field(i).Tag.Get("json")
		for _, gi := range items {
			// initialize coordinates if they do not exist
			if gi.Y == 0 {
				gi.Y = 0
			}
			if gi.X == 0 {
				gi.X = 0
			}

			err := gi.IsValid(GridSizes(layoutSize))
			if err != nil {
				return err
			}
		}

		if len(items) > 0 {
			// replace only non empty items, not the whole config
			tc.SetLayoutSizeItems(typeOfS.Field(i).Name, items)
		}
	}

	return nil
}

type DashboardTemplateBase struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type DashboardTemplate struct {
	BaseModel
	UserIdentityID uint                  `json:"userIdentityID"`
	Default        bool                  `gorm:"not null;default:false" json:"default"`
	TemplateBase   DashboardTemplateBase `gorm:"not null;default null; embedded" 'json:"templateBase"`
	TemplateConfig TemplateConfig        `gorm:"not null;default null; embedded" json:"templateConfig"`
}

type BaseDashboardTemplate struct {
	Name           string         `json:"name"`
	DisplayName    string         `json:"displayName"`
	TemplateConfig TemplateConfig `json:"templateConfig"`
}

type BaseTemplates map[AvailableTemplates]BaseDashboardTemplate

type WidgetIcons string

const (
	BellIcon             WidgetIcons = "BellIcon"
	HistoryIcon          WidgetIcons = "HistoryIcon"
	OutlinedBookmarkIcon WidgetIcons = "OutlinedBookmarkIcon"
	RocketIcon           WidgetIcons = "RocketIcon"
	StarIcon             WidgetIcons = "StarIcon"
	RhelIcon             WidgetIcons = "RhelIcon"
	OpenshiftIcon        WidgetIcons = "OpenshiftIcon"
	EdgeIcon             WidgetIcons = "EdgeIcon"
	AnsibleIcon          WidgetIcons = "AnsibleIcon"
)

func (wi WidgetIcons) IsValid() error {
	switch wi {
	case BellIcon, HistoryIcon, OutlinedBookmarkIcon, RocketIcon, StarIcon, RhelIcon, OpenshiftIcon, EdgeIcon, AnsibleIcon:
		return nil
	}

	return fmt.Errorf("invalid widget icon. Expected one of %s, %s, %s, %s, %s, %s, %s, %s, %s, got %s", BellIcon, HistoryIcon, OutlinedBookmarkIcon, RocketIcon, StarIcon, RhelIcon, OpenshiftIcon, EdgeIcon, AnsibleIcon, wi)
}

type WidgetHeaderLink struct {
	Title string `json:"title,omitempty"`
	Href  string `json:"href,omitempty"`
}

type WidgetConfiguration struct {
	Title      string           `json:"title"`
	Icon       WidgetIcons      `json:"icon,omitempty"`
	HeaderLink WidgetHeaderLink `json:"headerLink,omitempty"`
}

type ModuleFederationMetadata struct {
	Scope      string               `json:"scope"`
	Module     string               `json:"module"`
	ImportName string               `json:"importName,omitempty"`
	Defaults   BaseWidgetDimensions `json:"defaults"`
	Config     WidgetConfiguration  `json:"config"`
}

type WidgetModuleFederationMapping map[AvailableWidgets]ModuleFederationMetadata

func (dt DashboardTemplate) IsValid() error {
	if dt.TemplateBase.Name == "" {
		return errors.New("invalid template name")
	}

	if dt.TemplateBase.DisplayName == "" {
		return errors.New("invalid template display name")
	}

	if err := dt.TemplateConfig.IsValid(); err != nil {
		return err
	}

	return nil
}

func (dt *DashboardTemplate) EncodeBase64() (string, error) {
	strippedDt := DashboardTemplate{
		TemplateBase:   dt.TemplateBase,
		TemplateConfig: dt.TemplateConfig,
		Default:        false,
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(&strippedDt)
	if err != nil {
		return "", err
	}
	encoder.Close()
	return buf.String(), nil
}

func DecodeDashboardBase64(encoded string) (DashboardTemplate, error) {
	var dt DashboardTemplate
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(encoded))
	err := json.NewDecoder(decoder).Decode(&dt)
	if err != nil {
		return dt, err
	}

	err = dt.IsValid()

	// strip out user specific data
	return DashboardTemplate{
		TemplateBase:   dt.TemplateBase,
		TemplateConfig: dt.TemplateConfig,
		Default:        false,
	}, err
}
