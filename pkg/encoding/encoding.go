package encoding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func StructToBase64(s interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	err := json.NewEncoder(encoder).Encode(s)
	if err != nil {
		return "", fmt.Errorf("problem encoding user info: %v", err.Error())
	}
	encoder.Close()

	return buf.String(), nil
}
