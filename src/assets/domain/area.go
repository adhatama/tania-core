package domain

import (
	"time"

	"github.com/Tanibox/tania-server/src/helper/validationhelper"
	uuid "github.com/satori/go.uuid"
)

type Area struct {
	UID       uuid.UUID              `json:"uid"`
	Name      string                 `json:"name"`
	Size      AreaUnit               `json:"size"`
	Type      AreaType               `json:"type"`
	Location  string                 `json:"location"`
	Photo     AreaPhoto              `json:"photo"`
	Notes     map[uuid.UUID]AreaNote `json:"-"`
	Reservoir Reservoir              `json:"-"`
	Farm      Farm                   `json:"-"`
}

const (
	AreaTypeSeeding = "SEEDING"
	AreaTypeGrowing = "GROWING"
)

type AreaType struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

func AreaTypes() []AreaType {
	return []AreaType{
		{Code: AreaTypeSeeding, Label: "Seeding"},
		{Code: AreaTypeGrowing, Label: "Growing"},
	}
}

func GetAreaType(code string) AreaType {
	for _, v := range AreaTypes() {
		if v.Code == code {
			return v
		}
	}

	return AreaType{}
}

type AreaUnit interface {
	Symbol() string
}

type SquareMeter struct {
	Value float32
}

func (sm SquareMeter) Symbol() string {
	return "m2"
}

type Hectare struct {
	Value float32
}

func (h Hectare) Symbol() string {
	return "hectare"
}

type AreaPhoto struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int    `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type AreaNote struct {
	UID         uuid.UUID `json:"uid"`
	Content     string    `json:"content"`
	CreatedDate time.Time `json:"created_date"`
}

// CreateArea registers a new area to a farm
func CreateArea(farm Farm, name string, areaType string) (Area, error) {
	err := validateAreaName(name)
	if err != nil {
		return Area{}, err
	}

	err = validateAreaType(areaType)
	if err != nil {
		return Area{}, err
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return Area{}, err
	}

	return Area{
		UID:  uid,
		Farm: farm,
		Name: name,
		Type: GetAreaType(areaType),
	}, nil
}

// ChangeSize changes an area size
func (a *Area) ChangeSize(size AreaUnit) error {
	err := validateSize(size)
	if err != nil {
		return err
	}

	a.Size = size

	return nil
}

// ChangeLocation changes an area location
func (a *Area) ChangeLocation(location string) error {
	_, err := FindAreaLocationByCode(location)
	if err != nil {
		return err
	}

	a.Location = location

	return nil
}

func (a *Area) AddNewNote(content string) error {
	if content == "" {
		return AreaError{Code: AreaNoteErrorInvalidContent}
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	areaNote := AreaNote{
		UID:         uid,
		Content:     content,
		CreatedDate: time.Now(),
	}

	if len(a.Notes) == 0 {
		a.Notes = make(map[uuid.UUID]AreaNote)
	}

	a.Notes[uid] = areaNote

	return nil
}

func (a *Area) RemoveNote(uid string) error {
	if uid == "" {
		return AreaError{Code: AreaNoteErrorNotFound}
	}

	uuid, err := uuid.FromString(uid)
	if err != nil {
		return AreaError{Code: AreaNoteErrorNotFound}
	}

	found := false
	for _, v := range a.Notes {
		if v.UID == uuid {
			delete(a.Notes, uuid)
			found = true
		}
	}

	if !found {
		return AreaError{Code: AreaNoteErrorNotFound}
	}

	return nil
}

func validateAreaName(name string) error {
	if name == "" {
		return AreaError{AreaErrorNameEmptyCode}
	}
	if !validationhelper.IsAlphanumeric(name) {
		return AreaError{AreaErrorNameAlphanumericOnlyCode}
	}
	if len(name) < 5 {
		return AreaError{AreaErrorNameNotEnoughCharacterCode}
	}
	if len(name) > 100 {
		return AreaError{AreaErrorNameExceedMaximunCharacterCode}
	}

	return nil
}

func validateSize(size AreaUnit) error {
	isValidUnit := true
	sizeValue := float32(0)

	switch v := size.(type) {
	case SquareMeter:
		sizeValue = v.Value
		isValidUnit = true
	case Hectare:
		sizeValue = v.Value
		isValidUnit = true
	default:
		sizeValue = 0
		isValidUnit = false
	}

	if sizeValue <= 0 {
		return AreaError{AreaErrorSizeEmptyCode}
	}
	if isValidUnit == false {
		return AreaError{AreaErrorInvalidSizeUnitCode}
	}

	return nil
}

func validateAreaType(areaType string) error {
	if areaType == "" {
		return AreaError{AreaErrorTypeEmptyCode}
	}

	v := GetAreaType(areaType)
	if v == (AreaType{}) {
		return AreaError{AreaErrorInvalidAreaTypeCode}
	}

	return nil
}
