package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONMap map[string]interface{}

func ParseJSONToStruct(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}

// Value implements the gorm.Valuer interface for storing JSONMap in the database.
func (j JSONMap) Value() (driver.Value, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONMap: %w", err)
	}
	return string(bytes), nil
}

// Scan implements the sql.Scanner interface for retrieving JSONMap from the database.
func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONMap: value is not []byte")
	}
	return json.Unmarshal(bytes, j)
}
