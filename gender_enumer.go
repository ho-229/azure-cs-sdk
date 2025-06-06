// Code generated by "enumer -type=Gender -linecomment -json"; DO NOT EDIT.

package azure_cs_sdk

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _GenderName = "MaleFemaleNeutral"

var _GenderIndex = [...]uint8{0, 4, 10, 17}

const _GenderLowerName = "malefemaleneutral"

func (i Gender) String() string {
	if i < 0 || i >= Gender(len(_GenderIndex)-1) {
		return fmt.Sprintf("Gender(%d)", i)
	}
	return _GenderName[_GenderIndex[i]:_GenderIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _GenderNoOp() {
	var x [1]struct{}
	_ = x[GenderMale-(0)]
	_ = x[GenderFemale-(1)]
	_ = x[GenderNeutral-(2)]
}

var _GenderValues = []Gender{GenderMale, GenderFemale, GenderNeutral}

var _GenderNameToValueMap = map[string]Gender{
	_GenderName[0:4]:        GenderMale,
	_GenderLowerName[0:4]:   GenderMale,
	_GenderName[4:10]:       GenderFemale,
	_GenderLowerName[4:10]:  GenderFemale,
	_GenderName[10:17]:      GenderNeutral,
	_GenderLowerName[10:17]: GenderNeutral,
}

var _GenderNames = []string{
	_GenderName[0:4],
	_GenderName[4:10],
	_GenderName[10:17],
}

// GenderString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func GenderString(s string) (Gender, error) {
	if val, ok := _GenderNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _GenderNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Gender values", s)
}

// GenderValues returns all values of the enum
func GenderValues() []Gender {
	return _GenderValues
}

// GenderStrings returns a slice of all String values of the enum
func GenderStrings() []string {
	strs := make([]string, len(_GenderNames))
	copy(strs, _GenderNames)
	return strs
}

// IsAGender returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Gender) IsAGender() bool {
	for _, v := range _GenderValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for Gender
func (i Gender) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Gender
func (i *Gender) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Gender should be a string, got %s", data)
	}

	var err error
	*i, err = GenderString(s)
	return err
}
